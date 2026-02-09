package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/Mech654/chess-server/backend/game"
	"github.com/Mech654/chess-server/backend/ws"
	"github.com/Mech654/chess-server/frontend-stuff"
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	lobbyHandler := game.NewHandler(game.NewLobby())

	mux.HandleFunc("/join", auth.JoinHandler)
	mux.HandleFunc("/ws/lobby", ws.WSHandler(lobbyHandler))
	// TODO: Put another for match handler, "/ws/match"

	fmt.Println("Starting server on :8888")
	log.Fatal(http.ListenAndServe(":8888", mux))
}
