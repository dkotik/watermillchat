package httpmux

import (
	"errors"
	"net/http"
)

func MountChatWithDatastar(
	mux *http.ServeMux,
	prefix string,
) error {
	if mux == nil {
		return errors.New("cannot use a <nil> mux")
	}
	// TODO: move from datastar package
	return nil
}
