package probe

import (
	"net/http"
	"sync"
)

const (
	HTTPHealthzEndpoint = "/healthz"
)

var (
	mu    sync.Mutex
	ready = false
)

func SetReady() {
	mu.Lock()
	ready = true
	mu.Unlock()
}

// HealthzHandler writes back the HTTP status code 200 if the operator is ready, and 500 otherwise
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	isReady := ready
	mu.Unlock()
	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
