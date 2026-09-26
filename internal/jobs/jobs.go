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

type JSONLResult struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Duration   string `json:"duration"`
	Attempts   int    `json:"attempts"`
	Error      string `json:"error,omitempty"`
}

func ReadJSONL(filePath string) ([]Job, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var jobs []Job

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var job Job
		if err := json.Unmarshal(line, &job); err != nil {
			fmt.Fprintf(os.Stderr, "error unmarshaling line: %d | malformed JSON\n", lineNum)
			continue
		}

		jobs = append(jobs, job)
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return jobs, nil
}
