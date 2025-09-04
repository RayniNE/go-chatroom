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

func (u *User) Validate() error {
	if u.Email == "" {
		return &CustomError{
			Code:    http.StatusBadRequest,
			Message: "Invalid email",
		}
	}

	if u.Username == "" {
		return &CustomError{
			Code:    http.StatusBadRequest,
			Message: "Invalid username",
		}
	}

	if u.Password == "" {
		return &CustomError{
			Code:    http.StatusBadRequest,
			Message: "Invalid password",
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
	ChatroomID int       `json:"chat_message_chatroom_id,omitempty"`
	Message    string    `json:"chat_message_message,omitempty"`
	CreatedAt  time.Time `json:"chat_message_created_at,omitempty"`
}
