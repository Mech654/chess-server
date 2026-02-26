package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Username string
	conn     *websocket.Conn
	send     chan []byte
	done 	chan struct{}
	handler  Handler
	mutex    sync.RWMutex
}

func NewClient(username string, conn *websocket.Conn, handler Handler) *Client {
	return &Client{
		Username: username,
		conn:     conn,
		send:     make(chan []byte, 256),
		done:    make(chan struct{}),
		handler:  handler,
	}
}

func (c *Client) SetHandler(h Handler) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.handler = h
}

func (c *Client) GetHandler() Handler {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.handler
}

func (c *Client) Send(message []byte) {
    select {
    case <-c.done:
        return
    case c.send <- message:
    default:
        log.Printf("Client %s mailbox full, closing", c.Username)
        c.Close()
    }
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		select {
		case <-c.done:
			return
		case message := <-c.send:
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) PrepareClose() {
	close(c.done)
}

func (c *Client) Close() {
	c.conn.Close()
}
