package watermillchat

type Message struct {
	ID         string
	AuthorID   string
	AuthorName string
	Content    string
	CreatedAt  int64
	UpdatedAt  int64
}

type Broadcast struct {
	Message
	RoomName string
}
