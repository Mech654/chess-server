package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/Mech654/chess-server/backend/game"
	"github.com/Mech654/chess-server/backend/ws"
	"github.com/Mech654/chess-server/frontend-stuff"
)

var (
	matchHandlers = sync.Map{}
	globalLobby = game.NewLobbyHandler(game.NewLobby(&matchHandlers))
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	mux.HandleFunc("/join", auth.JoinHandler)
	mux.HandleFunc("/ws/lobby", newHandler(ws.LobbyHandler))
	mux.HandleFunc("/ws/match/{id}", newHandler(ws.MatchHandler))

	fmt.Println("Starting server on :8888")
	log.Fatal(http.ListenAndServe(":8888", mux))
}



func newHandler(handlerType ws.HandlerType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := auth.GetUsernameFromToken(r)
		id := r.PathValue("id")

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var handler ws.Handler
		switch handlerType {
			case ws.LobbyHandler:
				matchHandler := game.CheckOngoingMatch(&matchHandlers, username)
				if matchHandler != nil {
					handler = matchHandler
				} else {
					handler = globalLobby
				}
			case ws.MatchHandler:
				val, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					http.Error(w, "Invalid match ID", http.StatusBadRequest)
					return
				}

				matchHandler := game.FindPlayerMatch(&matchHandlers, val, username)
				if matchHandler == nil {
					http.Error(w, "No active match found", http.StatusNotFound)
					return
				}
				handler = matchHandler
		}

		ws.WSHandler(handler, w, r, username)
	}
}
