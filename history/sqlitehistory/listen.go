package sqlitehistory

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/watermillchat"
)

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

func (r *Repository) Listen(broadcasts <-chan *message.Message) {
	var err error
	for message := range broadcasts {
		b := watermillchat.Broadcast{}
		if err = json.Unmarshal(message.Payload, &b); err != nil {
			message.Ack()
			r.logger.Error(
				"dropped malformed history message broadcast",
				slog.String("message_id", message.UUID),
				slog.Any("error", err),
			)
			continue
		}
		if err = r.Insert(message.Context(), b); err == nil {
			message.Ack()
			continue
		}
		r.logger.Error(
			"failed to store broadcast message into SQLite database",
			slog.String("message_id", message.UUID),
			slog.Any("error", err),
		)
	}
}
