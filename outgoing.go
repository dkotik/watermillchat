package watermillchat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

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

func (c *Chat) Publish(b Broadcast) (err error) {
	payload, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("unable to encode broadcast message: %w", err)
	}
	return c.publisher.Publish(c.publisherTopic, message.NewMessage(watermill.NewUUID(), payload))
}
