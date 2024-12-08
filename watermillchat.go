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
	"slices"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Chat struct {
	publisherTopic   string
	publisher        message.Publisher
	history          HistoryRepository
	historyDepth     int
	historyRetention time.Duration

	rooms map[string]*Room
	mu    *sync.Mutex
}

func NewChat(withOptions ...Option) (c *Chat, err error) {
	o := &chatOptions{}
	for _, option := range append(withOptions, DefaultOptions{}) {
		if err = option.initializeChat(o); err != nil {
			return nil, fmt.Errorf("unable to initialize Watermill chat: %w", err)
		}
	}
	incoming, err := o.subscriber.Subscribe(o.context, o.publisherTopic)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Watermill subscriber: %w", err)
	}
	c = &Chat{
		publisherTopic:   o.publisherTopic,
		publisher:        o.publisher,
		history:          o.history,
		historyDepth:     o.historyDepth,
		historyRetention: o.historyRetention,

		rooms: make(map[string]*Room),
		mu:    &sync.Mutex{},
	}
	go c.Listen(incoming)
	go c.cleanup(o.context, o.historyCleanupFrequency)
	return c, nil
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
		message.Author = nil

		if err = json.Unmarshal(m.Payload, &message); err != nil {
			slog.Error("dropping malformed broadcast message", slog.Any("error", err), slog.String("ID", m.UUID))
			m.Ack()
			continue
		}
		message.ID = m.UUID
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)

		if err = c.send(ctx, message.RoomName, message.Message); err != nil {
			if errors.Is(err, context.Canceled) {
				m.Nack()
				cancel()
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
	if roomName == "" {
		return errors.New("chat room name is required")
	}
	if m.Content == "" {
		return errors.New("unable to send an empty message")
	}

	payload, err := json.Marshal(Broadcast{RoomName: roomName, Message: m})
	if err != nil {
		return fmt.Errorf("unable to encode Watermill broadcast to JSON: %w", err)
	}
	if err = c.publisher.Publish(
		c.publisherTopic,
		message.NewMessage(m.ID, payload),
	); err != nil {
		return fmt.Errorf("unable to publish Watermill message: %w", err)
	}
	return nil
}

func (c *Chat) send(ctx context.Context, roomName string, m Message) error {
	c.mu.Lock()
	room, ok := c.rooms[roomName]
	if !ok {
		history, err := c.history.GetRoomMessages(ctx, roomName)
		if err != nil {
			c.mu.Unlock()
			return err
		}
		if grow := c.historyDepth - len(history); grow > 0 {
			history = slices.Grow(history, grow) // increase capacity
		} else if grow < 0 {
			history = history[-grow:] // truncate earlier messages
		}

		room = &Room{
			messages: history,
		}
		c.rooms[roomName] = room
	}
	c.mu.Unlock()

	return room.Send(ctx, m)
}
