package multilimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	Limiters []*rate.Limiter
}

func New(limiters []*rate.Limiter) *Limiter {
	return &Limiter{Limiters: limiters}
}

// AllowN calls rate.Limiter.AllowN for all clients.
func (l *Limiter) AllowN(now time.Time, n int) bool {
	allowed := true
	for _, limiter := range l.Limiters {
		allowed = allowed && limiter.AllowN(now, n)
	}
	return allowed
}

// Allow calls rate.Limiter.Allow for all clients.
// This is equivalent to Limiter.AllowN(ctx, 1).
func (l *Limiter) Allow() bool {
	return l.AllowN(time.Now(), 1)
}

// WaitN calls rate.Limiter.WaitN for all clients.
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	workers := len(l.Limiters)
	sem := make(chan bool, workers)

	var errs error
	var errL sync.Mutex
	for _, lim := range l.Limiters {
		go func() {
			if err := lim.WaitN(ctx, n); err != nil {
				errL.Lock()
				errs = fmt.Errorf("%v: %v", errs, err)
				errL.Unlock()
			}
			sem <- true
		}()
	}

	for i := 0; i < workers; i++ {
		<-sem
	}
	return errs
}

// Wait calls rate.Limiter.Wait for all clients.
// This is equivalent to Limiter.WaitN(ctx, 1).
func (l *Limiter) Wait(ctx context.Context) error {
	return l.WaitN(ctx, 1)
}
