package http

import (
	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
}

func newWebSocketClient(conn *websocket.Conn) *client {
	return &client{
		conn: conn,
	}
}

func (c client) SendMessage(msg string) error {
	// TODO handle returned values
	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c client) Close() error {
	// TODO handle returned error handling
	return c.conn.Close()
}
