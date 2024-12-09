package httpmux

import (
	"context"
	_ "embed" // for template room.html
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed room.html
var roomTemplateSource string

var roomTemplate = template.Must(template.New("room").Parse(roomTemplateSource))

type RoomRenderer struct {
	RoomName        string
	MessageSendPath string
	MessageSource   string
}

func (r RoomRenderer) Render(ctx context.Context, w io.Writer, l *i18n.Localizer) error {

	return roomTemplate.Execute(w, r)
}

type RoomSelector func(*http.Request) (string, error)

func NewRoomSelectorClamp(rs RoomSelector, allowed ...string) RoomSelector {
	return func(r *http.Request) (roomName string, err error) {
		roomName, err = rs(r)
		for _, each := range allowed {
			if each == roomName {
				return roomName, nil
			}
		}
		return "", fmt.Errorf("room name is not in the allowed list: %s", roomName)
	}
}

func NewRoomSelectorFromURL(routePathSegmentName string) RoomSelector {
	return func(r *http.Request) (string, error) {
		roomName := strings.TrimSpace(r.PathValue(routePathSegmentName))
		if roomName == "" {
			return "", fmt.Errorf("route does not contain path segment: %s", routePathSegmentName)
		}
		return roomName, nil
	}
}

func NewRoomSelectorFromFormValue(formValueName string) RoomSelector {
	return func(r *http.Request) (string, error) {
		roomName := strings.TrimSpace(r.FormValue(formValueName))
		if roomName == "" {
			return "", fmt.Errorf("form value absent from request: %s", formValueName)
		}
		return roomName, nil
	}
}

func NewRoomSelectorFromURLQueryValue(name string) RoomSelector {
	return func(r *http.Request) (string, error) {
		roomName := strings.TrimSpace(r.URL.Query().Get(name))
		if roomName == "" {
			return "", fmt.Errorf("query value absent from request URL: %s", name)
		}
		return roomName, nil
	}
}
