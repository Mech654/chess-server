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

var (
	matchHandlers = make(map[string]*game.MatchHandler)
	globalLobby = game.NewLobbyHandler(game.NewLobby())
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	mux.HandleFunc("/join", auth.JoinHandler)
	mux.HandleFunc("/ws/lobby", newHandler(ws.LobbyHandler))
	mux.HandleFunc("/ws/match", newHandler(ws.MatchHandler))

	fmt.Println("Starting server on :8888")
	log.Fatal(http.ListenAndServe(":8888", mux))
}



func newHandler(handlerType ws.HandlerType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := auth.GetUsernameFromToken(r)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var handler ws.Handler
		switch handlerType {
			case ws.LobbyHandler:
				handler = globalLobby
			case ws.MatchHandler:
				handler = game.FindPlayerMatch(matchHandlers, username)
				if handler == nil {
					http.Error(w, "No active match found", http.StatusNotFound)
					return
				}
		}

		ws.WSHandler(handler, w, r, username)
	}
}
