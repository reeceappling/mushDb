package probes

import (
	"context"
	"encoding/json"
	rfid "github.com/reeceappling/mushDb/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func exampleDoNotUse() {
	// Initialize health service with 5-second timeout for checks
	healthService := NewHealthService(5 * time.Second)
	healthHandler := NewHandler(healthService)

	// Register external API health check
	healthService.Register(NewHTTPChecker(
		"payment-api",
		os.Getenv("PAYMENT_API_URL")+"/health",
		WithHttpCheckerTimeout(3*time.Second),
	))
	// Create HTTP server
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/healthz", healthHandler.LivenessHandler)
	mux.HandleFunc("/readyz", healthHandler.ReadinessHandler)
	mux.HandleFunc("/startupz", healthHandler.StartupHandler)

	// Application endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Application is running"))
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Run initialization in background
	go func() {
		log.Println("Starting initialization...")

		// Verify database connection
		_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//if err := db.PingContext(ctx); err != nil {
		//	log.Printf("Database not ready: %v", err)
		//	return
		//}

		// Run migrations, warm caches, etc.
		time.Sleep(5 * time.Second) // Simulate initialization

		// Mark startup as complete and ready for traffic
		healthHandler.SetStartupComplete()
		healthHandler.SetReady(true)
		log.Println("Application initialized and ready")

		// Graceful shutdown handling
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			log.Println("Shutdown signal received")

			// Mark as not ready to stop receiving new traffic
			healthHandler.SetReady(false)

			// Give load balancer time to update
			time.Sleep(5 * time.Second)

			// Shutdown server with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				log.Printf("Server shutdown error: %v", err)
			}
		}()
	}()
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

type HealthCheckResult struct {
	Name      string
	Status    Status
	Message   string
	Duration  time.Duration `json:"duration_ns"`
	Timestamp time.Time     `json:"duration_ns"`
}

type HealthResponse struct {
	Status    Status
	Timestamp time.Time
	Checks    map[string]HealthCheckResult `json:"checks,omitempty"`
}

type Checker interface {
	Name() string
	Check(context.Context) HealthCheckResult
}

type HealthService struct {
	mu       sync.RWMutex
	checkers []Checker
	timeout  time.Duration
}

func NewHealthService(timeout time.Duration) *HealthService {
	return &HealthService{
		checkers: make([]Checker, 0),
		timeout:  timeout,
	}
}
func (h *HealthService) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}
func (h *HealthService) CheckAll(ctx context.Context) HealthResponse {
	h.mu.RLock()
	checkers := make([]Checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	results := make(chan HealthCheckResult, len(checkers))
	var wg sync.WaitGroup
	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			results <- checker.Check(ctx)
		}(checker)
	}
	// Close results chan when checks complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Checks:    make(map[string]HealthCheckResult),
	}
	for result := range results {
		response.Checks[result.Name] = result
		if result.Status == StatusUnhealthy {
			response.Status = StatusUnhealthy
		} else if result.Status == StatusDegraded && response.Status != StatusUnhealthy {
			response.Status = StatusDegraded
		}
	}
	return response
}

// Handler provides HTTP handlers for health endpoints
type Handler struct {
	service       *HealthService
	ready         *atomic.Bool
	startupDone   *atomic.Bool
	livenessCheck func() bool
}

// NewHandler creates a new health handler
func NewHandler(service *HealthService) *Handler {
	ready := &atomic.Bool{}
	startupDone := &atomic.Bool{}
	return &Handler{
		service:     service,
		ready:       ready,
		startupDone: startupDone,
		livenessCheck: func() bool {
			return true // Default: always alive
		},
	}
}

// SetReady marks the application as ready to receive traffic
func (h *Handler) SetReady(isReady bool) {
	h.ready.Store(isReady)
}

// SetStartupComplete marks startup as complete
func (h *Handler) SetStartupComplete() {
	h.startupDone.Store(true)
}

// SetLivenessCheck sets a custom liveness check function
func (h *Handler) SetLivenessCheck(check func() bool) {
	h.livenessCheck = check
}

// LivenessHandler handles liveness probe requests
// Returns 200 if the application is alive, 503 otherwise
func (h *Handler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	var statusCode int = http.StatusOK
	w.Header().Set("Content-Type", "application/json") // TODO: ok?
	out := map[string]interface{}{
		"status": LivenessAlive,
	}
	if !h.livenessCheck() {
		statusCode = http.StatusServiceUnavailable
		out["status"] = LivenessUnhealthy
		out["message"] = "liveness check failed"
		return
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		rfid.HandleHttpWriteError(err)
	}
	return
}

// ReadinessHandler handles readiness probe requests
// Checks all registered dependencies before reporting ready
func (h *Handler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	var out interface{}
	var statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	// Check if marked as not ready
	if !Ready() {
		statusCode = http.StatusServiceUnavailable
		out = map[string]interface{}{
			"status":  ReadinessNotReady,
			"message": "application not ready",
		}
	} else {
		// Run all service dependency checks
		response := h.service.CheckAll(r.Context())
		if response.Status != StatusHealthy {
			statusCode = http.StatusServiceUnavailable
		}
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(out); err != nil {
		rfid.HandleHttpWriteError(err)
	}
}

// StartupHandler handles startup probe requests
func (h *Handler) StartupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !h.startupDone.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "starting",
			"message": "application still initializing",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
	})
}
