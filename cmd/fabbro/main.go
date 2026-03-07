package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TUIRunner launches the TUI with the given model. Production code uses
// runTUI; tests inject a no-op to skip the interactive UI.
type TUIRunner func(model tea.Model) error

func runTUI(model tea.Model) error {
	_, err := tea.NewProgram(model).Run()
	return err
}

func main() {
	os.Exit(realMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, runTUI))
}

func realMain(args []string, stdin io.Reader, stdout, stderr io.Writer, tui TUIRunner) int {
	rootCmd := buildRootCmd(stdin, stdout, tui)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func buildRootCmd(stdin io.Reader, stdout io.Writer, tuiRun TUIRunner) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "fabbro",
		Short: "A code review annotation tool",
		Long: `"For you, il miglior fabbro"
  — after T.S. Eliot, The Waste Land

A code review annotation tool with a terminal UI.`,
		Version: version,
	}

	rootCmd.AddCommand(buildInitCmd(stdout))
	rootCmd.AddCommand(buildReviewCmd(stdin, stdout, tuiRun))
	rootCmd.AddCommand(buildApplyCmd(stdout))
	rootCmd.AddCommand(buildSessionCmd(stdin, stdout, tuiRun))
	rootCmd.AddCommand(buildCompletionCmd())
	rootCmd.AddCommand(buildTutorCmd(stdout, tuiRun))
	rootCmd.AddCommand(buildPrimeCmd(stdout))

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
				return fmt.Errorf("unsupported shell: %s. Supported shells: bash, zsh, fish, powershell", args[0])
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

func buildReviewCmd(stdin io.Reader, stdout io.Writer, tuiRun TUIRunner) *cobra.Command {
	var stdinFlag bool
	var jsonFlag bool
	var idFlag string
	var editorFlag bool
	var noInteractiveFlag bool
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
						return fmt.Errorf("file not found: %s. Check the path and try again", sourceFile)
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
				return fmt.Errorf("no input file specified. Provide a file path as an argument or pipe content via --stdin")
			}

			var sess *session.Session
			if idFlag != "" {
				sess, err = session.CreateWithID(idFlag, content, sourceFile)
			} else {
				sess, err = session.Create(content, sourceFile)
			}
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			if noInteractiveFlag {
				fmt.Fprintln(stdout, sess.ID)
				return nil
			}

			if jsonFlag {
				json.NewEncoder(stdout).Encode(map[string]string{"sessionId": sess.ID})
			} else {
				fmt.Fprintf(stdout, "Created session: %s\n", sess.ID)
			}

			if editorFlag {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = os.Getenv("VISUAL")
				}
				if editor == "" {
					return fmt.Errorf("no editor configured. Set $EDITOR or $VISUAL")
				}

				sessionsDir, sdErr := config.GetSessionsDir()
				if sdErr != nil {
					return fmt.Errorf("failed to find sessions directory: %w", sdErr)
				}
				sessionPath := filepath.Join(sessionsDir, sess.ID+".fem")

				editorCmd := exec.Command(editor, sessionPath)
				editorCmd.Stdin = os.Stdin
				editorCmd.Stdout = os.Stdout
				editorCmd.Stderr = os.Stderr
				return editorCmd.Run()
			}

			model := tui.NewWithFile(sess, sourceFile)
			if err := tuiRun(model); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&stdinFlag, "stdin", false, "Read content from stdin")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output session ID as JSON")
	cmd.Flags().StringVar(&idFlag, "id", "", "Custom session ID (alphanumeric, dash, underscore; max 64 chars)")
	cmd.Flags().BoolVar(&editorFlag, "editor", false, "Open in $EDITOR instead of TUI")
	cmd.Flags().BoolVar(&noInteractiveFlag, "no-interactive", false, "Create session without opening TUI or editor")
	return cmd
}

