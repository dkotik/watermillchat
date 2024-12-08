package hypermedia

import (
	"log/slog"
	"net/http"
)

func ErrorHandlerWithLogger(
	after ErrorHandler,
	logger *slog.Logger,
) ErrorHandler {
	if after == nil {
		panic("cannot add logging to a <nil> error handler")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return ErrorHandlerFunc(
		func(w http.ResponseWriter, r *http.Request, err error) {
			after.HandlerError(w, r, err)
			logger.ErrorContext(
				r.Context(),
				"request failed to complete",
				slog.Any("error", err),
				// TODO: add *slog.Value as a group to error object
				// slog.Int("status_code", statusCode),
				slog.String("host", r.URL.Hostname()),
				slog.String("path", r.URL.Path),
				slog.String("client_address", r.RemoteAddr),
			)
		},
	)
}
