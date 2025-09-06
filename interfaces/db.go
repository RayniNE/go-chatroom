package interfaces

import "github.com/raynine/go-chatroom/models"

type DBRepo interface {
	GetChatroomByID(string) (*models.Chatroom, error)
	FindUserByEmail(string) (*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	AddMessage(models.ChatMessage) (*int, error)
	AddUser(*models.User) (*int, error)
	GetAllChatRooms() ([]*models.Chatroom, error)
	GetChatroomMessages(string) ([]*models.ChatMessage, error)
}
