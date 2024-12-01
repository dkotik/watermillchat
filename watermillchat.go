/*
Package watermillchat provides live hypermedia chat
for <watermill.io> event caster. Messages are delivered
as server side events.
*/
package watermillchat

import (
	"context"
	"sync"
)

type Chat struct {
	rooms map[string]*Room

	mu sync.Mutex
}

func (c *Chat) Subscribe(roomName string) (
	messages <-chan Message,
	history []Message,
	closer func(),
) {
	c.mu.Lock()
	room, ok := c.rooms[roomName]
	if !ok {
		room = &Room{}
		c.rooms[roomName] = room
	}
	c.mu.Unlock()

	return room.Subscribe()
}

func (c *Chat) Send(ctx context.Context, roomName string, m Message) error {
	c.mu.Lock()
	room, ok := c.rooms[roomName]
	if !ok {
		room = &Room{}
		c.rooms[roomName] = room
	}
	c.mu.Unlock()

	return room.Send(ctx, m)
}

// TODO: clean up rooms with no activity for a while
