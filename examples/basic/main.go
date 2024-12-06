package main

import (
	"context"

	"github.com/dkotik/watermillchat"
)

func main() {
	chat, err := watermillchat.NewChat()
	if err != nil {
		panic(err)
	}

	if err = chat.Send(
		context.Background(),
		"testRoom",
		watermillchat.Message{
			Author: &watermillchat.Identity{
				ID:   "test",
				Name: "test user",
			},
			Content: "test message contents",
		},
	); err != nil {
		panic(err)
	}
}
