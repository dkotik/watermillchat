package watermillchat

type Message struct {
	ID string

	// Author is the creator of the message. System messages
	// do not contain any author information.
	Author    *Identity
	Content   string
	CreatedAt int64
	UpdatedAt int64
}

type Broadcast struct {
	Message
	RoomName string
}
