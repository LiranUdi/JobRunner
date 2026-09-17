package runner

import (
	"context"
	"errors"
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

		for i := 0; i < retries; i++ {
			attempts++
			statusCode, err = makeRequest(ctx, client, timeout, job.URL)
			if !shouldRetry(statusCode, err) {
				break
			}
		}

		results <- jobs.Result{
			ID:         job.ID,
			URL:        job.URL,
			StatusCode: statusCode,
			Attempts:   attempts,
			Err:        err,
		}
	}
}

func makeRequest(ctx context.Context, client *http.Client, timeout int, url string) (int, error) {
	ctx_timeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx_timeout, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func shouldRetry(statusCode int, err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error

	if errors.As(err, &netErr) {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return false
		}

		return true
	}

	switch {
	case statusCode >= 200 && statusCode <= 299:
		return false
	case statusCode >= 400 && statusCode <= 499:
		return false
	case statusCode >= 500 && statusCode <= 599:
		return true
	default:
		return false
	}
}
