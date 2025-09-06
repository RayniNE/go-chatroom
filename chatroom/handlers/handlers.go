package handlers

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/raynine/go-chatroom/interfaces"
	"github.com/raynine/go-chatroom/models"

	"github.com/raynine/go-chatroom/utils"
)

type Handler struct {
	repo interfaces.DBRepo
	hubs map[string]*models.Hub
	ch   *amqp.Channel
}

func NewHandler(repo interfaces.DBRepo, ch *amqp.Channel, hubs map[string]*models.Hub) *Handler {
	return &Handler{
		repo: repo,
		hubs: hubs,
		ch:   ch,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (handler *Handler) GetAllChatrooms(w http.ResponseWriter, r *http.Request) {
	response, err := handler.repo.GetAllChatRooms()
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	data := models.ServerResponse{Data: response, Code: http.StatusOK}

	utils.EncodeResponse(w, data)
}

func (handler *Handler) ConnectToChatroomWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Invalid chatroom ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	chatroom, err := handler.repo.GetChatroomByID(id)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if chatroom == nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Chatroom not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("An error ocurred while upgrading connection to WS: %s\n", err.Error())
		return
	}

	hub := &models.Hub{}

	hub, ok = handler.hubs[id]
	if !ok {
		hub = models.NewHub(id, handler.repo)
		handler.hubs[id] = hub
		go hub.Run()
	}

	userId, userName, err := utils.GetUserDataFromContext(r.Context())
	if err != nil {
		utils.EncodeErrorResponse(w, err)
		return
	}

	client := &models.Client{
		Id:       userId,
		UserName: userName,
		Hub:      hub,
		Conn:     conn,
		Ch:       handler.ch,
		Send:     make(chan *models.ChatMessage),
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (handler *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}

	err := utils.DecodePayload(r, &user)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Invalid User",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = user.Validate(false)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	user.Password = hashedPassword

	_, err = handler.repo.AddUser(user)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	utils.EncodeResponse(w, models.ServerResponse{Code: http.StatusCreated})
}

func (handler *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}

	err := utils.DecodePayload(r, &user)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Invalid User",
			Code:    http.StatusBadRequest,
		})
		return
	}

	err = user.Validate(true)
	if err != nil {
		utils.EncodeErrorResponse(w, err)
		return
	}

	existingUser, err := handler.repo.GetUserByEmail(user.Email)
	if err != nil {
		utils.EncodeErrorResponse(w, err)
		return
	}

	match := utils.CheckPassword(user.Password, existingUser.Password)
	if !match {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Provided password does not match",
			Code:    http.StatusBadRequest,
		})
		return
	}

	token, err := utils.CreateJWTToken(existingUser)
	if err != nil {
		utils.EncodeErrorResponse(w, &models.CustomError{
			Message: "Error while creating token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	utils.EncodeResponse(w, models.ServerResponse{
		Code: http.StatusOK,
		Data: map[string]any{
			"token": token,
		},
	})
}
