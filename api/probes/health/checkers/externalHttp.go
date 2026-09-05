package checkers

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPChecker verifies external HTTP service availability
type HTTP struct {
	name    string
	url     string
	client  *http.Client
	method  string
	headers map[string]string
}

// HTTPCheckerOption configures the HTTP checker
type HTTPCheckerOption func(*HTTP)

// WithMethod sets the HTTP method for the health check request
func WithHttpCheckerMethod(method string) HTTPCheckerOption {
	return func(h *HTTP) {
		h.method = method
	}
}

// WithHeaders sets custom headers for the health check request
func WithHttpCheckerHeaders(headers map[string]string) HTTPCheckerOption {
	return func(h *HTTP) {
		h.headers = headers
	}
}

// WithHttpCheckerTimeout sets a custom timeout for the HTTP client
func WithHttpCheckerTimeout(timeout time.Duration) HTTPCheckerOption {
	return func(h *HTTP) {
		h.client.Timeout = timeout
	}
}

// NewHTTP creates a new HTTP health checker
func NewHTTP(name, url string, opts ...HTTPCheckerOption) *HTTP {
	checker := &HTTP{
		name:   name,
		url:    url,
		method: "GET",
		client: &http.Client{Timeout: 5 * time.Second},
	}

	for _, opt := range opts {
		opt(checker)
	}

	return checker
}

func (h *HTTP) Name() string {
	return h.name
}

// Check performs the HTTP health check
func (h *HTTP) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()
	result := HealthCheckResult{
		Name:      h.name,
		Timestamp: time.Now(),
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, h.method, h.url, nil)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("failed to create request: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Add custom headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := h.client.Do(req)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("request failed: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 500 {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("service returned %d", resp.StatusCode)
	} else if resp.StatusCode >= 400 {
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("service returned %d", resp.StatusCode)
	} else {
		result.Status = StatusHealthy
		result.Message = fmt.Sprintf("service returned %d", resp.StatusCode)
	}

	result.Duration = time.Since(start)
	return result
}
