package main

import (
	"log"

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
	chatroom.Main()
}
