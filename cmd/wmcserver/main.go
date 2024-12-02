package main

import (
	"net/http"
	"watermillchat/datastar"
)

var allowedRooms = []string{"test"}

func main() {
	http.Handle("/", datastar.NewMux("/", func(r *http.Request) (string, error) {
		return "test", nil
	}))
	http.ListenAndServe("localhost:8081", nil)
}
