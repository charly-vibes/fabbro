package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
	"github.com/charly-vibes/fabbro/internal/tui"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
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

	var stdinFlag bool
	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "Start a review session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			if !stdinFlag {
				return fmt.Errorf("--stdin flag is required")
			}

			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}

			content := string(data)
			sess, err := session.Create(content)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			fmt.Printf("Created session: %s\n", sess.ID)

			// Launch TUI
			model := tui.New(sess)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
	reviewCmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read content from stdin")

	var jsonFlag bool
	applyCmd := &cobra.Command{
		Use:   "apply [session-id]",
		Short: "Apply annotations from a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]
			sess, err := session.Load(sessionID)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}

			annotations, _, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse FEM: %w", err)
			}

			if jsonFlag {
				output := struct {
					SessionID   string           `json:"session_id"`
					Annotations []fem.Annotation `json:"annotations"`
				}{
					SessionID:   sess.ID,
					Annotations: annotations,
				}

				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			fmt.Printf("Session: %s\n", sess.ID)
			fmt.Printf("Annotations: %d\n", len(annotations))
			for _, a := range annotations {
				fmt.Printf("  Line %d: [%s] %s\n", a.Line, a.Type, a.Text)
			}
			return nil
		},
	}
	applyCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(applyCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
