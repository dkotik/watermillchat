package datastar

import (
	"net/http"
	"strings"

	"github.com/dkotik/watermillchat"
)

func NewMux(prefix string, rs RoomSelector, title string) *http.ServeMux {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	mux := &http.ServeMux{}
	chat, err := watermillchat.NewChat()
	if err != nil {
		panic(err)
	}
	index := ErrorHandler(NewRoomHandler(chat, RoomTemplateParameters{
		Title:             title,
		DataStarPath:      prefix + "datastar.js",
		MessageSourcePath: prefix + "messages.html",
		MessageSendPath:   prefix + "send.html",
	}, rs))
	mux.HandleFunc(prefix+"index.html", index)
	mux.HandleFunc(prefix+"{$}", index)

	mux.HandleFunc(prefix+"messages.html", ErrorHandler(NewRoomMessagesHandler(chat, NewRoomSelectorFromFormValue("roomName"))))
	mux.HandleFunc(prefix+"send.html", ErrorHandler(NewSendHandler(chat)))
	mux.HandleFunc(prefix+"datastar.js", SourceHandler)
	mux.HandleFunc(prefix+"datastar.js.map", SourceMapHandler)
	return mux
}
