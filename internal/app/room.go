package app

import (
	"sync"
)

type Room struct {
	ID           string
	Name         string
	clients      map[string]*client
	clientsMutex sync.RWMutex
}

func newRoom(id, name string) *Room {
	return &Room{
		ID:           id,
		Name:         name,
		clients:      make(map[string]*client, 0),
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
