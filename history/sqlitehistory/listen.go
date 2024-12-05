package sqlitehistory

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/watermillchat"
)

func (r *Repository) Listen(broadcasts <-chan *message.Message) {
	var err error
	for message := range broadcasts {
		b := watermillchat.Broadcast{}
		if err = json.Unmarshal(message.Payload, &b); err != nil {
			message.Ack()
			slog.Error(
				"malformed broadcast",
				slog.String("ID", message.UUID),
				slog.Any("error", err),
			)
			continue
		}
		if err = r.Insert(context.TODO(), b); err != nil {
			if err = r.Insert(context.TODO(), b); err != nil { // retryn
				slog.Error("failed to store broadcast message into SQLite database", slog.Any("error", err))
			}
		}
	}
}
