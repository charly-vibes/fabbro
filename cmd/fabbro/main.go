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

var version = "dev"

func main() {
	os.Exit(realMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func realMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	rootCmd := buildRootCmd(stdin, stdout)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func buildRootCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "fabbro",
		Short:   "A code review annotation tool",
		Version: version,
	}

	rootCmd.AddCommand(buildInitCmd(stdout))
	rootCmd.AddCommand(buildReviewCmd(stdin, stdout))
	rootCmd.AddCommand(buildApplyCmd(stdout))

	return rootCmd
}

func buildInitCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize fabbro in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.IsInitialized() {
				fmt.Fprintln(stdout, "fabbro already initialized")
				return nil
			}
			if err := config.Init(); err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}
			fmt.Fprintln(stdout, "Initialized fabbro in .fabbro/")
			return nil
		},
	}
}

func buildReviewCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
	var stdinFlag bool
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Start a review session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			if !stdinFlag {
				return fmt.Errorf("--stdin flag is required")
			}

			data, err := io.ReadAll(stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}

			content := string(data)
			sess, err := session.Create(content)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			fmt.Fprintf(stdout, "Created session: %s\n", sess.ID)

			model := tui.New(sess)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read content from stdin")
	return cmd
}

func buildApplyCmd(stdout io.Writer) *cobra.Command {
	var jsonFlag bool
	cmd := &cobra.Command{
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

				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			fmt.Fprintf(stdout, "Session: %s\n", sess.ID)
			fmt.Fprintf(stdout, "Annotations: %d\n", len(annotations))
			for _, a := range annotations {
				fmt.Fprintf(stdout, "  Line %d: [%s] %s\n", a.Line, a.Type, a.Text)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}
