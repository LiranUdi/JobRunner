package runner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"jobrunner/internal/jobs"
)

func worker(ctx context.Context, jobsChan <-chan jobs.Job, results chan<- jobs.Result, client *http.Client, timeout, retries int) {
	for job := range jobsChan {
		var statusCode int
		var err error
		attempts := 0

		err = retry(ctx, retries, time.Second, time.Duration(timeout)*time.Second, func() error {
			attempts++
			return makeRequest(ctx, client, timeout, job.URL, &statusCode)
		})

		results <- jobs.Result{
			ID:         job.ID,
			URL:        job.URL,
			StatusCode: statusCode,
			Attempts:   attempts,
			Err:        err,
		}
	}
}

func makeRequest(ctx context.Context, client *http.Client, timeout int, url string, statusCode *int) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxTimeout, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return Permanent(err)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		var netErr net.Error

		if errors.As(err, &netErr) {
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) {
				return Permanent(err)
			}

			return err
		}
		return err
	}

	defer resp.Body.Close()
	*statusCode = resp.StatusCode

	switch {
	case *statusCode >= 200 && *statusCode <= 299:
		return nil
	case *statusCode == 429:
		return fmt.Errorf("Rate limit exceeded for %s", url)
	case *statusCode >= 400 && *statusCode <= 499:
		return Permanent(fmt.Errorf("Failed to handle job for %s due to Status Code: %d", url, *statusCode))
	case *statusCode >= 500 && *statusCode <= 599:
		return fmt.Errorf("Failed to handle job for %s due to Status Code: %d", url, *statusCode)
	default:
		return fmt.Errorf("Unknown Status Code: %d", *statusCode)
	}
}
