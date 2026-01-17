package main

import (
	"fmt"
	"net/http"

	"github.com/Mech654/chess-server/backend/lobby"
	"github.com/Mech654/chess-server/frontend-stuff"
)

func main() {
	mux := http.NewServeMux()

	frontend.RegisterRoutes(mux)

	lb := lobby.New()
	mux.HandleFunc("/ws/lobby", lb.ServeWS)

	fmt.Println("Starting server on :80")
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
