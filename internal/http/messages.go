package http

type base struct {
	Type string `json:"type"`
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

type roomsUpdate struct {
	base
	Rooms []room `json:"rooms"`
}
