package main

import "golang.org/x/net/websocket"

type Server struct {
	clientCount    int
	clients        map[int]*Client
	addClientCh    chan *Client
	removeClientCh chan *Client
	messageCh      chan string
}

func NewServer() *Server {
	return &Server{
		clientCount:    0,
		clients:        map[int]*Client{},
		addClientCh:    make(chan *Client),
		removeClientCh: make(chan *Client),
		messageCh:      make(chan string),
	}
}

func (s *Server) addClient(c *Client) {
	s.clientCount++
	c.ID = s.clientCount
	s.clients[c.ID] = c
}

func (s *Server) removeClient(c *Client) {
	delete(s.clients, c.ID)
}

func (s *Server) sendMessage(m string) {
	for _, c := range s.clients {
		tmp := c
		go func() { tmp.Send(m) }()
	}
}

func (s *Server) Start() {
	for {
		select {
		case c := <-s.addClientCh:
			s.addClient(c)
		case c := <-s.removeClientCh:
			s.removeClient(c)
		case m := <-s.messageCh:
			s.sendMessage(m)
		}
	}
}

func (s *Server) WebsocketHandler() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		c := NewClient(ws, s.removeClientCh, s.messageCh)
		s.addClientCh <- c
		c.Start()
	})
}
