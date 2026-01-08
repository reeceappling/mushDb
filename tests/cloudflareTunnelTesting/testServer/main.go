package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/http", httpHandler)
	fmt.Println("WebSocket server started on :3728")
	err := http.ListenAndServe(":3728", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()
	lastMessage := ""
	ticker := time.NewTicker(2 * time.Second)
	countMutex := &sync.Mutex{}
	timesEchoed := 0
	go func() {
		for _ = range ticker.C {
			countMutex.Lock()
			if lastMessage != "" && timesEchoed < 5 {
				timesEchoed++
				toSend := fmt.Sprintf(`Echo %d. Last message received: %s`, timesEchoed, lastMessage)
				countMutex.Unlock()
				if err := conn.WriteMessage(websocket.TextMessage, []byte(toSend)); err != nil {
					fmt.Println("Error writing message:", err)

					return // TODO: break all
				}
			} else {
				countMutex.Unlock()
			}
		}
	}()

	for {
		// Read message from the client
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		if string(message) != lastMessage {
			countMutex.Lock()
			lastMessage = string(message)
			timesEchoed = 0
			countMutex.Unlock()
		}

		fmt.Printf("Received: %s\\n", message)
	}

}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body:"+err.Error(), 500)
		return
	}

	if _, err = w.Write(bs); err != nil {
		http.Error(w, "Error writing body:"+err.Error(), 500)
	}
}
