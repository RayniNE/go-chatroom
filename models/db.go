package models

import (
	"net/http"
	"time"
)

type User struct {
	Id       int    `json:"user_id,omitempty"`
	Username string `json:"user_user_name,omitempty"`
	Email    string `json:"user_email,omitempty"`
	Password string `json:"user_password,omitempty"`
}

func (u *User) Validate(isLogin bool) error {
	appContext := "User.Validate"
	if u.Email == "" {
		return &CustomError{
			Code:       http.StatusBadRequest,
			Message:    "Invalid email",
			AppContext: appContext,
		}
	}

	if !isLogin {
		if u.Username == "" {
			return &CustomError{
				Code:       http.StatusBadRequest,
				Message:    "Invalid username",
				AppContext: appContext,
			}
		}
	}

	if u.Password == "" {
		return &CustomError{
			Code:       http.StatusBadRequest,
			Message:    "Invalid password",
			AppContext: appContext,
		}
	}

	return nil
}

type Chatroom struct {
	Id   string `json:"chatroom_id,omitempty"`
	Name string `json:"chatroom_name,omitempty"`
}

type ChatMessage struct {
	Id         int       `json:"chat_message_id,omitempty"`
	UserID     int       `json:"chat_message_user_id,omitempty"`
	ChatroomID string    `json:"chat_message_chatroom_id,omitempty"`
	Message    string    `json:"chat_message_message,omitempty"`
	CreatedAt  time.Time `json:"chat_message_created_at,omitempty"`
	UserName   string    `json:"chat_message_users_user_name,omitempty"`
}

type ChatRepository interface {
	GetChatroomByID(string) (*Chatroom, error)
	FindUserByEmail(string) (*User, error)
	GetUserByEmail(string) (*User, error)
	AddMessage(ChatMessage) (*int, error)
	AddUser(*User) (*int, error)
	GetAllChatRooms() ([]*Chatroom, error)
	GetChatroomMessages(string) ([]*ChatMessage, error)
}
