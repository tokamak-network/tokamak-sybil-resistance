package database

import (
	"time"

	"golang.org/x/sync/semaphore"
)

// APIConnectionController is used to limit the SQL open connections used by the API
type APIConnectionController struct {
	smphr   *semaphore.Weighted
	timeout time.Duration
}

// NewAPIConnectionController initialize APIConnectionController
func NewAPIConnectionController(maxConnections int, timeout time.Duration) *APIConnectionController {
	return &APIConnectionController{
		smphr:   semaphore.NewWeighted(int64(maxConnections)),
		timeout: timeout,
	}
}