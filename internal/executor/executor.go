package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// CommandError represents a command execution failure.
type CommandError struct {
	ExitCode int
	Stderr   string
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command failed with exit code %d: %s", e.ExitCode, e.Stderr)
}

// Executor runs CLI commands with security controls.
type Executor struct {
	allowedCommands map[string]bool
	semaphore       chan struct{}
}

// New creates a new Executor with the given security constraints.
func New(allowedCommands []string, maxConcurrent int) *Executor {
	allowed := make(map[string]bool)
	for _, cmd := range allowedCommands {
		allowed[cmd] = true
	}
	return &Executor{
		allowedCommands: allowed,
		semaphore:       make(chan struct{}, maxConcurrent),
	}
}

// Execute runs a command with the given arguments.
// stdin can be nil if no input is needed.
// Returns stdout content or an error.
func (e *Executor) Execute(ctx context.Context, command string, args []string, stdin io.Reader) ([]byte, error) {
	// Security: verify command is in allowlist
	if !e.allowedCommands[command] {
		return nil, fmt.Errorf("command not allowed: %s", command)
	}

	// Acquire semaphore slot (limit concurrent executions)
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// SECURITY: Use exec.CommandContext with separate arguments.
	// This prevents shell injection because:
	// 1. No shell interpreter is invoked
	// 2. Arguments are passed directly to the program
	// 3. Special characters are not interpreted
	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if stdin != nil {
		cmd.Stdin = stdin
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, &CommandError{
				ExitCode: exitErr.ExitCode(),
				Stderr:   stderr.String(),
			}
		}
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	return stdout.Bytes(), nil
}

// IsAllowed checks if a command is in the allowlist.
func (e *Executor) IsAllowed(command string) bool {
	return e.allowedCommands[command]
}
