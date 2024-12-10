/*
Package sqlitehistory implements [watermillchat.HistoryRepository]
using a modern SQLite backend.
*/
package sqlitehistory

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/dkotik/watermillchat"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Repository struct {
	db                  *sqlite.Conn
	mostMessagesPerRoom int64
	retention           time.Duration

	stmtInsert  *sqlite.Stmt
	stmtCollect *sqlite.Stmt
	stmtClean   *sqlite.Stmt
}

type RepositoryParameters struct {
	Context    context.Context
	Connection *sqlite.Conn

	// Retention constraints the life time of messages before deletion. Defaults to [watermillchat.DefaultHistoryRetention].
	Retention time.Duration

	// CleanUpFrequency is the pause between message purge
	// cycles based on retention. Defaults to [watermillchat.DefaultHistoryCleanupFrequency].
	CleanUpFrequency time.Duration

	// MostMessagesPerRoom constraints the maximum number of
	// returned messages per room when history is loaded using
	// [Repository.GetRoomMessages]. More messages may still
	// be present in the database, if they retention duration
	// has not yet run out. Defaults to [watermillchat.DefaultHistoryDepth].
	MostMessagesPerRoom int64
}

func NewRepositoryUsingFile(f string, p RepositoryParameters) (*Repository, error) {
	if p.Connection != nil {
		return nil, errors.New("repository connection is already set")
	}
	db, err := sqlite.OpenConn(f, sqlite.OpenReadWrite|sqlite.OpenCreate)
	if err != nil {
		return nil, err
	}
	ctx := cmp.Or(p.Context, context.Background())
	go func() {
		<-ctx.Done()
		if err := db.Close(); err != nil {
			slog.Error("failed to close SQLite file", slog.Any("error", err))
		}
	}()
	p.Connection = db
	return NewRepository(p)
}

func NewRepository(p RepositoryParameters) (r *Repository, err error) {
	if p.Context == nil {
		p.Context = context.Background()
	}
	if p.Connection == nil {
		p.Connection, err = sqlite.OpenConn("file:memory:?mode=memory")
		if err != nil {
			return nil, err
		}
		go func() {
			<-p.Context.Done()
			if err = p.Connection.Close(); err != nil {
				slog.Error("failed to close SQLite file", slog.Any("error", err))
			}
		}()
	}
	if p.Retention < time.Minute {
		if p.Retention != 0 {
			return nil, errors.New("message retention duration cannot be less than one minute")
		}
		p.Retention = watermillchat.DefaultHistoryRetention
	}
	if p.CleanUpFrequency < time.Minute {
		if p.CleanUpFrequency != 0 {
			return nil, errors.New("clean up frequency cannot be less than one minute")
		}
		p.CleanUpFrequency = watermillchat.DefaultCleanupFrequency
	}
	if p.MostMessagesPerRoom < 1 {
		if p.MostMessagesPerRoom != 0 {
			return nil, errors.New("message retention limit cannot be less than one")
		}
		p.MostMessagesPerRoom = watermillchat.DefaultHistoryDepth
	}

	r = &Repository{
		db:                  p.Connection,
		mostMessagesPerRoom: p.MostMessagesPerRoom,
		retention:           p.Retention,
	}

	if err = sqlitex.ExecuteTransient(r.db, `
		CREATE TABLE IF NOT EXISTS wmc_messages (
			id BLOB NOT NULL PRIMARY KEY,
			room_name TEXT NOT NULL,
			author_id TEXT,
			author_name TEXT,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at
		)
	`, nil); err != nil {
		return nil, err
	}
	if err = sqlitex.ExecuteTransient(r.db, `
		CREATE INDEX IF NOT EXISTS wmc_room_name ON wmc_messages(room_name)
	`, nil); err != nil {
		return nil, err
	}
	if err = sqlitex.ExecuteTransient(r.db, `
		CREATE INDEX IF NOT EXISTS wmc_created_at ON wmc_messages(created_at)
	`, nil); err != nil {
		return nil, err
	}

	r.stmtInsert, err = r.db.Prepare(`INSERT INTO wmc_messages (id, room_name, author_id, author_name, content, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`)
	if err != nil {
		return nil, err
	}
	r.stmtCollect, err = r.db.Prepare(`SELECT * FROM wmc_messages WHERE room_name=? ORDER BY created_at DESC LIMIT ?`)
	if err != nil {
		return nil, err
	}
	r.stmtClean, err = r.db.Prepare(`DELETE FROM wmc_messages WHERE created_at<?`)
	if err != nil {
		return nil, err
	}

	go func(ctx context.Context, frequency, retention time.Duration) {
		tick := time.NewTicker(frequency)
		var t time.Time
		var err error
		for {
			select {
			case <-ctx.Done():
				return
			case t = <-tick.C:
				r.stmtClean.BindInt64(1, int64(t.Add(-retention).Unix()))
				_, err = r.stmtClean.Step()
				if err = errors.Join(err, r.stmtClean.Reset()); err != nil {
					slog.Error("failed to clean up messages", slog.Any("error", err))
				}
			}
		}
	}(p.Context, p.CleanUpFrequency, p.Retention)
	return r, nil
}

func (r *Repository) Insert(ctx context.Context, m watermillchat.Broadcast) (err error) {
	r.stmtInsert.BindText(1, m.ID)
	r.stmtInsert.BindText(2, m.RoomName)
	if m.Author != nil {
		r.stmtInsert.BindText(3, m.Author.ID)
		r.stmtInsert.BindText(4, m.Author.Name)
	}
	r.stmtInsert.BindText(5, m.Content)
	r.stmtInsert.BindInt64(6, m.CreatedAt)
	r.stmtInsert.BindInt64(7, m.UpdatedAt)
	_, err = r.stmtInsert.Step()
	return errors.Join(err, r.stmtInsert.Reset())
}

func (r *Repository) GetRoomMessages(ctx context.Context, roomName string) (messages []watermillchat.Message, err error) {
	r.stmtCollect.BindText(1, roomName)
	r.stmtCollect.BindInt64(2, r.mostMessagesPerRoom)
	for {
		if hasRow, err := r.stmtCollect.Step(); err != nil {
			return nil, err
		} else if !hasRow {
			break
		}
		var author *watermillchat.Identity
		if authorID := r.stmtCollect.GetText("author_id"); authorID != "" {
			author = &watermillchat.Identity{
				ID:   authorID,
				Name: r.stmtCollect.GetText("author_name"),
			}
		}
		messages = append(messages, watermillchat.Message{
			ID:        r.stmtCollect.GetText("id"),
			Author:    author,
			Content:   r.stmtCollect.GetText("content"),
			CreatedAt: r.stmtCollect.GetInt64("created_at"),
			UpdatedAt: r.stmtCollect.GetInt64("updated_at"),
		})
	}
	if err = r.stmtCollect.Reset(); err != nil {
		return nil, err
	}
	slices.Reverse(messages)
	return messages, nil
}
