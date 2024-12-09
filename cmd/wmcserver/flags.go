package main

import "github.com/urfave/cli/v3"

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Value:   "localhost:0",
			Usage:   "server address with port to listen on",
		},
		&cli.StringFlag{
			Name:    "history-file",
			Aliases: []string{"f"},
			Value:   "",
			Usage:   "save chat messages to SQLite disk file",
		},
	}
}
