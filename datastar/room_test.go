package datastar

import (
	"bytes"
	"testing"
	"watermillchat"
)

func TestRoomTemplateExecution(t *testing.T) {
	b := bytes.Buffer{}
	err := roomTemplate.Execute(&b, roomTemplateParameters{
		RoomName: "Test Room",
		Messages: []watermillchat.Message{},
	})
	if err != nil {
		t.Fatal("unable to render room template:", err.Error())
	}
}
