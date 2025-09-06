package repos

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/raynine/go-chatroom/models"
	"github.com/stretchr/testify/assert"
)

var chatRoomId string = "78fa7046-f8fc-4435-aed5-798b31cfd3e1"

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *ChatRepo) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}

	repo := NewChatRepo(db)
	return db, mock, repo
}

func TestGetChatroomByID(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	invalidId := chatRoomId + "2"

	t.Run("Chatroom does not exists", func(t *testing.T) {
		mock.ExpectQuery(getChatroomByIDQuery).WithArgs(invalidId).WillReturnError(sql.ErrNoRows)

		response, err := repo.GetChatroomByID(invalidId)
		assert.Nil(t, response)
		assert.Nil(t, err)

	})

	t.Run("Error while searching chatroom", func(t *testing.T) {
		mock.ExpectQuery(getChatroomByIDQuery).WithArgs(invalidId).WillReturnError(sql.ErrConnDone)

		response, err := repo.GetChatroomByID(invalidId)
		assert.Nil(t, response)
		assert.Equal(t, "error while searching for chatroom", err.Error())

	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(getChatroomByIDQuery).
			WithArgs(chatRoomId).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(chatRoomId, "CHATROOMTEST"))

		response, err := repo.GetChatroomByID(chatRoomId)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, chatRoomId, response.Id)

	})
}

func TestFindUserByEmail(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	notExistingEmail := "idonot@exist.com"

	user := &models.User{
		Username: "Raytest",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Id:       23,
	}

	t.Run("Email does not exists", func(t *testing.T) {
		mock.ExpectQuery(findUserByEmailQuery).WithArgs(notExistingEmail).WillReturnError(sql.ErrNoRows)

		response, err := repo.FindUserByEmail(notExistingEmail)
		assert.Nil(t, response)
		assert.Nil(t, err)
	})

	t.Run("Error while searching for user", func(t *testing.T) {
		mock.ExpectQuery(findUserByEmailQuery).WithArgs(notExistingEmail).WillReturnError(sql.ErrConnDone)

		response, err := repo.FindUserByEmail(notExistingEmail)
		assert.Nil(t, response)
		assert.Equal(t, "error while searching for user", err.Error())
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(findUserByEmailQuery).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password"}).
				AddRow(
					user.Id,
					user.Username,
					user.Email,
					user.Password,
				))

		response, err := repo.FindUserByEmail(user.Email)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, response.Id, user.Id)
		assert.Equal(t, response.Email, user.Email)
	})
}

func TestCheckIfEmailOrUsernameExists(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	user := &models.User{
		Username: "Raytest",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	t.Run("Error searching for user", func(t *testing.T) {
		mock.ExpectQuery(checkIfEmailOrUsernameExistsQuery).
			WithArgs(user.Email, user.Username).
			WillReturnError(sql.ErrConnDone)

		exists, err := repo.checkIfEmailOrUsernameExists(user.Email, user.Username)
		assert.Contains(t, err.Error(), "error while searching for user")
		assert.Equal(t, false, exists)
	})

	t.Run("User exists", func(t *testing.T) {
		mock.ExpectQuery(checkIfEmailOrUsernameExistsQuery).
			WithArgs(user.Email, user.Username).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.checkIfEmailOrUsernameExists(user.Email, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, true, exists)
	})

	t.Run("User does not exists", func(t *testing.T) {
		mock.ExpectQuery(checkIfEmailOrUsernameExistsQuery).
			WithArgs(user.Email, user.Username).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.checkIfEmailOrUsernameExists(user.Email, user.Username)
		// assert.Contains(t, err.Error(), "error while searching for user")
		assert.NoError(t, err)
		assert.Equal(t, false, exists)
	})

}

func TestGetUserByEmail(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	notExistingEmail := "idonot@exist.com"

	user := &models.User{
		Username: "Raytest",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Id:       23,
	}

	t.Run("Email does not exists", func(t *testing.T) {
		mock.ExpectQuery(GetUserByEmailQuery).WithArgs(notExistingEmail).WillReturnError(sql.ErrNoRows)

		response, err := repo.GetUserByEmail(notExistingEmail)
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("User with email: %s does not exists", notExistingEmail), err.Error())
	})

	t.Run("Error while searching for user", func(t *testing.T) {
		mock.ExpectQuery(GetUserByEmailQuery).WithArgs(notExistingEmail).WillReturnError(sql.ErrConnDone)

		response, err := repo.GetUserByEmail(notExistingEmail)
		assert.Nil(t, response)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("Error while searching for user: %s", notExistingEmail), err.Error())
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(GetUserByEmailQuery).
			WithArgs(user.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password"}).
				AddRow(
					user.Id,
					user.Username,
					user.Email,
					user.Password,
				))

		response, err := repo.GetUserByEmail(user.Email)
		assert.Nil(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, response.Id, user.Id)
		assert.Equal(t, response.Email, user.Email)
	})
}

