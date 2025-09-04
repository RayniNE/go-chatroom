package chatroom

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/raynine/go-chatroom/chatroom/handlers"
)

var PORT = flag.String("port", ":8080", "HTTP Service port address")

func Main() {
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

	handler := handlers.NewHandler(db)

	r.HandleFunc("/chatrooms", handler.GetAllChatrooms).Methods("GET")
	r.HandleFunc("/ws/chatroom/{id}", handler.ConnectToChatroomWS)

	log.Printf("Starting server in PORT %s", *PORT)
	err = http.ListenAndServe(*PORT, muxhandlers.CombinedLoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalf("An error ocurred while starting server: %s\n", err.Error())
	}
}
