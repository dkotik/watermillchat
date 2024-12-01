package watermillchat

import (
	"context"
	"testing"
)

func TestRoom(t *testing.T) {
	ctx := context.Background()
	r := &Room{}
	messages, history, closer := r.Subscribe()
	go func() {
		for m := range messages {
			t.Log("recieved message:", m)
		}
		t.Log("-- channel closed --")
	}()

	if len(history) > 0 {
		t.Error("history is not empty")
	}
	r.Send(ctx, Message{
		Content: "test message",
	})
	closer()
}
