package httpmux

import (
	"net/http"

	"github.com/dkotik/watermillchat"
)

type datastarOptions struct {
	Prefix        string
	Mux           *http.ServeMux
	Authenticator Middleware
	ErrorHandler  ErrorHandler
	ChatOptions   []watermillchat.Option
}

type DatastarOption interface {
	initializeDatastarFrontend(*datastarOptions) error
}
