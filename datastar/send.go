package datastar

import (
	"errors"
	"net/http"
	"strings"
	"watermillchat"
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

		roomName := r.FormValue("roomName")
		if roomName == "" {
			return errors.New("empty room name")
		}
		content := strings.TrimSpace(r.FormValue("content"))
		if content == "" {
			return errors.New("empty message")
		}
		return c.Send(r.Context(), roomName, watermillchat.Message{
			Content: content,
		})
	}
}