func TestAddMessage(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	message := models.ChatMessage{
		UserID:     23,
		ChatroomID: chatRoomId,
		Message:    "Hello World!",
	}

	t.Run("Error while inserting message", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(addMessageQuery).WithArgs(message.UserID, message.ChatroomID, message.Message).WillReturnError(sql.ErrConnDone)

		mock.ExpectRollback()

		id, err := repo.AddMessage(message)
		assert.Contains(t, err.Error(), "error while inserting message")
		assert.Nil(t, id)
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(addMessageQuery).WithArgs(message.UserID, message.ChatroomID, message.Message).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(23))

		mock.ExpectCommit()

		id, err := repo.AddMessage(message)
		assert.NoError(t, err)
		assert.Equal(t, 23, *id)
	})
}

func TestAddUser(t *testing.T) {
	db, mock, repo := setupTestDB(t)
	defer db.Close()

	user := &models.User{
		Username: "Raytest",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	t.Run("Invalid user email", func(t *testing.T) {
		invalidUser := &models.User{
			Username: "Raytest",
			Password: "password123",
		}

		id, err := repo.AddUser(invalidUser)
		assert.Contains(t, err.Error(), "Invalid email")
		assert.Nil(t, id)
	})

	t.Run("Invalid user username", func(t *testing.T) {
		invalidUser := &models.User{
			Email:    "Raytest@raytest.com",
			Password: "password123",
		}

		id, err := repo.AddUser(invalidUser)
		assert.Contains(t, err.Error(), "Invalid username")
		assert.Nil(t, id)
	})

	t.Run("Invalid user password", func(t *testing.T) {
		invalidUser := &models.User{
			Username: "Raytest",
			Email:    "Raytest@raytest.com",
		}

		id, err := repo.AddUser(invalidUser)
		assert.Contains(t, err.Error(), "Invalid password")
		assert.Nil(t, id)
	})

	t.Run("Invalid email", func(t *testing.T) {
		invalidUser := &models.User{
			Username: "Raytest",
			Email:    "invalidemail+1.com",
			Password: "password123",
		}

		id, err := repo.AddUser(invalidUser)
		assert.Contains(t, err.Error(), fmt.Sprintf("invalid email: %s", invalidUser.Email))
		assert.Nil(t, id)
	})

	t.Run("Error while inserting user", func(t *testing.T) {
		mock.ExpectQuery(checkIfEmailOrUsernameExistsQuery).
			WithArgs(user.Email, user.Username).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectBegin()

		mock.ExpectQuery(addUserQuery).WithArgs(user.Username, user.Email, user.Password).WillReturnError(sql.ErrConnDone)

		mock.ExpectRollback()

		id, err := repo.AddUser(user)
		assert.Contains(t, err.Error(), "error while creating user")
		assert.Nil(t, id)
	})

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(checkIfEmailOrUsernameExistsQuery).
			WithArgs(user.Email, user.Username).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectBegin()

		mock.ExpectQuery(addUserQuery).WithArgs(user.Username, user.Email, user.Password).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectRollback()

		id, err := repo.AddUser(user)
		assert.NoError(t, err)
		assert.Equal(t, 1, *id)
	})

}
