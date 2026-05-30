package application

import "time"

// Clock provides the current time (testable).
type Clock interface {
	Now() time.Time
}
