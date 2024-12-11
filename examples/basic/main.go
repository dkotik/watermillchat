package main

import (
	"context"

	"github.com/dkotik/watermillchat"
)

func main() {
	chat, err := watermillchat.New(context.Background(), watermillchat.Configuration{})
	if err != nil {
		panic(err)
	}

	if err = chat.Broadcast(
		context.Background(),
		watermillchat.Broadcast{
			RoomName: "testRoom",
			Message: watermillchat.Message{
				Author: &watermillchat.Identity{
					ID:   "test",
					Name: "test user",
				},
				Content: "test message contents",
			},
		},
	); err != nil {
		panic(err)
	}
}
