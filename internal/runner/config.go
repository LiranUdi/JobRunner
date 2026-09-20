package runner

import "errors"

type Config struct {
	Filename    string
	Workers     int
	Timeout     int
	MaxAttempts int
}

func (c Config) Validate() error {
	if c.Filename == "" {
		return errors.New("input file is required")
	}
	if c.Workers < 1 {
		return errors.New("number of workers must be at least 1")
	}
	if c.MaxAttempts < 1 {
		return errors.New("number of retries must be at least 1")
	}
	if c.Timeout < 1 {
		return errors.New("timeout must be at least 1 second")
	}
	return nil
}
