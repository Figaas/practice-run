package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/icrowley/fake"
	"github.com/segmentio/ksuid"

	"practice-run/internal/ports"
)

// IMPROVEMENT: upgrader could be configurable with some config file
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

type Server struct {
	app ports.App
}

func NewServer(app ports.App) *Server {
	return &Server{app: app}
}

func (s *Server) Run() {
	// IMPROVEMENT: server could be configurable with some config file
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

	name, clientID := extractClientDetails(r)
	s.handleNewClientConnection(conn, clientID, name)

	loopBreaker, stop := setupLoopBreaker()
	s.handleDisconnectedClient(conn, clientID, stop)

	for {
		select {
		case <-loopBreaker:
			return
		default:
			s.handleMessages(conn, clientID, stop)
		}
	}
}

func setupLoopBreaker() (chan struct{}, func()) {
	loopBreaker := make(chan struct{})
	stop := func() { loopBreaker <- struct{}{} }
	return loopBreaker, stop
}

func extractClientDetails(r *http.Request) (string, string) {
	// IMPROVEMENTS: client ID and name should be returned during authentication and extracted from request metadata
	name := r.Header.Get("name")
	clientID := r.Header.Get("client_id")
	if name == "" {
		name = fake.FullName()
	}

	if clientID == "" {
		clientID = ksuid.New().String()
	}
	return name, clientID
}

func (s *Server) handleMessages(conn *websocket.Conn, clientID string, stop func()) {
	messageType, raw, err := conn.ReadMessage()
	if errors.Is(err, net.ErrClosed) {
		stop()
		return
	}
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

	ctx := context.WithValue(context.Background(), messageIDKey{}, b.MessageID)
	switch b.Type {
	case "join_room":
		s.handleJoinRoom(ctx, raw, clientID)
	case "leave_room":
		s.handleLeaveRoom(ctx, raw, clientID)
	case "create_new_room":
		s.handleCreateNewRoom(ctx, raw)
	case "send_message":
		s.handleSendMessage(ctx, raw, clientID)
	default:
		// IMPROVEMENT: DisconnectClient should receive custom type with DisconnectReason
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseProtocolError, "invalid message type"),
			time.Now().Add(time.Second),
		)

		s.app.DisconnectClient(context.Background(), clientID)
	}

	return
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header["Origin"]
	if len(origin) == 0 {
		return true
	}
	u, err := url.Parse(origin[0])
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host) ||
		strings.EqualFold(u.Host, "websocketking.com") ||
		strings.EqualFold(u.Host, "localhost:8080")
}
