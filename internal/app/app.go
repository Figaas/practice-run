package app

import (
	"context"
	"log/slog"
	"sync"

	"github.com/segmentio/ksuid"

	"practice-run/internal/message"
	"practice-run/internal/ports"
)

var _ ports.App = (*App)(nil)

type App struct {
	rooms   map[string]*Room
	clients clients

	clientsMutex sync.RWMutex
	roomsMutex   sync.RWMutex
}

func NewApp() *App {
	return &App{
		rooms:        map[string]*Room{},
		clients:      map[string]*client{},
		clientsMutex: sync.RWMutex{},
	}
}

// IMPROVEMENT: clientID and roomID should be custom types

func (a *App) ConnectNewClient(_ context.Context, id, name string, ch ports.MessageChannel) {
	slog.Debug("Connect client", "client_id", id)
	a.clientsMutex.Lock()
	c := newClient(id, name, ch)
	a.clients[id] = c
	a.clientsMutex.Unlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	i := 0
	msg := message.HelloMessage{
		Rooms: make([]message.Room, len(a.rooms)),
	}
	for _, v := range a.rooms {
		msg.Rooms[i] = message.Room{Name: v.Name, ID: v.ID}
		i++
	}

	c.sendMessage(msg)
}

func (a *App) DisconnectClient(_ context.Context, id string) {
	slog.Debug("Disconnect client: ", id)

	a.clientsMutex.Lock()
	defer a.clientsMutex.Unlock()

	c := a.clients[id]
	c.leaveRooms()
	delete(a.clients, id)
	c.close()
}

func (a *App) CreateNewRoom(ctx context.Context, name string, clientID string) {
	slog.Debug("CreateNewRoom")
	id := a.createNewRoom(name)

	a.broadcastNewRoom(ctx, id, name, clientID)
}

func (a *App) createNewRoom(name string) string {
	a.roomsMutex.Lock()
	defer a.roomsMutex.Unlock()

	id := ksuid.New().String()
	a.rooms[id] = newRoom(id, name)

	return id
}

func (a *App) broadcastNewRoom(ctx context.Context, id string, name string, clientID string) {
	a.clientsMutex.RLock()
	defer a.clientsMutex.RUnlock()

	msg := message.NewRoomMessage{
		Room: message.Room{
			ID:   id,
			Name: name,
		},
	}

	a.clients[clientID].respondOK(ctx)
	a.clients.sendMessage(msg)
}

func (a *App) JoinRoom(ctx context.Context, clientID string, roomID string) {
	slog.Debug("JoinRoom")

	a.clientsMutex.RLock()
	defer a.clientsMutex.RUnlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	c := a.clients[clientID]
	room, ok := a.rooms[roomID]
	if !ok {
		c.respondError(ctx)(ports.ErrRoomNotFound)
		return
	}

	room.addClient(c)
	c.joinRoom(
		room,
		c.respondOK(ctx),
	)
}

func (a *App) LeaveRoom(ctx context.Context, clientID string, roomID string) {
	slog.Debug("LeaveRoom")

	a.clientsMutex.RLock()
	defer a.clientsMutex.RUnlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	c := a.clients[clientID]
	r, ok := a.rooms[roomID]
	if !ok {
		c.respondError(ctx)(ports.ErrRoomNotFound)
		return
	}

	r.deleteClient(clientID)
	c.leaveRoom(
		roomID,
		c.respondOK(ctx),
		c.respondError(ctx),
	)
}

func (a *App) SendMessageToRoom(ctx context.Context, clientID string, roomID string, content string) {
	slog.Debug("SendMessageToRoom")

	a.clientsMutex.RLock()
	defer a.clientsMutex.RUnlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	c := a.clients[clientID]
	r, ok := a.rooms[roomID]
	if !ok {
		c.respondError(ctx)(ports.ErrRoomNotFound)
	}

	msg := message.TextMessage{
		RoomID:         roomID,
		FromClientID:   clientID,
		FromClientName: c.name,
		Message:        content,
	}

	r.broadcastClientMessage(
		clientID,
		msg,
		c.respondOK(ctx),
		c.respondError(ctx),
	)
}
