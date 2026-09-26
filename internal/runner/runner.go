package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"jobrunner/internal/jobs"
	"net/http"
	"os"
	"sync"
)

func Run(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	client := &http.Client{}

	jobList, err := jobs.ReadJSONL(cfg.Filename)
	if err != nil {
		return fmt.Errorf("reading JSONL file: %w", err)
	}

	jobsChan := make(chan jobs.Job)
	resultsChan := make(chan jobs.Result)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, jobsChan, resultsChan, client, cfg.Timeout, cfg.MaxAttempts)
		}()
	}

	go func() {
	feedLoop:
		for _, job := range jobList {
			select {
			case jobsChan <- job:
			case <-ctx.Done():
				break feedLoop
			}
		}
		close(jobsChan)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var successfulJobs int
	var failedJobs int
	for result := range resultsChan {

		errMsg := ""
		if result.Err != nil {
			failedJobs++
			errMsg = result.Err.Error()
		} else {
			successfulJobs++
		}

		jsonlResult := jobs.JSONLResult{
			ID:         result.ID,
			URL:        result.URL,
			StatusCode: result.StatusCode,
			Duration:   result.Duration.String(),
			Attempts:   result.Attempts,
			Error:      errMsg,
		}

		bytes, err := json.Marshal(jsonlResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error marshaling result: %v\n", err)
			continue
		}
		fmt.Println(string(bytes))
	}

	if ctx.Err() != nil {
		fmt.Fprintf(os.Stderr, "\ninterrupted: completed [%d/%d] jobs\n", successfulJobs, successfulJobs+failedJobs)
		return nil
	}

	fmt.Fprintf(os.Stderr, "\n[%d/%d] jobs finished successfully!\n", successfulJobs, successfulJobs+failedJobs)

	return nil
}
