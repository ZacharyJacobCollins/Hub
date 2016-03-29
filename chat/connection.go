package chat

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Three hours is max time wihtout sending anything
	writeWait = 180 * time.Second
	// Three hours is max time wihtout sending anything
	pongWait = 180 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

//Buffer size for messages, stopping data loss
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection
	ws *websocket.Conn
	//name of the connection or the id of the person TODO
	name string
	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		h.broadcast <- message
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte, h *Hub) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	nickname := []byte(c.name + "")
	payload = append(nickname, payload...)
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump(h *Hub) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	//infinite loop, for as long as within connection times
	for {
		select {
		//if there is a message and is ok, add it to the send channel
		case message, ok := <-c.send:
			//send disconnected if not ok
			if !ok {
				c.write(websocket.CloseMessage, []byte{}, h)
				return
			}
			if err := c.write(websocket.TextMessage, message, h); err != nil {
				return
			}
			//Send empty ping message
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}, h); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket connection requests from the peer.
func (h *Hub) serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	//register connection to hub
	h.register <- c

	//if there are messages in message array, send messages to connection.  This is sending as one big message though for whatever reason.  TODO  MAKE SURE THEY"RE SEPARATE IN THE PUMP
	for _, m := range h.messages {
		if err := c.write(websocket.TextMessage, m, h); err != nil {
			return
		}
	}

	//Place Read pump in goroutine for indefinite listening.
	go c.readPump(h)
	//write pump contains for select so is indefinitely listening.
	c.writePump(h)
}
