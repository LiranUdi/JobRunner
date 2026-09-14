package runner

import (
	"context"
	"net/http"
	"time"

	"jobrunner/internal/jobs"
)

func worker(jobsChan <-chan jobs.Job, results chan<- jobs.Result, client *http.Client, timeout, retries int) {
	for job := range jobsChan {
		var statusCode int
		var err error
		attempts := 0

		for i := 0; i < retries; i++ {
			attempts++
			statusCode, err = makeRequest(client, timeout, job.URL)
			if err == nil {
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

func makeRequest(client *http.Client, timeout int, url string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
