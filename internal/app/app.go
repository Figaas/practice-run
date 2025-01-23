package app

import (
	"fmt"
	"sync"

	"github.com/segmentio/ksuid"
)

type App struct {
	rooms   map[string]*Room
	clients map[string]*client

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

func (a *App) ConnectNewClient(id, name string, ch MessageChannel) {
	fmt.Println("Connect client: ", id)
	a.clientsMutex.Lock()
	defer a.clientsMutex.Unlock()

	a.clients[id] = newClient(id, name, ch)
}

func (a *App) DisconnectClient(id string) {
	fmt.Println("Disconnect client: ", id)
	a.clientsMutex.Lock()
	defer a.clientsMutex.Unlock()

	// TODO leave rooms

	delete(a.clients, id)
}

func (a *App) CreateNewRoom(name string) {
	fmt.Println("CreateNewRoom")
	_ = a.createNewRoom(name)

	// TODO
	//a.broadcastNewRoom(id, name)
}

func (a *App) createNewRoom(name string) string {
	a.roomsMutex.Lock()
	defer a.roomsMutex.Unlock()

	id := ksuid.New().String()
	a.rooms[id] = newRoom(id, name)

	return id
}

func (a *App) broadcastNewRoom(id string, name string) {
	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	// TODO
}

func (a *App) JoinRoom(clientID string, roomID string) {
	fmt.Println("JoinRoom")

	a.clientsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	client := a.clients[clientID]
	room := a.rooms[roomID]

	client.joinRoom(room)
	room.addClient(client)
}

func (a *App) LeaveRoom(clientID string, roomID string) error {
	fmt.Println("LeaveRoom")

	a.clientsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	client := a.clients[clientID]
	a.rooms[roomID].deleteClient(clientID)

	client.leaveRoom(roomID)

	return nil
}

func (a *App) ListRooms() []Room {
	a.roomsMutex.RLock()
	defer a.roomsMutex.RUnlock()

	ret, i := make([]Room, 0, len(a.rooms)), 0
	for _, v := range a.rooms {
		ret[i] = *v
		i++
	}

	return ret
}

func (a *App) SendMessageToRoom(string, string, string) {
	fmt.Println("SendMessageToRoom")
}
