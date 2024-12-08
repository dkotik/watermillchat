package httpmux

import (
	"context"
	_ "embed" // for template room.html
	"io"
	"text/template"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed room.html
var roomTemplateSource string

var roomTemplate = template.Must(template.New("room").Parse(roomTemplateSource))

type RoomRenderer struct {
	RoomName          string
	MessageSendPath   string
	MessageSourcePath string
}

func (r RoomRenderer) Render(ctx context.Context, w io.Writer, l *i18n.Localizer) error {

	return roomTemplate.Execute(w, r)
}
