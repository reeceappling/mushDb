package probes

import (
	"encoding/json"
	rfid "github.com/reeceappling/mushDb/api"
	"net/http"
)

type Liveness string

const (
	LivenessAlive     = "alive"
	LivenessUnhealthy = "unhealthy"
)

var LivenessHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: use or no?
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
	if err != nil {
		rfid.HandleHttpWriteError(err)
	}
})
