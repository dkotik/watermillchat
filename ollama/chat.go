package ollama

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dkotik/watermillchat"
)

func (o *Ollama) JoinChat(ctx context.Context, c *watermillchat.Chat, botName, roomName string) {
	messages := c.Subscribe(ctx, roomName)
	next := make(chan watermillchat.Message)
	thinking := time.NewTicker(time.Millisecond * 50)
	me := &watermillchat.Identity{
		ID:   "3sadfsdfsdfsdfs",
		Name: botName,
	}

	send := func(parent context.Context, message string) {
		ctx, cancel := context.WithTimeout(parent, time.Second*3)
		defer cancel()

		err := c.Broadcast(ctx, watermillchat.Broadcast{
			RoomName: roomName,
			Message: watermillchat.Message{
				Author:    me,
				Content:   message,
				CreatedAt: time.Now().Unix(),
			},
		})
		if err != nil {
			o.logger.ErrorContext(ctx, "Ollama was unable to speak a message", slog.String("roomName", roomName), slog.Any("identity", me))
		}
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				close(next)
				return
			case <-thinking.C:
				if len(next) > 0 {
					send(ctx, "... I am thinking ...")
				}
			case batch := <-messages:
				for _, message := range batch {
					if message.Author != nil && message.Author.ID == me.ID {
						continue // do not react to own messages
					}

					select {
					case next <- message:
					default:
						o.logger.WarnContext(ctx, "Ollama got a message while busy answering the previous one", slog.String("roomName", roomName))
					}
					break // stop loop after first message
				}
			}
		}
	}(ctx)

	for message := range next {
		answer, err := o.SendMessage(ctx, message.Content)
		if err != nil {
			send(ctx, fmt.Errorf("Ollama API Error: %w", err).Error())
			<-time.After(time.Second * 5)
		} else {
			send(ctx, answer)
		}
	}
}
