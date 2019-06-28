package main

import (
   "log"
   "net/http"

   "github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// configure the upgrader
var upgrader = websocket.Upgrader{}

// Define our message object
type Message struct {
    Type     string `json:"type"`
    Name     string `json:"name"`
    Message  string `json:"msg"`
}
type Echo struct {
    Message string `json:"msg"`
}

func main() {
    // Create a simple file server
    fs := http.FileServer(http.Dir("../public"))
    http.Handle("/", fs)

    // Configure websocket route
    http.HandleFunc("/ws", handleConnections)

    // Start listening for incoming chat messages
    go handleMessages()

    log.Println("http server started on :8181")
    err := http.ListenAndServe(":8181", nil)
    if err != nil {
       log.Fatal("ListenAndServer: ", err) 	
    }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    // Upgrade initail GET request to a websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Make sure we close the connection when the function returns
    defer ws.Close()

    // Register our new client
    clients[ws] = true

    for {
    	var msg Message
	// Read in a new message as JSON and map it to a Message object
	err := ws.ReadJSON(&msg)
	if err != nil {
	    log.Printf("error: %v", err)
	    delete(clients, ws)
	    break
	}

	// Send the new received message to the broadcast channel
	broadcast <- msg
    }
}

func handleMessages() {
    for {
        // Grab the next message from the broadcast channel
	msg := <-broadcast

	var who string
        if msg.Type == "TEACHER" {
            who = msg.Type + " (" + msg.Name + ") "
        } else {
            who = msg.Name + " "
        }
	who = "<b>" + who + "</b>"
	var toSend Echo
        toSend.Message = "<p>" + who + msg.Message + "</p>"

	// Send it out to every client that is currently connected
	for client := range clients {
	    err := client.WriteJSON(toSend)
	    if err != nil {
	        log.Printf("error: %v", err)
		client.Close()
		delete(clients, client)
	    }
	}
    }
}
