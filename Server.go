package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/net/websocket"
)

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
	mutex := new(sync.Mutex)
	for {
		select {
		case c := <-s.addClientCh:
			s.addClient(c)
		case c := <-s.removeClientCh:
			s.removeClient(c)
		case m := <-s.messageCh:
			go messageLog(m, mutex)
			s.sendMessage(m)
		}
	}
}

func messageLog(s string, m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
	f, err := os.OpenFile("message.json", os.O_CREATE|os.O_RDWR, 0666)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	var messages []Message
	err = json.Unmarshal(b, &messages)
	if err != nil {
		messages = make([]Message, 0)
	}
	f.Close()
	var message Message
	err = json.Unmarshal([]byte(s), &message)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(messages) > 20 {
		messages = messages[2:]
	}
	f, err = os.Create("message.json")
	defer f.Close()
	messages = append(messages, message)
	b, err = json.Marshal(messages)
	_, err = f.Write(b)
	if err != nil {
		fmt.Println("failed")
		return
	}
}

func (s *Server) WebsocketHandler() websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		c := NewClient(ws, s.removeClientCh, s.messageCh)
		s.addClientCh <- c
		c.Start()
	})
}
