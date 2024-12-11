package watermillchat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

func (c *Chat) distributeToClients(ctx context.Context, roomName string, m Message) error {
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

		if err = c.distributeToClients(ctx, message.RoomName, message.Message); err != nil {
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

		history, err := c.history.GetRoomMessages(
			context.TODO(), roomName)
		if err != nil {
			c.logger.Error("unable to get history messages",
				slog.String("roomName", roomName),
			)
		}
		room = &Room{
			messages: history,
		}
		c.rooms[roomName] = room
	}
	c.mu.Unlock()

	return room.Subscribe(ctx)
}
