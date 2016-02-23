package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-neutrino/neutrino/src/common/client"
	"github.com/go-neutrino/neutrino/src/common/config"
	"github.com/go-neutrino/neutrino/src/common/log"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats"
)

var (
	upgrader    = client.NewWebsocketUpgrader()
	connections []*websocket.Conn
)

func jobsHandler(m *nats.Msg) {
	log.Info("Got message " + string(m.Data))

	for i, c := range connections {
		log.Info("Sending message to connection:", strconv.Itoa(i+1), string(m.Data))
		c.WriteMessage(websocket.TextMessage, m.Data)
	}
}

func main() {
	c := client.NewNatsClient(config.Get(config.KEY_QUEUE_ADDR))
	//TODO: handle subscription after the connection to nats is lost and restored
	for {
		if c.GetConnection() != nil {
			c.Subscribe(config.Get(config.CONST_REALTIME_JOBS_SUBJ), jobsHandler)
			break
		} else {
			time.Sleep(time.Second * 1)
		}
	}

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			panic(err)
			return
		}

		//TODO: handle removal of dead connections
		connections = append(connections, conn)
	})

	port := config.Get(config.KEY_BROKER_PORT)
	log.Info("Starting WS service on port " + port)
	http.ListenAndServe(port, nil)
}