package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

func main() {
	fs := http.FileServer(http.Dir("./public"))

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/ws", handleConnections)

	router.PathPrefix("/").Handler(fs)
	http.Handle("/", router)

	go handleMessages()

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	clients[ws] = true

 	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
						log.Printf("error: %v", err)
						delete(clients, ws)
						break
		}

		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}