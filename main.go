// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Event string `json:"event"`
}

type Score struct {
	Score      int    `json:"score"`
	Competitor string `json:"competitor"`
}

type ScoreMessage struct {
	Event  string  `json:"event"`
	Scores []Score `json:"scores"`
}

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "emojidome.xkcd.com", Path: "/2131/socket"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var m Message
			json.Unmarshal(message, &m)
			switch m.Event {
			case "score":
				var score ScoreMessage
				json.Unmarshal(message, &score)
				comp1 := score.Scores[0]
				comp2 := score.Scores[1]
				log.Printf("%s %d v %s %d", comp1.Competitor, comp1.Score, comp2.Competitor, comp2.Score)
				break
			case "start":
				log.Printf("start: %s", message)
				break
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
