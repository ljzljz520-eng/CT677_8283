package training

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type StepClock struct {
	mu   sync.Mutex
	next time.Time
	step time.Duration
}

func NewStepClock(start time.Time, step time.Duration) *StepClock {
	return &StepClock{next: start, step: step}
}

func (c *StepClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.next
	c.next = c.next.Add(c.step)
	return current
}
