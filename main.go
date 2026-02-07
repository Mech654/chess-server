package main

import (
	"fmt"
	"net/http"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/Mech654/chess-server/backend/game"
	"github.com/Mech654/chess-server/backend/ws"
	"github.com/Mech654/chess-server/frontend-stuff"
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	wsServer := ws.NewServer()
	lobby := game.NewLobby()
	lobbyHandler := game.NewHandler(lobby)

	mux.HandleFunc("/join", auth.JoinHandler)
	mux.HandleFunc("/ws/lobby", func(w http.ResponseWriter, r *http.Request) {
		username, err := auth.GetUsernameFromToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		wsServer.ServeWS(w, r, username, lobbyHandler)
	})

	fmt.Println("Starting server on :8888")
	err := http.ListenAndServe(":8888", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
