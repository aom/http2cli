package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"http2cli/internal/config"
	"http2cli/internal/executor"
	"http2cli/internal/validation"
)

// ToolHandler handles HTTP requests for a specific tool.
type ToolHandler struct {
	tool     *config.ToolConfig
	executor *executor.Executor
}

// NewToolHandler creates a new handler for a tool.
func NewToolHandler(tool *config.ToolConfig, exec *executor.Executor) *ToolHandler {
	return &ToolHandler{
		tool:     tool,
		executor: exec,
	}
}

// ServeHTTP handles the HTTP request.
func (h *ToolHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != h.tool.Method {
		sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("method %s not allowed", r.Method))
		return
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.tool.Timeout)
	defer cancel()

	// Parse and validate arguments
	args, err := h.buildArguments(r)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Handle input
	var stdin io.Reader
	if h.tool.Input.Type == "stdin" {
		stdin, err = h.getInputReader(r)
		if err != nil {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Execute command
	start := time.Now()
	output, err := h.executor.Execute(ctx, h.tool.Command, args, stdin)
	duration := time.Since(start)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendError(w, http.StatusGatewayTimeout, "command execution timed out")
			return
		}
		if cmdErr, ok := err.(*executor.CommandError); ok {
			sendError(w, http.StatusUnprocessableEntity, cmdErr.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Send response
	w.Header().Set("Content-Type", h.tool.Output.ContentType)
	w.Header().Set("X-Processing-Time", duration.String())
	if h.tool.Output.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", h.tool.Output.Filename))
	}
	w.Write(output)
}

func (h *ToolHandler) buildArguments(r *http.Request) ([]string, error) {
	var args []string

	// Add static arguments
	args = append(args, h.tool.Arguments.Static...)

	// Process mapped arguments
	for _, mapping := range h.tool.Arguments.Mapped {
		var value string

		switch mapping.Source {
		case "query":
			value = r.URL.Query().Get(mapping.Name)
		case "form":
			value = r.FormValue(mapping.Name)
		case "header":
			value = r.Header.Get(mapping.Name)
		default:
			value = r.URL.Query().Get(mapping.Name) // Default to query
		}

		// Use default if empty
		if value == "" && mapping.Default != "" {
			value = mapping.Default
		}

		// Skip if still empty and not required
		if value == "" && !mapping.Required {
			continue
		}

		// Validate
		parsed, err := validation.ValidateAndParse(&mapping, value)
		if err != nil {
			return nil, err
		}

		if parsed == nil {
			continue
		}

		// Add to args based on type
		switch v := parsed.(type) {
		case bool:
			if v {
				args = append(args, mapping.Flag)
			}
		case string:
			if strings.HasPrefix(mapping.Flag, "+") {
				// Special case for date-style format (e.g., +%Y-%m-%d)
				args = append(args, mapping.Flag+v)
			} else {
				args = append(args, mapping.Flag, v)
			}
		case int:
			args = append(args, mapping.Flag, fmt.Sprintf("%d", v))
		default:
			args = append(args, mapping.Flag, fmt.Sprintf("%v", v))
		}
	}

	return args, nil
}

func (h *ToolHandler) getInputReader(r *http.Request) (io.Reader, error) {
	if h.tool.Input.FormField == "" {
		return nil, fmt.Errorf("input form_field not configured")
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB memory limit
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	file, _, err := r.FormFile(h.tool.Input.FormField)
	if err != nil {
		if h.tool.Input.Required {
			return nil, fmt.Errorf("required file %q not provided", h.tool.Input.FormField)
		}
		return nil, nil
	}
	defer file.Close()

	// Read file into buffer (needed because file will be closed after this function)
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	return &buf, nil
}

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// HealthHandler returns a simple health check.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ToolsHandler returns the list of available tools.
func ToolsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type toolInfo struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Endpoint    string `json:"endpoint"`
			Method      string `json:"method"`
		}

		tools := make([]toolInfo, 0, len(cfg.Tools))
		for _, t := range cfg.Tools {
			tools = append(tools, toolInfo{
				Name:        t.Name,
				Description: t.Description,
				Endpoint:    t.Endpoint,
				Method:      t.Method,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"tools": tools})
	}
}

// ReadyHandler checks if all configured tools are available.
func ReadyHandler(cfg *config.Config, exec *executor.Executor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errors []string
		for _, tool := range cfg.Tools {
			if !exec.IsAllowed(tool.Command) {
				errors = append(errors, fmt.Sprintf("command %q not allowed", tool.Command))
			}
		}

		if len(errors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "not ready",
				"errors": errors,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// APIKeyMiddleware validates X-API-Key header if configured.
func APIKeyMiddleware(apiKey string, next http.Handler) http.Handler {
	if apiKey == "" {
		return next // No API key configured, skip auth
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/ready" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")
		if key != apiKey {
			sendError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}
