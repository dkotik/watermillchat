package watermillchat

import (
	"context"
)

// History is a list of messages grouped by room name.
// Must be sorted in ascending order by [Message.CreatedAt].
// Apply to [Chat] using [WithHistory] option.
type HistoryRepository interface {
	GetRoomMessages(context.Context, string) ([]Message, error)
}

type VoidHistoryRepository struct{}

func (r VoidHistoryRepository) GetRoomMessages(ctx context.Context, roomName string) ([]Message, error) {
	return nil, nil
}

// type EphemeralHistoryRepository struct {
// 	retention time.Duration
// 	size      int

// 	messages []Broadcast
// 	mu       sync.Mutex
// }

// func (r *EphemeralHistoryRepository) LoadRecentMessages(ctx context.Context, roomName string, cursor string) (history []Message, nextCursor string, err error) {
// 	history = make([]Message, 0, r.size)
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	return slices.DeleteFunc(r.messages, func(b Broadcast) bool {
// 		return b.RoomName != roomName
// 	}), "", nil
// }

// func (r *EphemeralHistoryRepository) SaveMessages(ctx context.Context, roomName string, messages []Message) error {
// 	total := len(messages)
// 	if total > r.size {
// 		messages = messages[total-r.size-1:]
// 		total = r.size
// 	}

// 	broadcasts := make([]Broadcast, total)
// 	for i, m := range messages {
// 		broadcasts[i].Message = m
// 		broadcasts[i].RoomName = roomName
// 	}
// 	r.mu.Lock()
// 	defer r.mu.Unlock()
// 	if tooMany := len(r.messages) + total - r.size; tooMany > 0 {
// 		r.messages = r.messages[tooMany:]
// 	}
// 	r.messages = append(r.messages, broadcasts...)
// 	return nil
// }

// func NewEphemeralHistoryRepository(
// 	retention time.Duration,
// 	size int,
// ) (HistoryRepository, error) {
// 	if retention < time.Second {
// 		return nil, errors.New("retention time cannot be less than one second")
// 	}
// 	if size < 1 {
// 		return nil, errors.New("repository retention size cannot be less than one message")
// 	}
// 	return &EphemeralHistoryRepository{
// 		retention: retention,
// 		size:      size,
// 	}, nil
// }
