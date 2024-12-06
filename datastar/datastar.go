/*
Package datastar provides <data-star.dev> view rendering for
[watermillchat.Server].
*/
package datastar

import (
	"bytes"
	_ "embed" // for bundindling datastar
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	datastar "github.com/starfederation/datastar/code/go/sdk"
)

//go:embed error.html
var errorTemplateSource string

var (
	errorTemplate        = template.Must(template.New("error").Parse(errorTemplateSource))
	errorMessageTemplate = errorTemplate.Lookup("message")
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func SendErrorSSE(sse *datastar.ServerSentEventGenerator, err error, ID string) {
	b := &bytes.Buffer{}
	if err = errorMessageTemplate.Execute(b, struct {
		ID      string
		Message string
	}{
		ID:      ID,
		Message: err.Error(),
	}); err != nil {
		panic(err)
	}
	sse.MergeFragments(
		b.String(),
		// datastar.WithSelector(".error"),
	)
}

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			slog.Error("request error", slog.Any("error", err))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			if err = errorTemplate.Execute(w, struct {
				Title   string
				Message string
			}{
				Title:   "Internal Server Error",
				Message: fmt.Sprintf("Error: %s.", err.Error()),
			}); err != nil {
				panic(err)
			}
		}
	}
}
