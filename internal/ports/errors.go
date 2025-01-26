package ports

import (
	"errors"
)

var (
	ErrClientNotInRoom = errors.New("client_not_in_room")
	ErrRoomNotFound    = errors.New("room_not_found")
)
