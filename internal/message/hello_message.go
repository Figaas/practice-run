package message

type HelloMessage struct {
	Rooms []Room
}

func (m HelloMessage) IsMessage() {}
