package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
	"github.com/charly-vibes/fabbro/internal/tui"
	"github.com/charly-vibes/fabbro/internal/tutor"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

const maxInputBytes = 10 * 1024 * 1024 // 10MB

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
		Use:   "fabbro",
		Short: "A code review annotation tool",
		Long: `"For you, il miglior fabbro"
  — after T.S. Eliot, The Waste Land

A code review annotation tool with a terminal UI.`,
		Version: version,
	}

	rootCmd.AddCommand(buildInitCmd(stdout))
	rootCmd.AddCommand(buildReviewCmd(stdin, stdout))
	rootCmd.AddCommand(buildApplyCmd(stdout))
	rootCmd.AddCommand(buildSessionCmd(stdout))
	rootCmd.AddCommand(buildCompletionCmd())
	rootCmd.AddCommand(buildTutorCmd(stdout))

	return rootCmd
}

func buildCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for fabbro.

To load completions:

Bash:
  source <(fabbro completion bash)

Zsh:
  source <(fabbro completion zsh)

Fish:
  fabbro completion fish | source

PowerShell:
  fabbro completion powershell | Out-String | Invoke-Expression
`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func buildInitCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize fabbro in the current directory",
		Long: `Initialize fabbro in the current directory by creating a .fabbro/ folder.

Pre-conditions:
  - You must be in a directory where you want to use fabbro.
  - The .fabbro/ directory must not already exist (idempotent: re-running is safe).

Post-conditions:
  - A .fabbro/ directory is created containing session storage.
  - You can now run 'fabbro review' to start annotating files.`,
		Example: `  # Initialize fabbro in the current project
  fabbro init

  # Typical workflow after init
  fabbro init && fabbro review myfile.go`,
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
	var jsonFlag bool
	cmd := &cobra.Command{
		Use:   "review [file]",
		Short: "Start a review session",
		Long: `Start a new review session to annotate code with FEM (Fabbro Edit Markers).

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - Provide either a file path or content via --stdin.

Post-conditions:
  - A new session is created and stored in .fabbro/sessions/.
  - The TUI opens for interactive annotation.
  - Session ID is printed for later reference.`,
		Example: `  # Review a specific file
  fabbro review main.go

  # Review content piped from another command
  git show HEAD:main.go | fabbro review --stdin

  # Review a file from a different directory
  fabbro review ../lib/utils.py`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			var content string
			var sourceFile string
			var err error

			if stdinFlag && len(args) == 1 {
				return fmt.Errorf("cannot use both --stdin and a file path")
			}

			if stdinFlag {
				limitedReader := io.LimitReader(stdin, maxInputBytes+1)
				data, err := io.ReadAll(limitedReader)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				if len(data) > maxInputBytes {
					return fmt.Errorf("input too large: exceeds %d bytes", maxInputBytes)
				}
				content = string(data)
			} else if len(args) == 1 {
				sourceFile = args[0]
				info, err := os.Stat(sourceFile)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("file not found: %s", sourceFile)
					}
					return fmt.Errorf("failed to stat file: %w", err)
				}
				if info.Size() > maxInputBytes {
					return fmt.Errorf("file too large: %s exceeds %d bytes", sourceFile, maxInputBytes)
				}
				data, err := os.ReadFile(sourceFile)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				content = string(data)
			} else {
				return fmt.Errorf("no input provided. Use --stdin or provide a file path")
			}

			sess, err := session.Create(content, sourceFile)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			if jsonFlag {
				fmt.Fprintf(stdout, `{"sessionId":"%s"}`+"\n", sess.ID)
			} else {
				fmt.Fprintf(stdout, "Created session: %s\n", sess.ID)
			}

			model := tui.NewWithFile(sess, sourceFile)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read content from stdin")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output session ID as JSON")
	return cmd
}

func buildApplyCmd(stdout io.Writer) *cobra.Command {
	var jsonFlag bool
	var fileFlag string
	cmd := &cobra.Command{
		Use:   "apply [session-id]",
		Short: "Apply annotations from a session",
		Long: `Extract and display annotations from a review session.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - Provide either a session ID or use --file to find a session by source file.

