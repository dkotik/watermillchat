package httpmux

import (
	"io"
	"net/http"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/dkotik/watermillchat"
)

func NewSendHandler(
	c *watermillchat.Chat,
	eh ErrorHandler,
) http.HandlerFunc {
	if c == nil {
		panic("cannot use a <nil> Watermill chat")
	}
	if eh == nil {
		panic("cannot use a <nil> error handler")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := watermillchat.IdentityFromContext(r.Context())
		if !ok {
			eh(w, r, ErrUnauthenticatedRequest)
			return
		}
		m := watermillchat.Message{
			ID:      watermill.NewULID(),
			Author:  &identity,
			Content: strings.TrimSpace(r.FormValue("content")),
		}

		err := c.Send(r.Context(), r.FormValue("roomName"), m)
		if err != nil {
			eh(w, r, err)
		}
		if _, err = io.WriteString(w, m.ID); err != nil {
			eh(w, r, err)
		}
	}
}
