package repos

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/raynine/go-chatroom/models"
)

type DBRepo struct {
	db *sql.DB
}

func NewDBRepo(db *sql.DB) *DBRepo {
	return &DBRepo{
		db: db,
	}
}

func (repo *DBRepo) GetChatroomByID(id string) (*models.Chatroom, error) {
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
