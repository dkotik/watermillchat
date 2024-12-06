package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/history/sqlitehistory"
)

func main() {
	// Save chat messages to a temporary file:
	sqliteHistory, err := sqlitehistory.NewRepositoryUsingFile(
		filepath.Join(os.TempDir(), "test.sqlite3"), sqlitehistory.RepositoryParameters{})
	if err != nil {
		panic(err)
	}

	chat, err := watermillchat.NewChat(
		watermillchat.WithHistoryRepository(sqliteHistory),
	)
	if err != nil {
		panic(err)
	}

	if err = chat.Send(
		context.Background(),
		"testRoom",
		watermillchat.Message{
			ID:         "test",
			AuthorName: "test user",
			Content:    "test message contents",
		},
	); err != nil {
		panic(err)
	}
}
