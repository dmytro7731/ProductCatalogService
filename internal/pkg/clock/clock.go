package clock

import (
	"time"
)

// Clock is an interface for time operations, allowing for testability.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the actual system time.
type RealClock struct{}

// NewRealClock creates a new RealClock.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current time.
func (c *RealClock) Now() time.Time {
	return time.Now().UTC()
}

// MockClock is a Clock implementation that returns a fixed time.
// Useful for testing.
type MockClock struct {
	fixedTime time.Time
}

// NewMockClock creates a new MockClock with the given time.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{fixedTime: t}
}

// Now returns the fixed time.
func (c *MockClock) Now() time.Time {
	return c.fixedTime
}

// SetTime sets a new fixed time.
func (c *MockClock) SetTime(t time.Time) {
	c.fixedTime = t
}

// Advance advances the clock by the given duration.
func (c *MockClock) Advance(d time.Duration) {
	c.fixedTime = c.fixedTime.Add(d)
}
