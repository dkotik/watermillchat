package datastar

import (
	"bytes"
	"testing"
)

func TestRoomTemplateExecution(t *testing.T) {
	b := bytes.Buffer{}
	err := roomTemplate.Execute(&b, RoomTemplateParameters{
		RoomName: "Test Room",
	})
	if err != nil {
		t.Fatal("unable to render room template:", err.Error())
	}
}
