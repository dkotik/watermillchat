package httpmux

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type options struct {
	Base             *http.ServeMux
	Prefix           string
	RenderingContext context.Context
	Localization     *i18n.Bundle
	PageHead         hypermedia.Head
	ChatOptions      []watermillchat.Option
	Authenticator    Middleware
	Logger           *slog.Logger
}

type Option interface {
	initializeMux(*options) error
}
