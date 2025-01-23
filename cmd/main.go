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
