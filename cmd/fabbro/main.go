package main

import (
	"fmt"
	"os"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fabbro",
		Short: "A code review annotation tool",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize fabbro in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.IsInitialized() {
				fmt.Println("fabbro already initialized")
				return nil
			}
			if err := config.Init(); err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}
			fmt.Println("Initialized fabbro in .fabbro/")
			return nil
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
