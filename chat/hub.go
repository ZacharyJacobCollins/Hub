package chat

import (
	"html/template"
	"net/http"
)

var homeTemplate = template.Must(template.ParseFiles("./html/chat_view.html"))

//hub is similar to a "channel" in slack.
type Hub struct {
	// Registered broadcasts.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection

	//TODO temporary message array
	messages [][]byte
}

//chat func
func (h *Hub) executeHub(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(w, r.Host)
}

//it's only using the last hub even though we created 3.
func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		//if there is message in channel broadcast, send to c of all h.connections
		case m := <-h.broadcast:
			//for testing purposes, checking to make sure appending is correct sptting out ascii shit in term currently

			m = append(m, []byte("\n")...)
			h.messages = append(h.messages, m)
			//TODO change from array
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
