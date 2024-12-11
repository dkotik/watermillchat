package sqlitehistory_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/history/sqlitehistory"
)

func TestFileBacked(t *testing.T) {
	target := filepath.Join(t.TempDir(), "test.sqlite3")
	writeBroadcast := func(b watermillchat.Broadcast) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		history, err := sqlitehistory.NewUsingFile(
			target, sqlitehistory.RepositoryParameters{
				Context: ctx,
			})
		if err != nil {
			t.Fatal(err)
		}
		err = history.Insert(ctx, b)
		if err != nil {
			t.Fatal(err)
		}
	}

	// open and close twice to make sure `IF NOT EXIST` did not clash
	writeBroadcast(watermillchat.Broadcast{
		Message: watermillchat.Message{
			ID:      "test1",
			Content: "test1",
		}, RoomName: "test"})
	writeBroadcast(watermillchat.Broadcast{
		Message: watermillchat.Message{
			ID:      "test2",
			Content: "test2",
		}, RoomName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	history, err := sqlitehistory.NewUsingFile(
		target, sqlitehistory.RepositoryParameters{
			Context: ctx,
		})
	if err != nil {
		t.Fatal(err)
	}
	messages, err := history.GetRoomMessages(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatal("unexpected number of messages")
	}
	if messages[0].ID != "test2" {
		t.Fatal("returned message ID does not match the original")
	}
	if messages[1].ID != "test1" {
		t.Fatal("returned message ID does not match the original")
	}
}

func TestSQLiteConnection(t *testing.T) {
	history, err := sqlitehistory.New(sqlitehistory.RepositoryParameters{
		MostMessagesPerRoom: 100,
		Retention:           time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if history == nil {
		t.Fatal("<nil> history returned by the constructor")
	}
	testInsert(t, history)
}

func testInsert(t *testing.T, history *sqlitehistory.Repository) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := history.Insert(ctx, watermillchat.Broadcast{Message: watermillchat.Message{
		ID: "test",
	}, RoomName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	messages, err := history.GetRoomMessages(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatal("unexpected number of messages")
	}
	if messages[0].ID != "test" {
		t.Fatal("returned message ID does not match the original")
	}
}
