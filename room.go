package watermillchat

import (
	"context"
	"sync"
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

func (r *Room) Subscribe() (
	messages <-chan Message,
	history []Message,
	closer func(),
) {
	client := make(chan Message)
	r.mu.Lock()
	history = make([]Message, len(r.messages))
	copy(history, r.messages)
	r.clients = append(r.clients, client)
	r.mu.Unlock()

	return client, history, func() {
		r.mu.Lock()
		for i, existing := range r.clients {
			if existing == messages {
				r.clients = append(r.clients[:i], r.clients[i+1:]...)
				close(client)
			}
		}
		r.mu.Unlock()
	}
}
