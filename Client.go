package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

type Client struct {
	ID             int
	ws             *websocket.Conn
	removeClientCh chan *Client
	messageCh      chan string
}

func NewClient(ws *websocket.Conn, remove chan *Client, message chan string) *Client {
	return &Client{
		ws:             ws,
		removeClientCh: remove,
		messageCh:      message,
	}
}

func (c *Client) Start() {
	for {
		var m string
		if err := websocket.Message.Receive(c.ws, &m); err != nil {
			c.removeClientCh <- c
			return
		} else {
			var message Message
			if err := json.Unmarshal([]byte(m), &message); err == nil {
				message.Message = html.EscapeString(message.Message)
				message.Time = time.Now().Format("2006/01/02 15:04:05 MST")
				username, err := GetUserName(message.Hash)
				if err != nil {
					continue
				}
				message.Username = html.EscapeString(username)
				if b, err := json.Marshal(message); err == nil {
					c.messageCh <- string(b)
				}
			} else {
				fmt.Sprintln(os.Stderr, err)
			}
		}
	}
}

func (c *Client) Send(m string) {
	if err := websocket.Message.Send(c.ws, m); err != nil {
		fmt.Fprintln(os.Stderr, "error:", m)
	}
}

func (c *Client) Close() {
	c.ws.Close()
}
