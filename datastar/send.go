package datastar

import (
	"errors"
	"net/http"
	"strings"
	"watermillchat"
)

func NewSendHandler(
	c *watermillchat.Chat,
	rs RoomSelector,
) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		content := strings.TrimSpace(r.FormValue("content"))
		if content == "" {
			return errors.New("empty message")
		}
		roomName, err := rs(r)
		if err != nil {
			return err
		}
		return c.Send(r.Context(), roomName, watermillchat.Message{
			Content: content,
		})
	}
}
