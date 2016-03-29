package chat

import (
	"net/http"
	"strconv"

	//chat imports
	"flag"
)

func NewChat() Chat {
	var chat = Chat{
		hubs: make([]*Hub, 0),
	}
	return chat
}

type Chat struct {
	hubs []*Hub
}

func (c *Chat) addHub() {
	h := Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		messages:    make([][]byte, 0),
	}
	c.hubs = append(c.hubs, &h)
}

//Run function found in all applications to startup this module  TODO initialize with number of hubs to run.
//Pass in how many hubs to run in server - run.
func (c *Chat) Run(n int) {
	flag.Parse()
	//Add 3 hubs for testing
	for i := 0; i < n; i++ {
		c.addHub()
	}
	//Start each hub in a goroutine.  Register handler for hub on creation.
	for i, h := range c.hubs {
		go h.run()
		num := strconv.Itoa(i)
		http.HandleFunc("/"+num, h.executeHub)
		http.HandleFunc("/ws"+"/"+num, h.serveWs)
	}
}
