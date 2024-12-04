package watermillchat

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRoom(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	r := &Room{
		messages: make([]Message, 0, 10),
	}
	messages := r.Subscribe(ctx)
	go func() {
		for i := range 20 {
			r.Send(ctx, Message{
				Content: fmt.Sprintf("test message: %d", i),
			})
		}
	}()

	for batch := range messages {
		for _, m := range batch {
			t.Log("received message:", m)
		}
		if len(batch) == 0 {
			t.Error("got an empty message batch")
		}
	}
	t.Log("-- channel closed --")
}
