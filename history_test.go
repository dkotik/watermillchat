package watermillchat

import (
	"context"
	"testing"
	"time"
)

func TestRoomMessagesRentention(t *testing.T) {
	const testRoomNameForRetention = "testRoomForRetention"
	const limitMessagesPerRoom = 25
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chat, err := New(ctx, Configuration{
		History: HistoryConfiguration{
			MostMessagesPerRoom: limitMessagesPerRoom,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _ = range 51 {
		if err = chat.Broadcast(ctx, Broadcast{
			RoomName: testRoomNameForRetention,
			Message: Message{
				Content: "test broadcast",
			},
		}); err != nil {
			t.Fatal(err)
		}
	}

	<-time.After(time.Millisecond * 60)

	chat.mu.Lock()
	count := len(chat.rooms[testRoomNameForRetention].messages)
	chat.mu.Unlock()

	if count != 50 {
		t.Fatal("expected 50 messages in memory history, but instead got:", count)
	}

	chat.mu.Lock()
	chat.rooms[testRoomNameForRetention].cleanOut(0, limitMessagesPerRoom)
	chat.mu.Unlock()

	chat.mu.Lock()
	count = len(chat.rooms[testRoomNameForRetention].messages)
	chat.mu.Unlock()

	if count != 25 {
		t.Fatal("expected 25 messages in memory history after clean out, but instead got:", count)
	}
}
