package http

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"
)

func (s *Server) handleNewClientConnection(conn *websocket.Conn, id, name string) {
	s.app.ConnectNewClient(context.Background(), id, name, newWebSocketClient(id, conn))
}

func (s *Server) handleDisconnectedClient(conn *websocket.Conn, clientID string, stop func()) <-chan struct{} {
	ch := make(chan struct{})
	conn.SetCloseHandler(func(code int, text string) error {
		s.app.DisconnectClient(context.Background(), clientID)

		stop()
		return nil
	})

	return ch
}

func (s *Server) handleJoinRoom(ctx context.Context, raw []byte, clientID string) {
	msg := &joinRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.JoinRoom(ctx, clientID, msg.RoomID)
}

func (s *Server) handleLeaveRoom(ctx context.Context, raw []byte, clientID string) {
	msg := &leaveRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.LeaveRoom(ctx, clientID, msg.RoomID)
}

func (s *Server) handleCreateNewRoom(ctx context.Context, raw []byte) {
	msg := &createNewRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.CreateNewRoom(ctx, msg.Name, "")
}

func (s *Server) handleSendMessage(ctx context.Context, raw []byte, clientID string) {
	msg := &sendMessage{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.SendMessageToRoom(ctx, clientID, msg.RoomID, msg.Content)
}
