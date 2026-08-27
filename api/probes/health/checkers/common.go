package checkers

import (
	"context"
	"time"
)

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
