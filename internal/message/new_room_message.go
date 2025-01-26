package message

type NewRoomMessage struct {
	Room Room
}

func (m NewRoomMessage) IsMessage() {}

type Room struct {
	ID   string
	Name string
}
