/*
Package httpmux injects front-end routing paths into a
[http.ServeMux] for rendering [watermillchat.Chat] view.
*/
package httpmux

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
)

type Error interface {
	error
	GetStatusCode() int
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type Middleware func(http.Handler) http.Handler

type muxError uint8

const (
	ErrUnknown muxError = iota
	ErrUnauthenticatedRequest
)

func (e muxError) Error() string {
	switch e {
	case ErrUnauthenticatedRequest:
		return "unauthenticated request"
	default:
		return "unexpected HTTP multiplexer error"
	}
}

func (e muxError) StatusCode() int {
	switch e {
	case ErrUnauthenticatedRequest:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	var muxError muxError
	if errors.As(err, &muxError) {
		code := muxError.StatusCode()
		w.WriteHeader(code)
		_, _ = io.WriteString(w, http.StatusText(code))
		slog.ErrorContext(
			r.Context(),
			"request failed to complete",
			slog.Any("error", err),
			slog.Int("status_code", code),
			slog.String("path", r.URL.Path),
			slog.String("IP", r.RemoteAddr),
		)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = io.WriteString(w, http.StatusText(http.StatusInternalServerError))
	slog.ErrorContext(
		r.Context(),
		"request failed to complete",
		slog.Any("error", err),
		slog.Int("status_code", http.StatusInternalServerError),
		slog.String("path", r.URL.Path),
		slog.String("IP", r.RemoteAddr),
	)
	return
}
