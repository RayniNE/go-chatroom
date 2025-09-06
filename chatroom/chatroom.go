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
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/raynine/go-chatroom/chatbot"
	"github.com/raynine/go-chatroom/chatroom/handlers"
	"github.com/raynine/go-chatroom/interfaces"
	"github.com/raynine/go-chatroom/models"
	"github.com/raynine/go-chatroom/repos"
	"github.com/raynine/go-chatroom/utils"
)

type ChatroomService struct {
	DB_URL        string
	PORT          string
	CHATBOT_EMAIL string
	RABBIT_MQ_URL string
}

var hubs = make(map[string]*models.Hub)

func (s *ChatroomService) Main() {
	r := mux.NewRouter()

	db, err := sql.Open("postgres", s.DB_URL)
	if err != nil {
		log.Fatalf("unable to create database connection: %s", err.Error())
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("unable to ping connection: %s", err.Error())
	}

	repo := repos.NewChatRepo(db)

	log.Println("Starting bot...")
	ch := s.startBroker(repo, s.CHATBOT_EMAIL)

	handler := handlers.NewHandler(repo, ch, hubs)

	r.HandleFunc("/user/", handler.AddUser).Methods("POST")
	r.HandleFunc("/login", handler.LoginUser).Methods("POST")

	s.protectedEndpoints(r, handler)

	log.Printf("Starting server in PORT %s", s.PORT)
	err = http.ListenAndServe(fmt.Sprintf(":%s", s.PORT), muxhandlers.CombinedLoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalf("An error ocurred while starting server: %s\n", err.Error())
	}
}

func (s *ChatroomService) startBroker(repo interfaces.DBRepo, botEmail string) *amqp.Channel {
	conn, err := amqp.Dial(s.RABBIT_MQ_URL)
	if err != nil {
		log.Fatalf("An error ocurred while starting rabbit mq: %s\n", err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("An error ocurred while starting rabbit mq channel: %s\n", err.Error())
	}

	_, err = ch.QueueDeclare(
		"stock_requests", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Fatalf("An error ocurred while declaring stock requests queue: %s\n", err.Error())
	}

	_, err = ch.QueueDeclare(
		"chatroom_messages", // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Fatalf("An error ocurred while declaring chatroom messages queue: %s\n", err.Error())
	}

	chatBot := chatbot.NewChatBot(hubs, repo, botEmail, ch)

	go chatBot.ConsumeStockRequests()
	go chatBot.ConsumeChatroomMessages()

	return ch
}

func (service *ChatroomService) protectedEndpoints(router *mux.Router, handler *handlers.Handler) {
	subRouter := router.PathPrefix("/").Subrouter()
	subRouter.Use(utils.AuthMiddleware)

	subRouter.HandleFunc("/chatrooms/", handler.AddChatroom).Methods("POST")
	subRouter.HandleFunc("/chatrooms", handler.GetAllChatrooms).Methods("GET")
	subRouter.HandleFunc("/ws/chatroom/{id}", handler.ConnectToChatroomWS)
}
