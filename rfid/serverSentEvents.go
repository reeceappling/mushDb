package rfid

import (
	"encoding/json"
	"net/http"
)

// TODO: UNSURE IF ACTUALLY USING SSEs

var (
	_ ServerSentEvent = readersListChangeEvent{}
)

type ServerSentEvent interface {
	write(w http.ResponseWriter) error
}

type readersListChangeEvent struct {
	readerNames []string
}

func (e readersListChangeEvent) write(w http.ResponseWriter) error {
	bs, err := json.Marshal(map[string]any{
		"EventType": RfidReadersListChangeEvent,
		"Data":      e.readerNames,
	})
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

type ServerSentEventType string

const (
	RfidReadersListChangeEvent = ServerSentEventType("RfidReadersListChangeEvent")
	UserPermsChangeEvent       = ServerSentEventType("UserPermsChangeEvent")
)

var SSEvents = make(chan ServerSentEvent, 10) // 10 ok??

func ServerSentEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow all origins. You may want to restrict this to specific origins in a production environment.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	select {
	case event := <-SSEvents: // TODO: DONT RECEIVE DIRECTLY SINCE MULTIPLE CLIENTS WILL BE LISTENING
		event.write(w)
		w.(http.Flusher).Flush()
	case <-r.Context().Done():
		break
	}

	// Simulate closing the connection
	closeNotify := w.(http.CloseNotifier).CloseNotify() // TODO: ok?
	<-closeNotify
}
