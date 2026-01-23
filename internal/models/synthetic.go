package models

import "time"

// SyntheticCheckResult holds the latest outcome for a synthetic probe.
type SyntheticCheckResult struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Target      string    `json:"target"`
	Status      string    `json:"status"`
	LatencyMs   int64     `json:"latency_ms"`
	LastRun     time.Time `json:"last_run"`
	Message     string    `json:"message"`
	Tags        []string  `json:"tags"`
}
