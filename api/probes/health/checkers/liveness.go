package checkers

import (
	"encoding/json"
	rfid "github.com/reeceappling/mushDb/api"
	"net/http"
)

type LivenessStatus string

const (
	LivenessAlive     LivenessStatus = "alive"
	LivenessUnhealthy LivenessStatus = "unhealthy"
)

var LivenessHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: use or no?
	w.WriteHeader(http.StatusOK)
	out := map[string]string{"status": string(LivenessAlive)}
	err := json.NewEncoder(w).Encode(out)
	rfid.HandleHttpWriteError(err)
})

//type Liveness struct {}
//// NewLiveness creates a new Liveness checker
//func NewLiveness() *Liveness {
//	return &Liveness{}
//}
//
//func (h *Liveness) Name() string {
//	return "liveness"
//}
//
//// Check performs the HTTP health check
//func (h *Liveness) Check(ctx context.Context) HealthCheckResult {
//	start := time.Now()
//	out := HealthCheckResult{
//		Name:      h.Name(),
//		Status: Status(LivenessAlive),
//		Message:   "", // TODO: ????
//		Duration:  start.Sub(start),
//		Timestamp: time.Now(), // TODO: ????
//	}
//	now := time.Now()
//	out.Timestamp = now
//	out.Duration = now.Sub(start)
//	return out
//}
