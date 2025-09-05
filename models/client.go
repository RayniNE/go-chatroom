package models

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Id       int
	UserName string
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan *ChatMessage
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			isWSError := websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			if isWSError {
				log.Printf("An error ocurred while trying to ready message from WS: %s\n", err.Error())
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		userMessage := string(message)

		chatMessage := &ChatMessage{
			Message:    userMessage,
			UserID:     c.Id,
			ChatroomID: c.Hub.ChatroomId,
			CreatedAt:  time.Now(),
		}

		_, err = c.Hub.repo.AddMessage(*chatMessage)
		if err != nil {
			log.Printf("An error ocurred while trying to save message from WS: %s\n", err.Error())
			break
		}

		c.Hub.ChatBotMessageChan <- chatMessage
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("An error ocurred while trying get WS next writer: %s\n", err.Error())
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("An error ocurred while encoding user message: %s\n", err.Error())
				return
			}

			w.Write(messageBytes)

			n := len(c.Send)
			for range n {
				w.Write(newline)

				nextMessage := <-c.Send
				messageBytes, err := json.Marshal(nextMessage)
				if err != nil {
					log.Printf("An error ocurred while encoding user message: %s\n", err.Error())
					return
				}

				w.Write(messageBytes)
			}

			err = w.Close()
			if err != nil {
				log.Printf("An error ocurred while trying to close WS Writer: %s\n", err.Error())
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Printf("An error ocurred while trying to write message %s\n", err.Error())
				return
			}
		}
	}
}
