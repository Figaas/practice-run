package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"

	"practice-run/internal/message"
)

type messageIDKey struct{}

type connection struct {
	id       string
	conn     *websocket.Conn
	msgChan  chan message.Message
	doneChan chan bool
}

func newWebSocketClient(id string, conn *websocket.Conn) *connection {
	c := &connection{
		id:       id,
		conn:     conn,
		msgChan:  make(chan message.Message, 10),
		doneChan: make(chan bool),
	}

	go c.run()

	return c
}

func (c *connection) SendMessage(raw message.Message) {
	c.msgChan <- raw
}

func (c *connection) RespondOK(ctx context.Context) {
	c.msgChan <- okMessage{
		base: base{Type: "ok_message", MessageID: extractMessageID(ctx)},
	}
}

func (c *connection) RespondError(ctx context.Context, err error) {
	c.msgChan <- errorMessage{
		base:  base{Type: "error_message", MessageID: extractMessageID(ctx)},
		Error: err.Error(),
	}
}

func extractMessageID(ctx context.Context) int64 {
	rawID := ctx.Value(messageIDKey{})
	messageID, ok := rawID.(int64)
	if !ok {
		panic("message ID is not an int64")
	}
	return messageID
}

func (c *connection) Close() {
	c.doneChan <- true
	if err := c.conn.Close(); err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		slog.Error("error closing websocket connection",
			"error_msg", err,
		)
	}
	close(c.msgChan)
	close(c.doneChan)
}

func (c *connection) run() {
	for {
		select {
		case <-c.doneChan:
			return
		case raw := <-c.msgChan:
			c.sendMessage(raw)
		}
	}
}

func (c *connection) sendMessage(raw message.Message) {
	var (
		bs  []byte
		err error
	)

	// IMPROVEMENT: we should send message ID
	switch m := raw.(type) {
	case message.TextMessage:
		out := messageReceived{
			base:     base{Type: "message_received"},
			RoomID:   m.RoomID,
			From:     m.FromClientID,
			FromName: m.FromClientName,
			Msg:      m.Message,
		}
		bs, err = json.Marshal(out)
	case message.NewRoomMessage:
		out := newRoomCreated{
			base: base{Type: "new_room_created"},
			room: room{
				ID:   m.Room.ID,
				Name: m.Room.Name,
			},
		}
		bs, err = json.Marshal(out)
	case message.HelloMessage:
		out := hello{
			base:  base{Type: "hello"},
			Rooms: make([]room, len(m.Rooms)),
		}
		for i, r := range m.Rooms {
			out.Rooms[i] = room{ID: r.ID, Name: r.Name}
		}
		bs, err = json.Marshal(out)
	case errorMessage, okMessage:
		bs, err = json.Marshal(m)
	default:
		err = fmt.Errorf("unsupported message type: %T", m)
	}

	if err != nil {
		slog.Error(" failed to marshal message",
			"error_msg", err,
			"message_type", fmt.Sprintf("%T", raw),
		)
	}

	err = c.conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		// IMPROVEMENT: increase metrics
		// IMPROVEMENT: retry message sending and break in case a connection has been closed in meantime
		slog.Error("failed to send message to client",
			"error_msg", err,
			"client_id", c.id,
			"message_type", fmt.Sprintf("%T", raw),
		)
	}
}
