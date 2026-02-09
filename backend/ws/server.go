package ws

import (
	"log"
	"net/http"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/gorilla/websocket"
)

var server *Server = NewServer()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler interface {
	HandleConnect(client *Client)
	HandleMessage(client *Client, data []byte)
	HandleDisconnect(client *Client)
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func WSHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := auth.GetUsernameFromToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		server.ServeWS(w, r, username, handler)
	}
}

func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request, username string, handler Handler) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := NewClient(username, conn, handler)

	handler.HandleConnect(client)

	go client.writePump()

	defer func() {
		client.GetHandler().HandleDisconnect(client)
		client.Close()
	}()

	log.Printf("%s connected", username)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		client.GetHandler().HandleMessage(client, message)
	}
}
