package app

import (
	"context"
	"sync"

	"practice-run/internal/message"
	"practice-run/internal/ports"
)

type client struct {
	id             string
	name           string
	rooms          map[string]*Room
	roomsMutex     sync.RWMutex
	messageChannel ports.MessageChannel
}

func newClient(id, name string, messageChannel ports.MessageChannel) *client {
	return &client{
		id:             id,
		name:           name,
		messageChannel: messageChannel,
		rooms:          make(map[string]*Room),
		roomsMutex:     sync.RWMutex{},
	}
}

func (c *client) sendMessage(message message.Message) {
	c.messageChannel.SendMessage(message)
}

func (c *client) respondOK(ctx context.Context) func() {
	return func() {
		c.messageChannel.RespondOK(ctx)
	}
}

func (c *client) respondError(ctx context.Context) func(err error) {
	return func(err error) {
		c.messageChannel.RespondError(ctx, err)
	}
}

func (c *client) leaveRoom(roomID string, onSuccess func(), onError func(err error)) {
	c.roomsMutex.Lock()
	defer c.roomsMutex.Unlock()

	_, ok := c.rooms[roomID]
	if !ok {
		onError(ports.ErrClientNotInRoom)
		return
	}

	delete(c.rooms, roomID)
	onSuccess()
}

func (c *client) joinRoom(room *Room, onSuccess func()) {
	c.roomsMutex.Lock()
	defer c.roomsMutex.Unlock()

	c.rooms[room.ID] = room
	onSuccess()
}

func (c *client) leaveRooms() {
	c.roomsMutex.Lock()
	defer c.roomsMutex.Unlock()

	for _, room := range c.rooms {
		room.deleteClient(c.id)
	}
}

func (c *client) close() {
	c.messageChannel.Close()
}

type clients map[string]*client

func (cs clients) sendMessage(msg message.Message) {
	for _, c := range cs {
		// IMPROVEMENT: Implement/use existing worker pool to limit spawning enormous number of goroutines
		go func(c *client) {
			c.sendMessage(msg)
		}(c)
	}
}
