package watermillchat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Message struct {
	ID string

	// Author is the creator of the message. System messages
	// do not contain any author information.
	Author    *Identity
	Content   string
	CreatedAt int64
	UpdatedAt int64
}

type Broadcast struct {
	Message
	RoomName string
}

func (c *Chat) Broadcast(ctx context.Context, b Broadcast) (err error) {
	if b.RoomName == "" {
		return errors.New("chat room name is required")
	}
	if b.Content == "" {
		return errors.New("unable to send an empty message")
	}
	b.ID = watermill.NewUUID()
	payload, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("unable to encode broadcast message: %w", err)
	}
	m := message.NewMessage(b.ID, payload)
	m.SetContext(ctx)
	return c.publisher.Publish(c.publisherTopic, m)
}
