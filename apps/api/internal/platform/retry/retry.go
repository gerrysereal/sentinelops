package retry

import (
	"context"
	"time"
)

type Policy struct {
	Attempts int
	Backoff  time.Duration
}

func (p Policy) Do(ctx context.Context, fn func() error) error {
	attempts := p.Attempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := p.Backoff
	if backoff <= 0 {
		backoff = 200 * time.Millisecond
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == attempts {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff * time.Duration(attempt)):
			}
			continue
		}
		return nil
	}
	return lastErr
}
