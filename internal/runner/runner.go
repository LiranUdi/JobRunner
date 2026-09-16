package runner

import (
	"context"
	"fmt"
	"jobrunner/internal/jobs"
	"net/http"
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
			worker(ctx, jobsChan, resultsChan, client, cfg.Timeout, cfg.Retries)
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

	for result := range resultsChan {
		if result.Err != nil {
			fmt.Printf("Error processing job %s: %v\n", result.ID, result.Err)
			continue
		}
		fmt.Printf("Result for job %s: %s. StatusCode: %d\n", result.ID, result.URL, result.StatusCode)
	}

	fmt.Println("All jobs processed successfully")
	return nil
}
