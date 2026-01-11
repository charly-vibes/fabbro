package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fabbro",
		Short: "A code review annotation tool",
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
