package ports

import (
	"context"
)

type App interface {
	ConnectNewClient(ctx context.Context, id, name string, ch MessageChannel)
	DisconnectClient(ctx context.Context, id string)
	CreateNewRoom(ctx context.Context, name string, clientID string)
	JoinRoom(ctx context.Context, clientID string, roomID string)
	LeaveRoom(ctx context.Context, clientID string, roomID string)
	SendMessageToRoom(ctx context.Context, clientID string, roomID string, content string)
}
