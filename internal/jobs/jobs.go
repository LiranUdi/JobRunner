package jobs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Job struct {
	ID  string
	URL string
}

type Result struct {
	ID         string
	URL        string
	StatusCode int
	Duration   time.Duration
	Attempts   int
	Err        error
}

func ReadJSONL(filePath string) ([]Job, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var jobs []Job

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var job Job
		if err := json.Unmarshal(line, &job); err != nil {
			return nil, fmt.Errorf("error unmarshaling line: %w", err)
		}

		jobs = append(jobs, job)
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return jobs, nil
}