func buildApplyCmd(stdout io.Writer) *cobra.Command {
	var jsonFlag bool
	var compactFlag bool
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
				return fmt.Errorf("no session specified. Provide a session ID as an argument or use --file to find by source file. Run 'fabbro session list' to see available sessions")
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
				sess, err = session.LoadPartial(sessionID)
				if err != nil {
					return fmt.Errorf("failed to load session %q: %w", sessionID, err)
				}
			}

			annotations, _, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse FEM in session %q: %w", sess.ID, err)
			}

			// Verify source file hash
			if valid, hashErr := sess.VerifySourceHash(); hashErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", hashErr)
			} else if !valid {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: source file has changed since session was created. Line numbers may have drifted.\n")
			}

			if jsonFlag {
				output := struct {
					SessionID   string           `json:"sessionId"`
					SourceFile  string           `json:"sourceFile"`
					CreatedAt   string           `json:"createdAt"`
					Annotations []fem.Annotation `json:"annotations"`
				}{
					SessionID:   sess.ID,
					SourceFile:  sess.SourceFile,
					CreatedAt:   sess.CreatedAt.Format(time.RFC3339),
					Annotations: annotations,
				}

				enc := json.NewEncoder(stdout)
				if !compactFlag {
					enc.SetIndent("", "  ")
				}
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
	cmd.Flags().BoolVar(&compactFlag, "compact", false, "Output minified JSON (use with --json)")
	cmd.Flags().StringVar(&fileFlag, "file", "", "Find session by source file path")
	return cmd
}

func buildSessionCmd(stdin io.Reader, stdout io.Writer, tuiRun TUIRunner) *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:   "session",
		Short: "Manage editing sessions",
		Long: `Manage fabbro editing sessions.

Sessions store your annotation work and can be listed, resumed, or queried.
Each session is identified by a unique ID and stored in .fabbro/sessions/.`,
	}

	sessionCmd.AddCommand(buildSessionListCmd(stdout))
	sessionCmd.AddCommand(buildSessionShowCmd(stdout))
	sessionCmd.AddCommand(buildSessionResumeCmd(stdout, tuiRun))
	sessionCmd.AddCommand(buildSessionDeleteCmd(stdin, stdout))
	sessionCmd.AddCommand(buildSessionCleanCmd(stdin, stdout))
	sessionCmd.AddCommand(buildSessionExportCmd(stdout))
	return sessionCmd
}

// parseDaysDuration parses a duration string like "7d", "14d", "30d".
func parseDaysDuration(s string) (time.Duration, error) {
	if !strings.HasSuffix(s, "d") {
		return 0, fmt.Errorf("invalid duration format: %s (use Nd, e.g. 7d)", s)
	}
	numStr := strings.TrimSuffix(s, "d")
	var days int
	if _, err := fmt.Sscanf(numStr, "%d", &days); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	if days < 0 {
		return 0, fmt.Errorf("duration must be positive: %s", s)
	}
	return time.Duration(days) * 24 * time.Hour, nil
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

			// Count annotations for each session
			type sessionInfo struct {
				session     *session.Session
				annotations int
			}
			infos := make([]sessionInfo, len(sessions))
			for i, s := range sessions {
				annotations, _, _ := fem.Parse(s.Content)
				infos[i] = sessionInfo{session: s, annotations: len(annotations)}
			}

			if jsonFlag {
				type sessionOutput struct {
					ID          string `json:"id"`
					CreatedAt   string `json:"createdAt"`
					SourceFile  string `json:"sourceFile,omitempty"`
					Annotations int    `json:"annotations"`
				}
				output := make([]sessionOutput, len(infos))
				for i, info := range infos {
					output[i] = sessionOutput{
						ID:          info.session.ID,
						CreatedAt:   info.session.CreatedAt.Format("2006-01-02 15:04:05"),
						SourceFile:  info.session.SourceFile,
						Annotations: info.annotations,
					}
				}
				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(stdout, "No sessions found.")
				fmt.Fprintln(stdout, "Start a review with: fabbro review <file>")
				return nil
			}

			for _, info := range infos {
				s := info.session
				date := s.CreatedAt.Format("2006-01-02 15:04")
				source := "(stdin)"
				if s.SourceFile != "" {
					source = s.SourceFile
				}
				fmt.Fprintf(stdout, "%s  %s  %-20s  %d annotations\n", s.ID, date, source, info.annotations)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}

func buildSessionShowCmd(stdout io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "show <session-id>",
		Short: "Show session details and annotation breakdown",
		Long: `Display detailed information about a review session.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - The session ID must exist (use 'fabbro session list' to find IDs).

Post-conditions:
  - Session metadata and annotation breakdown are printed to stdout.`,
		Example: `  # Show details for a session
  fabbro session show abc123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]
			sess, err := session.LoadPartial(sessionID)
			if err != nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			annotations, _, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse session content: %w", err)
			}

			source := "(stdin)"
			if sess.SourceFile != "" {
				source = sess.SourceFile
			}

			contentLines := len(strings.Split(sess.Content, "\n"))

			fmt.Fprintf(stdout, "Session ID:     %s\n", sess.ID)
			fmt.Fprintf(stdout, "Created:        %s\n", sess.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(stdout, "Source:         %s\n", source)
			fmt.Fprintf(stdout, "Content lines:  %d\n", contentLines)
			fmt.Fprintln(stdout)

			if len(annotations) == 0 {
				fmt.Fprintln(stdout, "No annotations.")
				return nil
			}

			breakdown := make(map[string]int)
			for _, a := range annotations {
				breakdown[string(a.Type)]++
			}

			fmt.Fprintf(stdout, "Annotations (%d total):\n", len(annotations))
			for typ, count := range breakdown {
				fmt.Fprintf(stdout, "  %s:  %d\n", typ, count)
			}

			return nil
		},
	}
}

func buildSessionDeleteCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
	var forceFlag bool
	cmd := &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a session",
		Long: `Delete a review session by its ID.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - The session ID must exist (use 'fabbro session list' to find IDs).

Post-conditions:
  - The session file is removed from .fabbro/sessions/.`,
		Example: `  # Delete with confirmation prompt
  fabbro session delete abc123

  # Delete without confirmation
  fabbro session delete abc123 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]

			// Verify session exists and resolve partial ID
			sess, err := session.LoadPartial(sessionID)
			if err != nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}
			sessionID = sess.ID

			if !forceFlag {
				fmt.Fprintf(stdout, "Delete session %s? [y/N] ", sessionID)
				var answer string
				fmt.Fscanln(stdin, &answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(stdout, "Aborted.")
					return nil
				}
			}

			if err := session.Delete(sessionID); err != nil {
				return fmt.Errorf("failed to delete session: %w", err)
			}

			fmt.Fprintf(stdout, "Deleted session: %s\n", sessionID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Delete without confirmation prompt")
	return cmd
}

