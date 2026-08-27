package probes

import (
	"encoding/json"
	rfid "github.com/reeceappling/mushDb/api"
	"net/http"
	"sync/atomic"
)

type Readiness string

const (
	ReadinessReady    Readiness = "ready"
	ReadinessNotReady Readiness = "not ready"
)

func ReadinessString() Readiness {
	if Ready() {
		return ReadinessReady
	}
	return ReadinessNotReady
}

var ready *atomic.Bool

func Ready() bool {
	return ready.Load()
}
func SetReady() {
	ready.Store(true)
}
func SetUnready() {
	ready.Store(false)
}

var ReadinessHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: use or no?
	statusCode := http.StatusOK
	status := "ready"

	if !Ready() {
		statusCode = http.StatusServiceUnavailable
		status = "not ready"
	}
	result := map[string]string{"status": status}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		rfid.HandleHttpWriteError(err)
	}
})
