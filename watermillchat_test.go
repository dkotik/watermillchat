package watermillchat_test

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dkotik/watermillchat"
)

type mockHistoryRepository struct {
	totalMessagesRecieved int
}

func (r *mockHistoryRepository) Listen(broadcasts <-chan *message.Message) {
	for m := range broadcasts {
		r.totalMessagesRecieved += 1
		m.Ack()
	}
}

func (r *mockHistoryRepository) GetRoomMessages(ctx context.Context, roomName string) ([]watermillchat.Message, error) {
	return nil, nil
}

func TestHistoryDeliver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	history := &mockHistoryRepository{}
	chat, err := watermillchat.New(ctx, watermillchat.Configuration{
		History: watermillchat.HistoryConfiguration{
			Repository: history,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _ = range 50 {
		if err = chat.Broadcast(ctx, watermillchat.Broadcast{
			RoomName: "testRoom",
			Message: watermillchat.Message{
				Content: "test message",
			},
		}); err != nil {
			t.Fatal("unable to broadcast a message:", err)
		}
	}

	<-time.After(time.Second)
	if history.totalMessagesRecieved != 50 {
		t.Fatal("unexpected messages in history:", history.totalMessagesRecieved)
	}
}