func buildSessionExportCmd(stdout io.Writer) *cobra.Command {
	var outputFlag string
	cmd := &cobra.Command{
		Use:   "export <session-id>",
		Short: "Export session content",
		Long: `Export the content of a review session.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - The session ID must exist (use 'fabbro session list' to find IDs).

Post-conditions:
  - Session content is printed to stdout or written to a file.`,
		Example: `  # Export to stdout
  fabbro session export abc123

  # Export to a file
  fabbro session export abc123 --output review.fem`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]

			// Resolve partial ID
			sess, err := session.LoadPartial(sessionID)
			if err != nil {
				return fmt.Errorf("session not found: %s", sessionID)
			}

			sessionsDir, err := config.GetSessionsDir()
			if err != nil {
				return fmt.Errorf("failed to find sessions directory: %w", err)
			}
			sessionPath := filepath.Join(sessionsDir, sess.ID+".fem")

			data, err := os.ReadFile(sessionPath)
			if err != nil {
				return fmt.Errorf("failed to read session file: %w", err)
			}

			if outputFlag != "" {
				if err := os.WriteFile(outputFlag, data, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Fprintf(stdout, "Exported session %s to %s\n", sessionID, outputFlag)
				return nil
			}

			fmt.Fprint(stdout, string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&outputFlag, "output", "", "Write output to file instead of stdout")
	return cmd
}

func buildSessionCleanCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
	var olderThan string
	var dryRun bool
	var forceFlag bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove old sessions",
		Long: `Remove sessions older than a specified duration.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).

Post-conditions:
  - Sessions older than the threshold are deleted.
  - With --dry-run, only lists what would be deleted.`,
		Example: `  # Delete sessions older than 7 days (with confirmation)
  fabbro session clean --older-than 7d

  # Preview what would be deleted
  fabbro session clean --older-than 7d --dry-run

  # Delete without confirmation
  fabbro session clean --older-than 7d --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			duration, err := parseDaysDuration(olderThan)
			if err != nil {
				return err
			}
			if duration < 24*time.Hour && !forceFlag {
				return fmt.Errorf("minimum --older-than is 1d (safety limit). Use --force to override")
			}

			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			cutoff := time.Now().UTC().Add(-duration)
			var matches []*session.Session
			for _, s := range sessions {
				if s.CreatedAt.Before(cutoff) {
					matches = append(matches, s)
				}
			}

			if len(matches) == 0 {
				fmt.Fprintln(stdout, "No sessions older than "+olderThan+".")
				return nil
			}

			if dryRun {
				fmt.Fprintf(stdout, "Would delete %d session(s):\n", len(matches))
				for _, s := range matches {
					fmt.Fprintf(stdout, "  %s  %s\n", s.ID, s.CreatedAt.Format("2006-01-02 15:04"))
				}
				return nil
			}

			if !forceFlag {
				fmt.Fprintf(stdout, "Delete %d session(s) older than %s? [y/N] ", len(matches), olderThan)
				var answer string
				fmt.Fscanln(stdin, &answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(stdout, "Aborted.")
					return nil
				}
			}

			for _, s := range matches {
				if err := session.Delete(s.ID); err != nil {
					fmt.Fprintf(stdout, "  Failed to delete %s: %v\n", s.ID, err)
					continue
				}
				fmt.Fprintf(stdout, "  Deleted %s\n", s.ID)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&olderThan, "older-than", "", "Delete sessions older than this duration (e.g. 7d, 30d)")
	cmd.MarkFlagRequired("older-than")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "List sessions that would be deleted without deleting")
	cmd.Flags().BoolVar(&forceFlag, "force", false, "Skip confirmation and safety limit")
	return cmd
}

func buildSessionResumeCmd(stdout io.Writer, tuiRun TUIRunner) *cobra.Command {
	var editorFlag bool
	cmd := &cobra.Command{
		Use:   "resume <session-id>",
		Short: "Resume a previous editing session",
		Long: `Resume a previous fabbro editing session by its ID.

Pre-conditions:
  - fabbro must be initialized (run 'fabbro init' first).
  - The session ID must exist (use 'fabbro session list' to find IDs).

Post-conditions:
  - The session is loaded with its existing annotations.
  - The TUI opens for continued annotation work (or $EDITOR with --editor).`,
		Example: `  # Resume a session by ID
  fabbro session resume abc123

  # Resume in $EDITOR instead of TUI
  fabbro session resume abc123 --editor

  # Find a session ID first, then resume
  fabbro session list
  fabbro session resume <id-from-list>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.IsInitialized() {
				return fmt.Errorf("fabbro not initialized. Run 'fabbro init' first")
			}

			sessionID := args[0]
			sess, err := session.LoadPartial(sessionID)
			if err != nil {
				return fmt.Errorf("failed to load session %q: %w", sessionID, err)
			}

			if editorFlag {
				editor := os.Getenv("EDITOR")
				if editor == "" {
					editor = os.Getenv("VISUAL")
				}
				if editor == "" {
					return fmt.Errorf("no editor configured. Set $EDITOR or $VISUAL")
				}

				sessionsDir, err := config.GetSessionsDir()
				if err != nil {
					return fmt.Errorf("failed to find sessions directory: %w", err)
				}
				sessionPath := filepath.Join(sessionsDir, sess.ID+".fem")

				fmt.Fprintf(stdout, "Opening session %s in %s\n", sess.ID, editor)

				editorCmd := exec.Command(editor, sessionPath)
				editorCmd.Stdin = os.Stdin
				editorCmd.Stdout = os.Stdout
				editorCmd.Stderr = os.Stderr
				return editorCmd.Run()
			}

			annotations, cleanContent, err := fem.Parse(sess.Content)
			if err != nil {
				return fmt.Errorf("failed to parse session content: %w", err)
			}

			sess.Content = cleanContent

			fmt.Fprintf(stdout, "Resuming session: %s\n", sess.ID)

			model := tui.NewWithAnnotations(sess, sess.SourceFile, annotations)
			if err := tuiRun(model); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&editorFlag, "editor", false, "Open session in $EDITOR instead of TUI")
	return cmd
}

