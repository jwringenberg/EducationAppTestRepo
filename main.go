package main

import (
   "log"
   "net/http"

   "github.com/gorilla/websocket"
)

type Status struct {
    LoggedIn bool
    Teacher bool
    Connected bool
}

var clients = make(map[*websocket.Conn]int) // connected clients
var broadcast = make(chan Message)          // broadcast channel

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
    // Initialize the DB
    InitDB()

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
    // Upgrade initial GET request to a websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Make sure we close the connection when the function returns
    defer ws.Close()

    // Register our new client
    clients[ws] = 1

    for {
    	var msg Message
	// Read in a new message as JSON and map it to a Message object
	err := ws.ReadJSON(&msg)
	if err != nil {
	    log.Printf("error: %v", err)
	    delete(clients, ws)
	    break
	}

	if clients[ws] > 2 {
	   if clients[ws] > 4 {
	      msg.Type = "TEACHER"
	   } else {
	      msg.Type = "STUDENT"
	   }
	   // Send the new received message to the broadcast channel
	   broadcast <- msg
	} else {
	   if msg.Type == "LOGIN" {
	      if ValidLogin(msg.Name, msg.Message) {
	      	 teacher := 0
		 if IsTeacher(msg.Name) {
		    teacher = 4;
		 }
	         clients[ws] = 3 + teacher
	         var toSend Echo
		 if clients[ws] > 4 {
	            toSend.Message = "TEACHER"
		 } else {
	            toSend.Message = "STUDENT"
		 }
	         err := ws.WriteJSON(toSend)
	         if err != nil {
	            log.Printf("error: %v", err)
		    delete(clients, ws)
		    return
	         }
	      } else {
	         var toSend Echo
	         toSend.Message = "Invalid login"
	         err := ws.WriteJSON(toSend)
	         if err != nil {
	            log.Printf("error: %v", err)
		    delete(clients, ws)
		    return
	         }
	      }
	   } else {
	      var toSend Echo
	      toSend.Message = "Login first"
	      err := ws.WriteJSON(toSend)
	      if err != nil {
	         log.Printf("error: %v", err)
		 delete(clients, ws)
		 return
	     }
	  }
       }  
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
