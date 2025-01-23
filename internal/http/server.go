package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	app App
}

func NewServer(app App) *Server {
	return &Server{app: app}
}

func (s *Server) Run() {
	// TODO use custom server
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(w, r)
	})

	addr := ":8080"
	slog.Info("starting server at " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to upgrade connection to websocket", "error_msg", err)
		return
	}

	// IMPROVEMENTS: client ID and name should be returned during authentication and extracted for request metadata
	name := r.Header.Get("name")
	clientID := r.Header.Get("client_id")
	s.handleNewClientConnection(conn, clientID, name)
	disconnectChan := s.handleDisconnectClient(conn, clientID)

	for {
		select {
		case <-disconnectChan:
			return
		default:
			s.handleMessages(conn, clientID)
		}
	}
}

func (s *Server) handleNewClientConnection(conn *websocket.Conn, id, name string) {
	s.app.ConnectNewClient(id, name, newWebSocketClient(conn))
	rooms := s.app.ListRooms()

	msg := roomsUpdate{
		base:  base{Type: "rooms_update"},
		Rooms: make([]room, len(rooms)),
	}

	for i, r := range rooms {
		msg.Rooms[i] = room{Name: r.Name, ID: r.ID}
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal hello message", "error_msg", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil && !errors.Is(err, websocket.ErrCloseSent) {
		slog.Error("failed to send hello message", "error_msg", err)
	}
}

func (s *Server) handleMessages(conn *websocket.Conn, clientID string) {
	messageType, raw, err := conn.ReadMessage()
	if err != nil {
		slog.Error("failed to read message", "error_msg", err)
		return
	}

	if messageType != websocket.TextMessage {
		// IMPROVEMENT: we should validate messages and gather metrics instead logging them
		// For now, it's better to log anything.
		slog.Error("invalid message type", "type", messageType)
	}

	b := &base{}
	err = json.Unmarshal(raw, b)
	if err != nil {
		slog.Error("failed to unmarshal base message", "error_msg", err)

	}

	switch b.Type {
	case "join_room":
		s.handleJoinRoom(raw, clientID)
	case "leave_room":
		s.handleLeaveRoom(raw, clientID)
	case "create_new_room":
		s.handleCreateNewRoom(raw)
	case "send_message":
		s.handleSendMessage(raw, clientID)
	}

	resMessage := fmt.Sprintf("created room with name %s", string(raw))
	if err := conn.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
		log.Println(err)
		return
	}

	return
}

func (s *Server) handleDisconnectClient(conn *websocket.Conn, clientID string) <-chan struct{} {
	ch := make(chan struct{})
	conn.SetCloseHandler(func(code int, text string) error {
		s.app.DisconnectClient(clientID)

		ch <- struct{}{}
		return nil
	})

	return ch
}

func (s *Server) handleJoinRoom(raw []byte, clientID string) {
	msg := &joinRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	// TODO handle error
	s.app.JoinRoom(clientID, msg.RoomID)
}

func (s *Server) handleLeaveRoom(raw []byte, clientID string) {
	msg := &leaveRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	// TODO handle error
	_ = s.app.LeaveRoom(clientID, msg.RoomID)
}

func (s *Server) handleCreateNewRoom(raw []byte) {
	msg := &createNewRoom{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.CreateNewRoom(msg.Name)
}

func (s *Server) handleSendMessage(raw []byte, clientID string) {
	msg := &sendMessage{}
	if err := json.Unmarshal(raw, msg); err != nil {
		slog.Error("failed to unmarshal join room message", "error_msg", err)
	}

	s.app.SendMessageToRoom(clientID, msg.RoomID, msg.Content)
}
