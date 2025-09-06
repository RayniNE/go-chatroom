package models

import (
	"log"
	"sync"
)

type Hub struct {
	repo       ChatRepository
	ChatroomId string
	mu         sync.RWMutex

	Clients map[*Client]bool

	Broadcast  chan *ChatMessage
	Register   chan *Client
	Unregister chan *Client
}

// A hub is considered a chatroom. It handles the logic to broadcast the messages to all the clients connected to itself
func NewHub(chatroomId string, repo ChatRepository) *Hub {
	return &Hub{
		mu:         sync.RWMutex{},
		repo:       repo,
		ChatroomId: chatroomId,
		Broadcast:  make(chan *ChatMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

/*
Run method is a goroutine that gets launched when the first user connects to the chatroom.
Manages all the clients register, unregister and broadcast logic
*/
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("New client registered: %s", client.UserName)
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()

			chatMessages, err := h.repo.GetChatroomMessages(h.ChatroomId)
			if err != nil {
				log.Println("An error ocurred while getting chatroom messages:", err.Error())
				delete(h.Clients, client)
				close(client.Send)
				client.Conn.Close()
			}

			for _, chatMessage := range chatMessages {
				client.Send <- chatMessage
			}

		case client := <-h.Unregister:
			h.mu.Lock()
			_, ok := h.Clients[client]
			if ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					delete(h.Clients, client)
					close(client.Send)
				}
			}
			h.mu.Unlock()
		}
	}
}
