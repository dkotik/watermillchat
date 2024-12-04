/*
Package watermillchat provides live hypermedia chat
for <watermill.io> event caster. Messages are delivered
as server side events.
*/
package watermillchat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Chat struct {
	publisherTopic string
	publisher      message.Publisher

	rooms map[string]*Room
	mu    sync.Mutex
}

func NewChat(topic string, publisher message.Publisher) (*Chat, error) {
	if topic == "" {
		return nil, errors.New("cannot publish to empty topic")
	}
	if publisher == nil {
		return nil, errors.New("cannot use a <nil> publisher")
	}
	return &Chat{
		publisherTopic: topic,
		publisher:      publisher,
	}, nil
}

func (c *Chat) Publish(b Broadcast) (err error) {
	payload, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("unable to encode broadcast message: %w", err)
	}
	return c.publisher.Publish(c.publisherTopic, message.NewMessage(watermill.NewUUID(), payload))
}

func (c *Chat) Listen(messages <-chan *message.Message) {
	var err error
	var message = Broadcast{}
	var ctx context.Context
	var cancel func()

	for m := range messages {
		message.RoomName = ""
		message.ID = ""
		message.Content = ""
		message.AuthorID = ""
		message.AuthorName = ""

		if err = json.Unmarshal(m.Payload, &message); err != nil {
			slog.Error("dropping malformed broadcast message", slog.Any("error", err), slog.String("ID", m.UUID))
			m.Ack()
			continue
		}
		message.ID = m.UUID
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)

		if err = c.Send(ctx, message.RoomName, message.Message); err != nil {
			if errors.Is(err, context.Canceled) {
				m.Nack()
				continue
			}
			slog.Error("dropping malformed message", slog.Any("error", err), slog.Any("ID", message.ID), slog.Any("roomName", message.RoomName))
		}
		cancel()
		m.Ack()
	}
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
