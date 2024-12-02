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

func (c *Chat) Subscribe(ctx context.Context, roomName string) <-chan []Message {
	c.mu.Lock()
	room, ok := c.rooms[roomName]
	if !ok {
		room = &Room{}
		if c.rooms == nil {
			c.rooms = make(map[string]*Room)
		}
		c.rooms[roomName] = room
	}
	c.mu.Unlock()

	return room.Subscribe(ctx)
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
