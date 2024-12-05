/*
Package hypermedia provides templates and [http.Handler]s for rendering chat sessions.
*/
package hypermedia

import (
	_ "embed" // for including assets
	"errors"
	"log/slog"
	"text/template"

	"bytes"
	"io"
	"net/http"
)

type (
	Handler      func(main io.Writer, r *http.Request) error
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
	Error        interface {
		error
		HyperTextStatusCode() int
	}
)

var (
	//go:embed icon/favicon.png
	faviconPNG []byte

	errorTemplate = template.Must(template.New("error").Parse(`<h1>{{ .Title }} (#{{ .StatusCode }})</h1><p>{{ .Error }}</p>`))
)

func NewFavIcon() http.Handler {
	return NewAsset("image/png", faviconPNG)
}

func NewAsset(contentType string, body []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		_, _ = io.Copy(w, bytes.NewReader(body))
	}
}

func NewErrorHandler(head, tail []byte, logger *slog.Logger) ErrorHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var htError Error
		var statusCode = http.StatusInternalServerError
		if errors.As(err, &htError) {
			statusCode = htError.HyperTextStatusCode()
		}
		w.WriteHeader(statusCode)
		_, _ = io.Copy(w, bytes.NewReader(head))
		_ = errorTemplate.Execute(w, struct {
			StatusCode int
			Title      string
			Error      string
		}{
			StatusCode: statusCode,
			Title:      http.StatusText(statusCode),
			Error:      err.Error(),
		})
		_, _ = io.Copy(w, bytes.NewReader(tail))
		logger.Error(
			"page request failed",
			slog.Any("error", err),
			slog.Int("statusCode", statusCode),
		)
	}
}
