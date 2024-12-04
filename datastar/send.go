package datastar

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dkotik/watermillchat"
)

func NewSendHandler(c *watermillchat.Chat) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		defer func() {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				err = nil
			}
		}()

		name := strings.TrimSpace(r.FormValue("authorName"))
		if name == "" {
			return errors.New("author name must be specified")
		}
		return c.Send(r.Context(), r.FormValue("roomName"), watermillchat.Message{
			AuthorName: name,
			Content:    strings.TrimSpace(r.FormValue("content")),
		})
	}
}
