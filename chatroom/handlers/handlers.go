package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/raynine/go-chatroom/models"
	"github.com/raynine/go-chatroom/repos"
	"github.com/raynine/go-chatroom/utils"
)

type Handler struct {
	repo *repos.ChatRepo
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		repo: repos.NewChatRepo(db),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var hubs = make(map[string]*models.Hub)

func (handler *Handler) GetAllChatrooms(w http.ResponseWriter, r *http.Request) {

	response, err := handler.repo.GetAllChatRooms()
	if err != nil {
		utils.EncodeErrorResponse(w, models.CustomError{
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
		utils.EncodeErrorResponse(w, models.CustomError{
			Message: "Invalid chatroom ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	chatroom, err := handler.repo.GetChatroomByID(id)
	if err != nil {
		utils.EncodeErrorResponse(w, models.CustomError{
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if chatroom == nil {
		utils.EncodeErrorResponse(w, models.CustomError{
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

	hub, ok = hubs[id]
	if !ok {
		hub = models.NewHub()
		hubs[id] = hub
		go hub.Run()
	}

	clientId := uuid.New().String()

	client := &models.Client{
		Id:   clientId,
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
