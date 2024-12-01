package main

import (
	"net/http"
	"watermillchat"
	"watermillchat/datastar"
)

var allowedRooms = []string{"test"}

func main() {
	chat := &watermillchat.Chat{}
	index := datastar.ErrorHandler(datastar.NewRoomHandler(chat,
		func(r *http.Request) (string, error) {
			return allowedRooms[0], nil
		}))

	http.HandleFunc("/api/v1/chat/send", datastar.ErrorHandler(
		datastar.NewSendHandler(chat,
			datastar.NewRoomSelectorClamp(
				datastar.NewRoomSelectorFromFormValue("roomName"), allowedRooms...))))
	http.HandleFunc("/datastar.js", datastar.SourceHandler)
	http.HandleFunc("/datastar.js.map", datastar.SourceMapHandler)
	http.HandleFunc("/events.html", EventHandler)
	http.HandleFunc("/index.html", index)
	http.HandleFunc("/room/{roomName}", datastar.ErrorHandler(datastar.NewRoomHandler(chat,
		datastar.NewRoomSelectorClamp(
			datastar.NewRoomSelectorFromURL("roomName"), allowedRooms...))))
	http.HandleFunc("/", index)
	http.ListenAndServe("localhost:8081", nil)
}
