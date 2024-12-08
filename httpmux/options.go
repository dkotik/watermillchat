package httpmux

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Middleware func(http.Handler) http.Handler

type options struct {
	Context       context.Context
	Localization  *i18n.Bundle
	Mux           *http.ServeMux
	Prefix        string
	Head          hypermedia.Head
	ChatOptions   []watermillchat.Option
	Authenticator Middleware
	Logger        *slog.Logger
}

type DatastarOption interface {
	initializeDatastarFrontend(*options) error
}
