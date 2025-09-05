package chatbot

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/raynine/go-chatroom/models"
)

type chatBot struct {
	MessagesChan chan *models.ChatMessage
	Hub          *models.Hub
	User         *models.User
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

func NewChatBot(messageChan chan *models.ChatMessage, hub *models.Hub) *chatBot {
	return &chatBot{
		MessagesChan: messageChan,
		Hub:          hub,
	}
}

func (cb *chatBot) StartBot() {

	for message := range cb.MessagesChan {
		if len(message.Message) > 7 && message.Message[:7] == "/stock" {
			stockCode := message.Message[8:]

			stock, err := getStockInformation(stockCode)
			if err != nil {
				log.Println(err.Error())
				return
			}

			stockMessage := &models.ChatMessage{
				Message: fmt.Sprintf(
					"SYMBOL: %s; OPEN: %s HIGH: %s; LOW: %s; CLOSE: %s; VOLUME: %s; DATE: %s; TIME: %s;",
					stock.Symbol,
					stock.Open,
					stock.High,
					stock.Low,
					stock.Close,
					stock.Volume,
					stock.Date,
					stock.Time),
				UserID:     cb.User.Id,
				ChatroomID: cb.Hub.ChatroomId,
			}

			cb.Hub.Broadcast <- stockMessage
		}
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
	}

	return stockInformation, nil

}
