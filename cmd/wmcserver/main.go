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

	"github.com/dkotik/watermillchat/datastar"
	"github.com/urfave/cli/v3"
)

var allowedRooms = []string{"test"}

func serve(ctx context.Context, address string) error {
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

	return http.Serve(listener, datastar.NewMux("/", func(r *http.Request) (string, error) {
		return "test", nil
	}, "Watermill Chat Demonstration"))
}

func main() {
	(&cli.Command{
		Name:  "wmcserver",
		Usage: "run a live text chat server demonstration",
		Action: func(ctx context.Context, c *cli.Command) error {
			return serve(ctx, c.String("address"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "address",
				Value: "localhost:0",
				Usage: "server address with port to listen on",
			},
		},
	}).Run(context.Background(), os.Args)
}
