package clock

import "time"

// SystemClock implements application.Clock using the system clock.
type SystemClock struct{}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
