package main

import (
	"practice-run/internal/app"
	"practice-run/internal/http"
)

func main() {
	chat := app.NewApp()
	srv := http.NewServer(chat)
	srv.Run()
}

// IMPROVEMENT: implement graceful shutdown
// IMPROVEMENT: for better performance we could introduce protobuf messages format (perhaps with gRPC) for internal communication
// IMPROVEMENT: for a production use the app should support retrieving state from some stateful service
