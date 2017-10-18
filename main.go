package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

var (
	counter uint64
)

type message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type responseRtmStart struct {
	OK    bool         `json:"ok"`
	Error string       `json:"error"`
	URL   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	ID string `json:"id"`
}

func main() {
	token := os.Getenv("SLACK_BOT_TOKEN")
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	start := &responseRtmStart{}
	json.NewDecoder(resp.Body).Decode(start)
	log.Printf("%+v", start)
	ws, err := websocket.Dial(start.URL, "", "https://api.slack.com/")
	if err != nil {
		log.Fatal(err)
	}
	for {
		msg, err := receive(ws)
		if err != nil {
			log.Println("Receive error:", err)
		}

		if msg.Type != "message" {
			continue
		}
		log.Println(msg.Text)
		reply := message{
			Type:    "message",
			Channel: msg.Channel,
			Text:    "I got your message",
		}
		go func(m message) {
			err := send(ws, m)
			if err != nil {
				log.Println("Send error:", err)
			}
		}(reply)
	}
}

func receive(ws *websocket.Conn) (m message, err error) {
	err = websocket.JSON.Receive(ws, &m)
	return
}

func send(ws *websocket.Conn, m message) error {
	m.ID = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(ws, m)
}
