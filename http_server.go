package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// holds connection
var clients = make(map[*websocket.Conn]bool)

// allows us to broad message on ws to all connected clients
var broadcast = make(chan Job)

// Configure the upgrader
var upgrader = websocket.Upgrader{}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	log.Printf("We got a new connection %v", r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Register our new client
	clients[conn] = true
	for {
		msg := "{}"
		// Read in a new message as JSON and map it to a Message object
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, conn)
			break
		}
		// Send the newly received message to the broadcast channel
		for {
			msg := "{}"
			// Read in a new message as JSON and map it to a Message object
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				// let's remove connection from map
				delete(clients, conn)
				break
			}
		}
	}
	// Make sure we close the connection when the function returns
	defer conn.Close()
}

func StartServer() {
	port := 8000
	// Create a simple file server
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	// send message from broadcast channel to all connected clients
	go handleMessages()
	// setup ws route
	http.HandleFunc("/ws", handleConnections)
	log.Printf("Started HTTP Interface on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
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
