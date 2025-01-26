package http

// IMPROVEMENT: documentation for external usage for ex. with asyncAPI

type base struct {
	// IMPROVEMENT: define messageType type and variables and use them instead raw strings
	Type      string `json:"type"`
	MessageID int64  `json:"message_id"`
}

type joinRoom struct {
	base
	RoomID string `json:"room_id"`
}
type leaveRoom struct {
	base
	RoomID string `json:"room_id"`
}
type sendMessage struct {
	base
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

type messageReceived struct {
	base
	RoomID   string `json:"room_id"`
	From     string `json:"from"`
	FromName string `json:"from_name"`
	Msg      string `json:"msg"`
}

type createNewRoom struct {
	base
	Name string `json:"name"`
}

type room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type newRoomCreated struct {
	base
	room
}

type hello struct {
	base
	Rooms []room `json:"rooms"`
}

type errorMessage struct {
	base
	Error string `json:"error"`
}

func (m errorMessage) IsMessage() {}

type okMessage struct {
	base
}

func (m okMessage) IsMessage() {}
