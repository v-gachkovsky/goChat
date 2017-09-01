package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"github.com/gorilla/websocket"
)

type Client struct {
	ws *websocket.Conn
	send chan []byte
}

type Room struct {
	broadcast chan []byte
	join 			chan *Client
	leave 		chan *Client
	clients 	map[*Client]bool

}

func (room *Room) listeningForClients() {

}

var Rooms = make(map[string]Room)

func main() {
	Rooms["first"] = Room{
		broadcast: make(chan []byte),
		join: make(chan *Client),
		leave: make(chan *Client),
		clients: make(map[*Client]bool),
	}

	Rooms["second"] = Room{
		broadcast: make(chan []byte),
		join: make(chan *Client),
		leave: make(chan *Client),
		clients: make(map[*Client]bool),
	}

	for _, room := range Rooms {
		log.Println(room)
	}

	router := mux.NewRouter().StrictSlash(true)
	public := http.FileServer(http.Dir("./public"))

	router.HandleFunc("/ws/{id}", handleRoom)
	router.PathPrefix("/").Handler(public)
	http.Handle("/", router)

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func handleRoom(w http.ResponseWriter, r *http.Request) {
	log.Println(mux.Vars(r))
}