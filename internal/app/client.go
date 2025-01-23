package app

import (
	"sync"
)

type MessageChannel interface {
	SendMessage(msg string) error
	Close() error
}

type client struct {
	id             string
	name           string
	rooms          map[string]*Room
	roomsMutex     sync.RWMutex
	messageChannel MessageChannel
}

func newClient(id, name string, messageChannel MessageChannel) *client {
	return &client{
		id:             id,
		name:           name,
		messageChannel: messageChannel,
		rooms:          make(map[string]*Room),
		roomsMutex:     sync.RWMutex{},
	}
}

func (c *client) SendMessage(msg string) error {
	return c.messageChannel.SendMessage(msg)
}

func (c *client) Close() error {
	return c.messageChannel.Close()
}

func (c *client) leaveRoom(id string) {
	c.roomsMutex.Lock()
	defer c.roomsMutex.Unlock()

	delete(c.rooms, id)
}

func (c *client) joinRoom(room *Room) {
	c.roomsMutex.Lock()
	defer c.roomsMutex.Unlock()

	c.rooms[room.ID] = room
}
