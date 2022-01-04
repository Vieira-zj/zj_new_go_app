package pkg

import "context"

// Watcher .
type Watcher interface {
	Run(context.Context)
}
