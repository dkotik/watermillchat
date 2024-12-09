package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dkotik/watermillchat/httpmux"
	"github.com/urfave/cli/v3"
)

func serve(ctx context.Context, address string) error {
	// history, err := sqlitehistory.NewRepositoryUsingFile(
	// 	filepath.Join(os.TempDir(), "wmcsever-demo.sqlite3"),
	// )
	// if err != nil {
	// 	return fmt.Errorf("unable to set up history file: %w", err)
	// }

	mux, err := httpmux.New(
	// httpmux.WithHistory(history),
	)
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
	(&cli.Command{
		Name:  "wmcserver",
		Usage: "run a live text chat server demonstration",
		Action: func(ctx context.Context, c *cli.Command) error {
			return serve(ctx, c.String("address"))
		},
		Flags: flags(),
	}).Run(context.Background(), os.Args)
}
