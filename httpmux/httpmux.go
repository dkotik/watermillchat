/*
Package httpmux injects front-end routing paths into a
[http.ServeMux] for rendering [watermillchat.Chat] view.
*/
package httpmux

import (
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

}
