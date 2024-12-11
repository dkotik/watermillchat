package httpmux

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/dkotik/watermillchat"
	"github.com/dkotik/watermillchat/httpmux/hypermedia"
	datastar "github.com/starfederation/datastar/code/go/sdk"
)

var messageTemplate = template.Must(template.New("message").Parse(
	`<div class="message" data-scroll-into-view.smooth.vend>
  <p class="author{{if .System}} system{{end}}">{{ with .Author }}{{or .Name "???"}}{{else}}???{{end}}</p>
  <p class="content">{{- .Content -}}</p>
</div>`))

func NewRoomMessagesHandler(
	c *watermillchat.Chat,
	selector RoomSelector,
	eh hypermedia.ErrorHandler,
) http.HandlerFunc {
	if c == nil {
		panic("cannot use a <nil> Watermill chat")
	}
	if selector == nil {
		panic("cannot use a <nil> selector")
	}
	if eh == nil {
		panic("cannot use a <nil> error handler")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		roomName, err := selector(r)
		if err != nil {
			eh.HandlerError(w, r, err)
			return
		}
		sse := datastar.NewSSE(w, r)
		if err = sse.RemoveFragments("section.messages > .message"); err != nil {
			eh.HandlerError(w, r, err)
			return
		}
		b := &bytes.Buffer{}

		for batch := range c.Subscribe(r.Context(), roomName) {
			for _, message := range batch {
				if err = messageTemplate.Execute(b, struct {
					Author  *watermillchat.Identity
					Content string
					System  bool
				}{
					Author:  message.Author,
					Content: message.Content,
					System:  message.Author == nil,
				}); err != nil {
					panic(fmt.Errorf("message template execution failed: %w", err))
				}
			}

			if err = sse.MergeFragments(
				b.String(),
				datastar.WithSelector(".messages"),
				datastar.WithMergeAppend(),
			); err != nil {
				slog.DebugContext(r.Context(), "failed to deliver server sent event to the client", slog.Any("error", err))
			}
			b.Reset()
		}
	}
}

func NewMessageSendHandler(
	c *watermillchat.Chat,
	eh hypermedia.ErrorHandler,
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
			eh.HandlerError(w, r, hypermedia.ErrForbidden)
			return
		}
		roomName := r.FormValue("roomName")
		if roomName == "" {
			eh.HandlerError(w, r, hypermedia.ErrNotFound)
			return
		}

		m := watermillchat.Message{
			ID:        watermill.NewULID(),
			Author:    &identity,
			Content:   strings.TrimSpace(r.FormValue("content")),
			CreatedAt: time.Now().Unix(),
		}

		// err := c.Send(r.Context(), r.FormValue("roomName"), m)
		err := c.Broadcast(r.Context(), watermillchat.Broadcast{
			RoomName: roomName,
			Message:  m,
		})
		if err != nil {
			eh.HandlerError(w, r, err)
		}
		if _, err = io.WriteString(w, m.ID); err != nil {
			eh.HandlerError(w, r, err)
		}
	}
}