Post-conditions:
  - Annotations are parsed from the session content.
  - Output is printed to stdout (human-readable or JSON with --json).`,
		Example: `  # Apply annotations from a specific session
  fabbro apply abc123

  # Find and apply session by source file
  fabbro apply --file main.go

  # Get annotations as JSON for programmatic use
  fabbro apply abc123 --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			// Validate mutual exclusivity
			if fileFlag != "" && len(args) == 1 {
				return fmt.Errorf("cannot use both session-id and --file")
			}
			if fileFlag == "" && len(args) == 0 {
				return fmt.Errorf("must provide either session-id or --file")
			}

			var sess *session.Session
			var err error

			if fileFlag != "" {
				sess, err = session.FindBySourceFile(fileFlag)
				if err != nil {
					return fmt.Errorf("failed to find session for file %q: %w", fileFlag, err)
				}
			} else {
				sessionID := args[0]
				sess, err = session.Load(sessionID)
				if err != nil {
					return fmt.Errorf("failed to load session %q: %w", sessionID, err)
				}
			}

			annotations, _, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse FEM in session %q: %w", sess.ID, err)
			}

			if jsonFlag {
				output := struct {
					SessionID   string           `json:"sessionId"`
					SourceFile  string           `json:"sourceFile"`
					Annotations []fem.Annotation `json:"annotations"`
				}{
					SessionID:   sess.ID,
					SourceFile:  sess.SourceFile,
					Annotations: annotations,
				}

				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			fmt.Fprintf(stdout, "Session: %s\n", sess.ID)
			if sess.SourceFile != "" {
				fmt.Fprintf(stdout, "Source: %s\n", sess.SourceFile)
			}
			fmt.Fprintf(stdout, "Annotations: %d\n", len(annotations))
			for _, a := range annotations {
				if a.StartLine == a.EndLine {
					fmt.Fprintf(stdout, "  Line %d: [%s] %s\n", a.StartLine, a.Type, a.Text)
				} else {
					fmt.Fprintf(stdout, "  Lines %d-%d: [%s] %s\n", a.StartLine, a.EndLine, a.Type, a.Text)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&fileFlag, "file", "", "Find session by source file path")
	return cmd
}

func buildSessionCmd(stdout io.Writer) *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Manage editing sessions",
		Long: `Manage fabbro editing sessions.

Sessions store your annotation work and can be listed, resumed, or queried.
Each session is identified by a unique ID and stored in .fabbro/sessions/.`,
	}

	sessionCmd.AddCommand(buildSessionListCmd(stdout))
	sessionCmd.AddCommand(buildSessionResumeCmd(stdout))
	return sessionCmd
}

func buildSessionListCmd(stdout io.Writer) *cobra.Command {
	var jsonFlag bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all editing sessions",
		Long: `List all fabbro editing sessions stored in the current directory.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).

Post-conditions:
  - All sessions are listed with their ID, creation date, and source file.
  - Use --json for machine-readable output.`,
		Example: `  # List all sessions
  fabbro session list

  # List sessions as JSON for scripting
  fabbro session list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if jsonFlag {
				type sessionOutput struct {
					ID         string `json:"id"`
					CreatedAt  string `json:"createdAt"`
					SourceFile string `json:"sourceFile,omitempty"`
				}
				output := make([]sessionOutput, len(sessions))
				for i, s := range sessions {
					output[i] = sessionOutput{
						ID:         s.ID,
						CreatedAt:  s.CreatedAt.Format("2006-01-02 15:04:05"),
						SourceFile: s.SourceFile,
					}
				}
				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(stdout, "No sessions found.")
				return nil
			}

			for _, s := range sessions {
				date := s.CreatedAt.Format("2006-01-02 15:04")
				if s.SourceFile != "" {
					fmt.Fprintf(stdout, "%s  %s  %s\n", s.ID, date, s.SourceFile)
				} else {
					fmt.Fprintf(stdout, "%s  %s  (stdin)\n", s.ID, date)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}

func buildSessionResumeCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "resume <session-id>",
		Short: "Resume a previous editing session",
		Long: `Resume a previous fabbro editing session by its ID.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - The session ID must exist (use 'fabbro session list' to find IDs).

Post-conditions:
  - The session is loaded with its existing annotations.
  - The TUI opens for continued annotation work.`,
		Example: `  # Resume a session by ID
  fabbro session resume abc123

  # Find a session ID first, then resume
  fabbro session list
  fabbro session resume <id-from-list>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]
			sess, err := session.Load(sessionID)
			if err != nil {
				return fmt.Errorf("failed to load session %q: %w", sessionID, err)
			}

			annotations, cleanContent, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse session content: %w", err)
			}

			sess.Content = cleanContent

			fmt.Fprintf(stdout, "Resuming session: %s\n", sess.ID)

			model := tui.NewWithAnnotations(sess, sess.SourceFile, annotations)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
}

func buildTutorCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "tutor",
		Short: "Start the interactive tutorial",
		Long: `Launch an interactive tutorial that teaches fabbro basics.

The tutor opens a guided lesson file in the TUI where you can
practice navigation, selection, and annotation. Like vimtutor,
this is hands-on learning.

Your practice session is temporary and won't be saved.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sess := &session.Session{
				ID:        tutor.SessionID,
				Content:   tutor.Content,
				CreatedAt: time.Now(),
			}

			fmt.Fprintln(stdout, "Welcome to the fabbro tutor!")
			fmt.Fprintln(stdout, "")

			model := tui.NewWithFile(sess, "(tutorial)")
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
}
