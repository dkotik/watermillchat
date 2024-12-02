package datastar

import (
	"net/http"
	"strings"
	"watermillchat"
)

func NewMux(prefix string, rs RoomSelector) *http.ServeMux {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	mux := &http.ServeMux{}
	chat := &watermillchat.Chat{}
	index := ErrorHandler(NewRoomHandler(chat, RoomTemplateParameters{
		DataStarPath:      prefix + "datastar.js",
		MessageSourcePath: prefix + "messages.html",
		MessageSendPath:   prefix + "send.html",
	}, rs))
	mux.HandleFunc(prefix+"index.html", index)
	mux.HandleFunc(prefix+"{$}", index)

	mux.HandleFunc(prefix+"messages.html", ErrorHandler(NewRoomMessagesHandler(chat, NewRoomSelectorFromFormValue("roomName"))))
	mux.HandleFunc(prefix+"send.html", ErrorHandler(
		NewSendHandler(chat, NewRoomSelectorFromFormValue("roomName"))))
	mux.HandleFunc(prefix+"datastar.js", SourceHandler)
	mux.HandleFunc(prefix+"datastar.js.map", SourceMapHandler)
	return mux
}