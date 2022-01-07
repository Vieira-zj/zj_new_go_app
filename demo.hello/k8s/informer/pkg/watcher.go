package pkg

import "context"

const (
	// MaxRetries .
	MaxRetries = 3
)

// Watcher .
type Watcher interface {
	Run(context.Context)
}
