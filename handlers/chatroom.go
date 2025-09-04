package handlers

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/raynine/go-chatroom/models"
	"github.com/raynine/go-chatroom/repos"
	"github.com/raynine/go-chatroom/utils"
)

var PORT = flag.String("port", ":8080", "HTTP Service port address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var hubs = make(map[string]*models.Hub)

func StartWS() {
	flag.Parse()

	r := mux.NewRouter()

	connectionString := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("unable to create database connection: %s", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("unable to ping connection: %s", err.Error())
	}

	repo := repos.NewDBRepo(db)

	r.HandleFunc("/ws/chatroom/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, ok := vars["id"]
		if !ok {
			utils.EncodeErrorResponse(w, models.CustomError{
				Message: "Invalid chatroom ID",
				Code:    http.StatusBadRequest,
			})
			return
		}

		chatroom, err := repo.GetChatroomByID(id)
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
	})

	log.Printf("Starting server in PORT %s", *PORT)
	err = http.ListenAndServe(*PORT, muxhandlers.CombinedLoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalf("An error ocurred while starting server: %s\n", err.Error())
	}
}
