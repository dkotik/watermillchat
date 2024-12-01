package datastar

import (
	_ "embed" // for templates
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"watermillchat"
)

//go:embed room.html
var roomHTML string

var (
	roomTemplate    = template.Must(template.New("room").Parse(roomHTML))
	messageTemplate = roomTemplate.Lookup("message")
)

type roomTemplateParameters struct {
	RoomName string
	SendPath string
	Messages []watermillchat.Message
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

func NewRoomHandler(
	c *watermillchat.Chat,
	rs RoomSelector,
) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		roomName, err := rs(r)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err = roomTemplate.Execute(w, roomTemplateParameters{
			RoomName: roomName,
			// Messages: history,
		}); err != nil {
			return err
		}

		// messages, history, closer := c.Subscribe(roomName)
		// defer closer()
		// b := &bytes.Buffer{}
		// sse := datastar.NewSSE(w, r)
		// for message := range messages {
		// 	if err = messageTemplate.Execute(b, message); err != nil {
		// 		return err // TODO: render error event instead?
		// 	}
		// 	if err = sse.MergeFragments(
		// 		b.String(),
		// 		datastar.WithSelector("title"),
		// 	); err != nil {
		// 		return err // TODO: render error event instead?
		// 	}
		// 	b.Reset()
		// }
		return nil
	}
}
