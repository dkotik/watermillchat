package httpmux

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dkotik/watermillchat"
)

type defaultDatastarOptions struct{}

func (d defaultDatastarOptions) initializeDatastarFrontend(o *datastarOptions) error {
	if !strings.HasPrefix(o.Prefix, "/") {
		o.Prefix = "/" + o.Prefix
	}
	if !strings.HasSuffix(o.Prefix, "/") {
		o.Prefix = o.Prefix + "/"
	}
	if o.Mux == nil {
		o.Mux = &http.ServeMux{}
	}
	if o.ErrorHandler == nil {
		o.ErrorHandler = DefaultErrorHandler
	}
	return nil
}

func MountDatastarFrontend(withOptions ...DatastarOption) (mux *http.ServeMux, err error) {
	o := &datastarOptions{}
	for _, option := range append(withOptions, defaultDatastarOptions{}) {
		if err = option.initializeDatastarFrontend(o); err != nil {
			return nil, fmt.Errorf("unable to mount Datastar front-end to HTTP multiplexer: %w", err)
		}
	}

	chat, err := watermillchat.NewChat(o.ChatOptions...)
	if err != nil {
		return nil, err
	}

	// index := ErrorHandler(NewRoomHandler(chat, RoomTemplateParameters{
	// 	Title:             title,
	// 	DataStarPath:      prefix + "datastar.js",
	// 	MessageSourcePath: prefix + "messages.html",
	// 	MessageSendPath:   prefix + "send.html",
	// }, rs))
	// mux.HandleFunc(prefix+"index.html", index)
	// mux.HandleFunc(prefix+"{$}", index)

	// mux.HandleFunc(prefix+"messages.html", ErrorHandler(NewRoomMessagesHandler(chat, NewRoomSelectorFromFormValue("roomName"))))
	o.Mux.HandleFunc(o.Prefix+"send", NewSendHandler(
		chat,
		DefaultErrorHandler,
	))
	// mux.HandleFunc(prefix+"datastar.js", hypermedia.DatastarHandler)
	// mux.HandleFunc(prefix+"datastar.js.map", hypermedia.DatastarMapHandler)
	// TODO: move from datastar package
	return o.Mux, nil
}
