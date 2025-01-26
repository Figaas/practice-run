package app

import (
	"sync"

	"practice-run/internal/message"
	"practice-run/internal/ports"
)

type Room struct {
	ID           string
	Name         string
	clients      clients
	clientsMutex sync.RWMutex
}

func newRoom(id, name string) *Room {
	return &Room{
		ID:           id,
		Name:         name,
		clients:      make(map[string]*client),
		clientsMutex: sync.RWMutex{},
	}
}

func (r *Room) addClient(client *client) {
	r.clientsMutex.Lock()
	defer r.clientsMutex.Unlock()

	r.clients[client.id] = client
}

func (r *Room) deleteClient(id string) {
	r.clientsMutex.Lock()
	defer r.clientsMutex.Unlock()

	delete(r.clients, id)
}

func (r *Room) broadcastClientMessage(clientID string, msg message.TextMessage, onSuccess func(), onError func(err error)) {
	r.clientsMutex.Lock()
	defer r.clientsMutex.Unlock()

	_, ok := r.clients[clientID]
	if !ok {
		onError(ports.ErrClientNotInRoom)
		return
	}

	r.clients.sendMessage(msg)
	onSuccess()
}