func buildTutorCmd(stdout io.Writer, tuiRun TUIRunner) *cobra.Command {
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
			if err := tuiRun(model); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			return nil
		},
	}
}

func buildPrimeCmd(stdout io.Writer) *cobra.Command {
	var jsonFlag bool
	cmd := &cobra.Command{
		Use:   "prime",
		Short: "Output AI-optimized workflow context for fabbro",
		Long: `Output a concise summary of fabbro for AI agents.

This command provides an AI-optimized overview of fabbro's purpose, 
key commands, and FEM syntax. Designed to quickly onboard AI coding
assistants to the fabbro workflow.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			primeInfo := PrimeInfo{
				Purpose: "fabbro is a local-first code review annotation tool with a terminal UI. It lets you annotate code using FEM (Fabbro Editing Markup) syntax, designed for human-AI review workflows.",
				Commands: []CommandInfo{
					{Name: "fabbro init", Description: "Initialize fabbro in current directory (creates .fabbro/)"},
					{Name: "fabbro review <file>", Description: "Start review session with file content"},
					{Name: "fabbro review --stdin", Description: "Start review session from stdin (e.g., git diff | fabbro review --stdin)"},
					{Name: "fabbro apply <session-id>", Description: "Show annotations from a session"},
					{Name: "fabbro apply <session-id> --json", Description: "Output annotations as JSON for programmatic use"},
					{Name: "fabbro apply --file <path>", Description: "Find and apply latest session for a source file"},
					{Name: "fabbro session list", Description: "List all editing sessions"},
					{Name: "fabbro session resume <id>", Description: "Resume a previous session in TUI"},
					{Name: "fabbro tutor", Description: "Interactive tutorial (like vimtutor)"},
				},
				FEMSyntax: []FEMInfo{
					{Syntax: "{>> text <<}", Type: "comment", Description: "General comment"},
					{Syntax: "{-- text --}", Type: "delete", Description: "Mark for deletion"},
					{Syntax: "{?? text ??}", Type: "question", Description: "Ask a question"},
					{Syntax: "{!! text !!}", Type: "expand", Description: "Request more detail"},
					{Syntax: "{== text ==}", Type: "keep", Description: "Mark as good/keep"},
					{Syntax: "{~~ text ~~}", Type: "unclear", Description: "Mark as unclear"},
					{Syntax: "{++ text ++}", Type: "change", Description: "Replacement text"},
				},
				TUIKeys: []KeyInfo{
					{Key: "j/k", Action: "Navigate up/down"},
					{Key: "v", Action: "Toggle line selection"},
					{Key: "c", Action: "Add comment annotation"},
					{Key: "Space", Action: "Open annotation palette"},
					{Key: "w", Action: "Save session"},
					{Key: "q", Action: "Quit"},
				},
				Docs: []string{
					"README.md - Overview and quick start",
					"docs/cli.md - Full CLI reference",
					"docs/tui.md - TUI keybindings",
					"docs/fem.md - FEM syntax reference",
				},
			}

			if jsonFlag {
				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(primeInfo)
			}

			fmt.Fprintln(stdout, "# fabbro — AI Workflow Context")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "## Purpose")
			fmt.Fprintln(stdout, primeInfo.Purpose)
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "## Commands")
			for _, c := range primeInfo.Commands {
				fmt.Fprintf(stdout, "  %s\n    %s\n", c.Name, c.Description)
			}
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "## FEM Syntax (annotations)")
			for _, f := range primeInfo.FEMSyntax {
				fmt.Fprintf(stdout, "  %s → %s (%s)\n", f.Syntax, f.Type, f.Description)
			}
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "## TUI Quick Reference")
			for _, k := range primeInfo.TUIKeys {
				fmt.Fprintf(stdout, "  %s: %s\n", k.Key, k.Action)
			}
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "## Documentation")
			for _, d := range primeInfo.Docs {
				fmt.Fprintf(stdout, "  %s\n", d)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	return cmd
}

type PrimeInfo struct {
	Purpose   string        `json:"purpose"`
	Commands  []CommandInfo `json:"commands"`
	FEMSyntax []FEMInfo     `json:"femSyntax"`
	TUIKeys   []KeyInfo     `json:"tuiKeys"`
	Docs      []string      `json:"docs"`
}

type CommandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FEMInfo struct {
	Syntax      string `json:"syntax"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type KeyInfo struct {
	Key    string `json:"key"`
	Action string `json:"action"`
}
