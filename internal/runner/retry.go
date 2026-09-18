package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

func retry(ctx context.Context, attempts int, baseDelay, maxDelay time.Duration, f func() error) error {
	var err error
	delay := baseDelay

	for attempt := 1; attempt <= attempts; attempt++ {
		if err = f(); err == nil {
			return nil
		}

		var pe *permanentError
		if errors.As(err, &pe) {
			// do not retry
			return pe.err
		}

		if attempt == attempts {
			break
		}

		// Jitter
		sleep := time.Duration(rand.Int64N(int64(delay)))
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return fmt.Errorf("retry aborted after %d attempts: %w", attempt, ctx.Err())
		}

		if delay *= 2; delay > maxDelay {
			delay = maxDelay
		}
	}
	return fmt.Errorf("all %d attempts failed, last error: %w", attempts, err)
}

// mark failures that a are not worth retrying
type permanentError struct{ err error }

func (p *permanentError) Error() string {
	return p.err.Error()
}

func (p *permanentError) Unwrap() error {
	return p.err
}

func Permanent(err error) error {
	return &permanentError{err: err}
}
