package datastar

import (
	"errors"
	"log/slog"
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
		slog.Info("sending message", slog.String("roomName", roomName), slog.String("content", content))

		_ = c.Send(r.Context(), roomName, watermillchat.Message{
			Content: content,
		}) // double everything

		return c.Send(r.Context(), roomName, watermillchat.Message{
			Content: content,
		})
	}
}
