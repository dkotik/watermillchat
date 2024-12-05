package sqlitehistory_test

import (
	"context"
	"testing"
	"time"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/history/sqlitehistory"
)

func TestSQLiteConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	history, err := sqlitehistory.NewRepository(sqlitehistory.RepositoryParameters{
		MostMessagesPerRoom: 100,
		Retention:           time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if history == nil {
		t.Fatal("<nil> history returned by the constructor")
	}

	if err = history.Insert(ctx, watermillchat.Broadcast{Message: watermillchat.Message{
		ID: "test",
	}, RoomName: "test"}); err != nil {
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
