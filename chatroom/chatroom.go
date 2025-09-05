package chatroom

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/raynine/go-chatroom/chatroom/handlers"
	"github.com/raynine/go-chatroom/utils"
)

type ChatroomService struct {
	DB_URL string
	PORT   string
}

func (s *ChatroomService) Main() {
	r := mux.NewRouter()

	db, err := sql.Open("postgres", s.DB_URL)
	if err != nil {
		log.Fatalf("unable to create database connection: %s", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("unable to ping connection: %s", err.Error())
	}

	handler := handlers.NewHandler(db)

	r.HandleFunc("/user/", handler.AddUser).Methods("POST")
	r.HandleFunc("/login", handler.LoginUser).Methods("POST")

	s.protectedEndpoints(r, handler)

	log.Printf("Starting server in PORT %s", s.PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%s", s.PORT), muxhandlers.CombinedLoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalf("An error ocurred while starting server: %s\n", err.Error())
	}
}

func (service *ChatroomService) protectedEndpoints(router *mux.Router, handler *handlers.Handler) {
	subRouter := router.PathPrefix("/").Subrouter()
	subRouter.Use(utils.AuthMiddleware)

	subRouter.HandleFunc("/chatrooms", handler.GetAllChatrooms).Methods("GET")
	subRouter.HandleFunc("/ws/chatroom/{id}", handler.ConnectToChatroomWS)
}
