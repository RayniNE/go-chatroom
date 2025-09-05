package repos

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"github.com/raynine/go-chatroom/models"
)

type ChatRepo struct {
	db *sql.DB
}

func NewChatRepo(db *sql.DB) *ChatRepo {
	return &ChatRepo{
		db: db,
	}
}

func (repo *ChatRepo) GetChatroomByID(id string) (*models.Chatroom, error) {
	query := "SELECT * FROM public.chatrooms WHERE id = $1"

	chatroom := &models.Chatroom{}

	err := repo.db.QueryRow(query, id).Scan(&chatroom.Id, &chatroom.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while searching for chatroom with ID %s: %s", id, err.Error())
		return nil, fmt.Errorf("error while searching for chatroom")
	}

	return chatroom, nil
}

func (repo *ChatRepo) FindUserByEmail(email string) (*models.User, error) {
	query := "SELECT * FROM public.users WHERE LOWER(email) = LOWER($1)"

	user := &models.User{}

	err := repo.db.QueryRow(query, email).Scan(&user.Id, &user.Username, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while searching for email ID %s: %s", email, err.Error())
		return nil, fmt.Errorf("error while searching for user")
	}

	return user, nil
}

func (repo *ChatRepo) checkIfEmailOrUsernameExists(email, username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM public.users WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($2))"

	exists := false

	err := repo.db.QueryRow(query, email, username).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		log.Printf("An error ocurred while searching for email ID %s: %s", email, err.Error())
		return false, fmt.Errorf("error while searching for user")
	}

	return exists, nil
}

func (repo *ChatRepo) GetUserByEmail(email string) (*models.User, error) {
	appContext := "ChatRepo.GetUserByEmail"
	query := "SELECT * FROM public.users WHERE LOWER(email) = LOWER($1)"

	user := &models.User{}

	err := repo.db.QueryRow(query, email).Scan(&user.Id, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &models.CustomError{
				Message:    fmt.Sprintf("User with email: %s does not exists", user.Email),
				Code:       http.StatusNotFound,
				AppContext: appContext,
			}
		}

		log.Printf("An error ocurred while searching for email ID %s: %s", email, err.Error())
		return nil, &models.CustomError{
			Message:    fmt.Sprintf("Error while searching for user: %s", user.Email),
			Code:       http.StatusInternalServerError,
			AppContext: appContext,
		}
	}

	return user, nil
}

func (repo *ChatRepo) AddMessage(chatMessage models.ChatMessage) (*int, error) {
	query := `
			INSERT INTO 
				public.messages(id, user_id, chatroom_id, message, created_at)
			VALUES (default, $1, $2, $3, CURRENT_TIMESTAMP) returning id
		`

	var newId *int

	tx, err := repo.db.Begin()
	if err != nil {
		log.Printf("An error ocurred while starting transaction: %s", err.Error())
		return nil, fmt.Errorf("error while adding message")
	}

	defer tx.Rollback()

	err = repo.db.QueryRow(query, chatMessage.UserID, chatMessage.ChatroomID, chatMessage.Message).Scan(&newId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while inserting message: %s", err.Error())
		return nil, fmt.Errorf("error while inserting message")
	}

	tx.Commit()

	return newId, nil
}

func (repo *ChatRepo) AddUser(user *models.User) (*int, error) {

	err := user.Validate(false)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	_, err = mail.ParseAddress(user.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %s", user.Email)
	}

	exists, err := repo.checkIfEmailOrUsernameExists(user.Email, user.Username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("email: %s or username: %s is already registered", user.Email, user.Username)
	}

	query := `
			INSERT INTO 
				public.users(id, username, email, password)
			VALUES (default, $1, $2, $3) returning id
		`

	var newId *int

	tx, err := repo.db.Begin()
	if err != nil {
		log.Printf("An error ocurred while starting transaction: %s", err.Error())
		return nil, fmt.Errorf("error while creating user")
	}

	defer tx.Rollback()

	err = repo.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&newId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while creating user: %s", err.Error())
		return nil, fmt.Errorf("error while creating user")
	}

	tx.Commit()

	return newId, nil
}

func (repo *ChatRepo) GetAllChatRooms() ([]*models.Chatroom, error) {
	query := `
			SELECT * FROM
				public.chatrooms
	`

	rows, err := repo.db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while getting all chatrooms: %s", err.Error())
		return nil, fmt.Errorf("error while getting all chatrooms")
	}

	response := []*models.Chatroom{}

	for rows.Next() {
		chatroom := &models.Chatroom{}

		err = rows.Scan(
			&chatroom.Id,
			&chatroom.Name,
		)
		if err != nil {
			log.Printf("An error ocurred while getting scanning chatrooms: %s", err.Error())
			return nil, fmt.Errorf("error while getting all chatrooms")
		}

		response = append(response, chatroom)
	}

	return response, nil
}

func (repo *ChatRepo) GetChatroomMessages(chatroomId string) ([]*models.ChatMessage, error) {
	query := `
			SELECT messages.*,
			users.username
				 FROM
			public.messages
			INNER JOIN public.users ON users.id = messages.user_id
			WHERE chatroom_id = $1
			ORDER BY created_at ASC
			LIMIT 50
	`

	rows, err := repo.db.Query(query, chatroomId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		log.Printf("An error ocurred while getting chatroom messages: %s", err.Error())
		return nil, fmt.Errorf("error while getting chatroom messages")
	}

	response := []*models.ChatMessage{}

	for rows.Next() {
		message := &models.ChatMessage{}

		err = rows.Scan(
			&message.Id,
			&message.UserID,
			&message.ChatroomID,
			&message.Message,
			&message.CreatedAt,
			&message.UserName,
		)
		if err != nil {
			log.Printf("An error ocurred while getting scanning chatroom messages: %s", err.Error())
			return nil, fmt.Errorf("error while getting all messages")
		}

		response = append(response, message)
	}

	return response, nil
}
