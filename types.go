package promise

import (
	"sync"
)

// Promise represents the result of an asynchronous operation.
type Promise struct {
	result any
	err    error
	once   sync.Once
	done   chan struct{}
}

func (p *Promise) Await() (any, any) {
	panic("unimplemented")
}

// AllSettled waits until all of the promises have settled (either resolved or rejected).
// It returns a slice of objects that describe the outcome of each promise.
type Settlement struct {
	Status string
	Value  any
	Reason error
}

// AggregateError aggregates multiple errors into one.
type AggregateError struct {
	Errors []error
}
