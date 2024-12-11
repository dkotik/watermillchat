package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/history/sqlitehistory"
	"github.com/dkotik/watermillchat/httpmux"
	"github.com/dkotik/watermillchat/ollama"
	"github.com/urfave/cli/v3"
)

func serve(ctx context.Context, address string, chat *watermillchat.Chat) error {
	mux, err := httpmux.New(httpmux.Configuration{
		Chat:          chat,
		Authenticator: httpmux.NaiveBearerHeaderAuthenticatorUnsafe,
	})
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	at := listener.Addr().(*net.TCPAddr)

	fmt.Printf("Launching chat server at: http://%s:%d/\n", at.IP, at.Port)

	go func() {
		// TODO: there is no grace timeout - use the version in oakmux
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		if err := listener.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}()

	return http.Serve(listener, mux)
}

func main() {
	err := (&cli.Command{
		Name:  "wmcserver",
		Usage: "run a live text chat server demonstration",
		Action: func(ctx context.Context, c *cli.Command) (err error) {
			configuration := watermillchat.Configuration{}
			historyFile := strings.TrimSpace(c.String("history-file"))
			if historyFile != "" {
				history, err := sqlitehistory.NewRepositoryUsingFile(
					historyFile,
					sqlitehistory.RepositoryParameters{
						Context: ctx,
					},
				)
				if err != nil {
					return fmt.Errorf("unable to set up history file: %w", err)
				}
				configuration.History.Repository = history
			}
			chat, err := watermillchat.New(ctx, configuration)
			if err != nil {
				return err
			}
			bot := ollama.New("", "")
			go bot.JoinChat(ctx, chat, "Ollama", "ollama")
			return serve(ctx, c.String("address"), chat)
		},
		Flags: flags(),
	}).Run(context.Background(), os.Args)

	if err != nil {
		slog.Error("chat server shutdown", slog.Any("reason", err))
	}
}
