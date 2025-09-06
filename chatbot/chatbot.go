package chatbot

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/raynine/go-chatroom/interfaces"
	"github.com/raynine/go-chatroom/models"
)

type chatBot struct {
	ch   *amqp.Channel
	Hubs map[string]*models.Hub
	User *models.User
	repo interfaces.DBRepo
}

type stockInformation struct {
	Symbol string
	Date   string
	Time   string
	Open   string
	High   string
	Low    string
	Close  string
	Volume string
}

var ENDPOINT = "https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv"

func NewChatBot(hubs map[string]*models.Hub, repo interfaces.DBRepo, botEmail string, ch *amqp.Channel) *chatBot {

	user, err := repo.GetUserByEmail(botEmail)
	if err != nil {
		log.Fatalf("An error ocurred while finding bot email: %s", err.Error())
	}

	return &chatBot{
		Hubs: hubs,
		User: user,
		ch:   ch,
		repo: repo,
	}
}

func (cb *chatBot) ConsumeStockRequests() {
	msgs, _ := cb.ch.Consume("stock_requests", "", true, false, false, false, nil)
	for d := range msgs {

		msg := &models.ChatMessage{}

		err := json.Unmarshal(d.Body, &msg)
		if err != nil {
			log.Println("Error parsing request:", err.Error())
			continue
		}

		log.Println("Stock request received: ", msg)

		stockCode := msg.Message[7:]

		stock, err := getStockInformation(stockCode)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		hub, ok := cb.Hubs[msg.ChatroomID]
		if !ok {
			log.Printf("Hub: %s does not exists", msg.ChatroomID)
			continue
		}

		stockMessage := &models.ChatMessage{
			Message: fmt.Sprintf(
				"%s quote is $%s per share",
				stock.Symbol,
				stock.Close),
			UserID:     cb.User.Id,
			ChatroomID: hub.ChatroomId,
		}

		log.Println("Bot message: ", stockMessage)

		body, _ := json.Marshal(stockMessage)

		cb.ch.Publish("", "chatroom_messages", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	}
}

func (cb *chatBot) ConsumeChatroomMessages() {
	msgs, _ := cb.ch.Consume("chatroom_messages", "", true, false, false, false, nil)
	for d := range msgs {
		log.Println("Received message from bot")
		msg := &models.ChatMessage{}

		err := json.Unmarshal(d.Body, &msg)
		if err != nil {
			log.Println("Error parsing request:", err.Error())
			continue
		}

		hub := cb.Hubs[msg.ChatroomID]

		log.Println("Publishing message: ", msg)
		_, err = cb.repo.AddMessage(*msg)
		if err != nil {
			log.Printf("An error ocurred while trying to save message from WS: %s\n", err.Error())
			continue
		}

		hub.Broadcast <- msg

	}
}

func getStockInformation(stockCode string) (*stockInformation, error) {
	res, err := http.Get(fmt.Sprintf(ENDPOINT, stockCode))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("Request failed with status code: %d", res.StatusCode)
		return nil, fmt.Errorf("Request failed")
	}

	reader := csv.NewReader(res.Body)

	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	stockInformation := &stockInformation{}

	for {
		record, err := reader.Read()
		log.Println("Record", record)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error al leer el CSV:", err)
			return nil, err
		}

		stockInformation.Symbol = record[0]
		stockInformation.Date = record[1]
		stockInformation.Time = record[2]
		stockInformation.Open = record[3]
		stockInformation.High = record[4]
		stockInformation.Low = record[5]
		stockInformation.Close = record[6]
		log.Println("Stock ingo", stockInformation)

	}

	return stockInformation, nil

}
