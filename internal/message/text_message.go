package message

type TextMessage struct {
	RoomID         string
	FromClientID   string
	FromClientName string
	Message        string
}

func (m TextMessage) IsMessage() {}
