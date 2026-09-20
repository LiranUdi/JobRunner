package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"jobrunner/internal/runner"

	"github.com/spf13/cobra"
)

var (
	filename string
	workers  int
	timeout  int
	retries  int
)

var rootCmd = &cobra.Command{
	Use:   "jobrunner [text]",
	Short: "HTTP job runner",
	Long:  "HTTP job runner",
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := runner.Config{
			Filename:    filename,
			Workers:     workers,
			Timeout:     timeout,
			MaxAttempts: retries + 1,
		}
		if err := runner.Run(cmd.Context(), cfg); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&filename, "filename", "f", "", "Input file")
	rootCmd.MarkFlagRequired("filename")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of workers")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Timeout in seconds")
	rootCmd.Flags().IntVarP(&retries, "retries", "r", 3, "Number of retries (retries+1 total attempts)")
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
