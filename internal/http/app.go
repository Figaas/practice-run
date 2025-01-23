package http

import (
	"practice-run/internal/app"
)

type App interface {
	ConnectNewClient(id, name string, ch app.MessageChannel)
	DisconnectClient(id string)
	CreateNewRoom(name string)
	JoinRoom(clientID string, roomID string)
	LeaveRoom(clientID string, roomID string) error
	SendMessageToRoom(clientID string, roomID string, content string)
	ListRooms() []app.Room
}
