package watermillchat

import (
	"context"
	"slices"
	"sync"
	"time"
)

type Room struct {
	messages []Message
	clients  []chan Message

	mu sync.Mutex
}

func (r *Room) Send(ctx context.Context, m Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if length := len(r.messages); length >= cap(r.messages) && length > 0 {
		r.messages = r.messages[1:]
	}
	r.messages = append(r.messages, m)
	for _, client := range r.clients {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case client <- m:
		}
	}
	return nil
}

func (r *Room) Subscribe(ctx context.Context) <-chan []Message {
	r.mu.Lock()
	history := make([]Message, len(r.messages))
	copy(history, r.messages)
	client := make(chan Message, cap(r.messages)/4+1)
	r.clients = append(r.clients, client)
	r.mu.Unlock()

	batches := make(chan []Message, cap(client)/2+1)
	if len(history) > 0 {
		slices.Reverse(history)
		batches <- history
	}

	go func(ctx context.Context) {
		defer func() {
			r.mu.Lock()
			for i, existing := range r.clients {
				if existing == client {
					r.clients = slices.Delete(r.clients, i, i+1)
					close(client)
				}
			}
			r.mu.Unlock()
			close(batches)
		}()
		tick := time.NewTicker(time.Millisecond * 300)
		limit := cap(client)
		batch := make([]Message, 0, limit)

		for {
			select {
			case <-ctx.Done():
				return
			case item := <-client:
				batch = append(batch, item)
				if len(batch) >= limit {
					batchCopy := make([]Message, len(batch))
					copy(batchCopy, batch)
					slices.Reverse(batchCopy)
					batches <- batchCopy
					batch = batch[:0] // truncate
				}
			case <-tick.C:
				if len(batch) > 0 {
					batchCopy := make([]Message, len(batch))
					copy(batchCopy, batch)
					slices.Reverse(batchCopy)
					batches <- batchCopy
					batch = batch[:0] // truncate
				}
			}
		}
	}(ctx)

	return batches
}
