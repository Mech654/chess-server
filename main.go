package main

import (
	"fmt"
	"net/http"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/Mech654/chess-server/backend/lobby"
	"github.com/Mech654/chess-server/frontend-stuff"
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	mux.HandleFunc("/join", auth.JoinHandler)
	mux.HandleFunc("/ws/lobby", lobby.New().ServeWS)

	fmt.Println("Starting server on :8888")
	err := http.ListenAndServe(":8888", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
