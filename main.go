package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
	"fmt"
)

var upgrader = websocket.Upgrader{}

type Hub struct {
	clients map[*Client]bool
	broadcast     chan []byte
	addClient     chan *Client
	removeClient  chan *Client
}

var Hubs = make(map[string] Hub)

func (hub *Hub) start() {
	for {
		select {
		case conn := <-hub.addClient:
			hub.clients[conn] = true
		case conn := <-hub.removeClient:
			if _, ok := hub.clients[conn]; ok {
				delete(hub.clients, conn)
				close(conn.send)
			}
		case message := <-hub.broadcast:
			for conn := range hub.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(hub.clients, conn)
				}
			}
		}
	}
}

type Client struct {
	ws *websocket.Conn
	send chan []byte
}

func (c *Client) write() {
	defer func() {
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *Client) read(hubId string) {
	defer func() {
		Hubs[hubId].removeClient <- c
		c.ws.Close()
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			Hubs[hubId].removeClient <- c
			c.ws.Close()
			break
		}

		Hubs[hubId].broadcast <- message
	}
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	hubId := vars["hubId"]
	fmt.Println(hubId)

	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		http.NotFound(res, req)
		return
	}

	client := &Client{
		ws:   conn,
		send: make(chan []byte),
	}

	Hubs[hubId].addClient <- client

	go client.write()
	go client.read(hubId)
}

func homePage(res http.ResponseWriter, req *http.Request){
	http.ServeFile(res, req, "./public/index.html")
}

func main(){
	Hubs["first"] = Hub{
		broadcast:     make(chan []byte),
		addClient:     make(chan *Client),
		removeClient:  make(chan *Client),
		clients:       make(map[*Client]bool),
	}

	Hubs["second"] = Hub{
		broadcast:     make(chan []byte),
		addClient:     make(chan *Client),
		removeClient:  make(chan *Client),
		clients:       make(map[*Client]bool),
	}

	for _, hub := range Hubs {
		go hub.start()
	}

	//go hub.start()
	fs := http.FileServer(http.Dir("./public"))

	router := mux.NewRouter()
	router.HandleFunc("/ws/{hubId}", wsPage)

	router.PathPrefix("/").Handler(fs)
	http.Handle("/", router)

	http.ListenAndServe(":8080", nil)
}
