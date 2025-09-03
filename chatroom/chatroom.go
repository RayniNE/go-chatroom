package chatroom

import (
	"flag"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raynine/go-chatroom/models"
)

var PORT = flag.String("port", ":8080", "HTTP Service port address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func severHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed", http.StatusNotFound)
		return
	}

	// http.ServeFile(w, r, filepath.Abs())
}

func StartWS() {
	flag.Parse()
	hub := models.NewHub()

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("An error ocurred while upgrading connection to WS: %s\n", err.Error())
			return
		}

		clientId := uuid.New()

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
	err := http.ListenAndServe(*PORT, nil)
	if err != nil {
		log.Fatalf("An error ocurred while starting server: %s\n", err.Error())
	}
}
