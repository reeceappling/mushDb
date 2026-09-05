package checkers

import (
	"sync/atomic"
)

func NewReadinessSensor() *ReadinessSensor {
	return &ReadinessSensor{
		ready: atomic.Bool{},
	}
}

type ReadinessSensor struct {
	ready atomic.Bool
}

func (rs *ReadinessSensor) Ready() bool {
	return rs.ready.Load()
}
func (rs *ReadinessSensor) SetReady() {
	rs.ready.Store(true)
}
func (rs *ReadinessSensor) SetUnready() {
	rs.ready.Store(false)
}

type Readiness string

const (
	ReadinessReady    Readiness = "ready"
	ReadinessNotReady Readiness = "not ready"
)

func ReadinessString(isReady bool) Readiness {
	if isReady {
		return ReadinessReady
	}
	return ReadinessNotReady
}

//var ready *atomic.Bool
//
//func Ready() bool {
//	return ready.Load()
//}
//func SetReady() {
//	ready.Store(true)
//}
//func SetUnready() {
//	ready.Store(false)
//}
//
//var ReadinessHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: use or no?
//	statusCode := http.StatusOK
//	status := "ready"
//
//	if !Ready() {
//		statusCode = http.StatusServiceUnavailable
//		status = "not ready"
//	}
//	result := map[string]string{"status": status}
//	w.WriteHeader(statusCode)
//	err := json.NewEncoder(w).Encode(result)
//	rfid.HandleHttpWriteError(err)
//})
