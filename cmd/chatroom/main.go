package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/raynine/go-chatroom/chatroom"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	service := chatroom.ChatroomService{
		DB_URL: dbUrl,
		PORT:   port,
	}

	service.Main()
}
