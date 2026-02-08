package config

import (
	"testing"
	"time"
)

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{}
	applyDefaults(cfg)

	// Server defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("expected read timeout 30s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 300*time.Second {
		t.Errorf("expected write timeout 300s, got %v", cfg.Server.WriteTimeout)
	}

	// Security defaults
	if cfg.Security.CommandTimeout != 30*time.Second {
		t.Errorf("expected command timeout 30s, got %v", cfg.Security.CommandTimeout)
	}
	if cfg.Security.MaxConcurrentCommands != 10 {
		t.Errorf("expected max concurrent 10, got %d", cfg.Security.MaxConcurrentCommands)
	}
}

func TestApplyDefaultsPreservesValues(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 9000,
		},
	}
	applyDefaults(cfg)

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host to be preserved, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected port to be preserved, got %d", cfg.Server.Port)
	}
}

func TestApplyDefaultsToolDefaults(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			CommandTimeout: 60 * time.Second,
		},
		Tools: []ToolConfig{
			{Name: "test"},
		},
	}
	applyDefaults(cfg)

	tool := cfg.Tools[0]
	if tool.Method != "POST" {
		t.Errorf("expected method POST, got %s", tool.Method)
	}
	if tool.Input.Type != "none" {
		t.Errorf("expected input type none, got %s", tool.Input.Type)
	}
	if tool.Output.Type != "stdout" {
		t.Errorf("expected output type stdout, got %s", tool.Output.Type)
	}
	if tool.Output.ContentType != "application/octet-stream" {
		t.Errorf("expected content-type application/octet-stream, got %s", tool.Output.ContentType)
	}
	if tool.Timeout != 60*time.Second {
		t.Errorf("expected tool timeout to inherit from security, got %v", tool.Timeout)
	}
}

func TestValidateNoTools(t *testing.T) {
	cfg := &Config{}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for no tools configured")
	}
}

func TestValidateMissingName(t *testing.T) {
	cfg := &Config{
		Tools: []ToolConfig{
			{Endpoint: "/test", Command: "/bin/test"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for missing tool name")
	}
}

func TestValidateMissingEndpoint(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/test"},
		},
		Tools: []ToolConfig{
			{Name: "test", Command: "/bin/test"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for missing endpoint")
	}
}

func TestValidateMissingCommand(t *testing.T) {
	cfg := &Config{
		Tools: []ToolConfig{
			{Name: "test", Endpoint: "/test"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestValidateCommandNotAllowed(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/allowed"},
		},
		Tools: []ToolConfig{
			{Name: "test", Endpoint: "/test", Command: "/bin/notallowed"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for command not in allowlist")
	}
}

func TestValidateDuplicateEndpoints(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/test"},
		},
		Tools: []ToolConfig{
			{Name: "test1", Endpoint: "/test", Command: "/bin/test", Method: "POST"},
			{Name: "test2", Endpoint: "/test", Command: "/bin/test", Method: "POST"},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for duplicate endpoints")
	}
}

func TestValidateSameEndpointDifferentMethods(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/test"},
		},
		Tools: []ToolConfig{
			{Name: "test1", Endpoint: "/test", Command: "/bin/test", Method: "GET", Input: InputConfig{Type: "none"}, Output: OutputConfig{Type: "stdout"}},
			{Name: "test2", Endpoint: "/test", Command: "/bin/test", Method: "POST", Input: InputConfig{Type: "none"}, Output: OutputConfig{Type: "stdout"}},
		},
	}
	err := validate(cfg)
	if err != nil {
		t.Errorf("same endpoint with different methods should be allowed: %v", err)
	}
}

func TestValidateInvalidInputType(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/test"},
		},
		Tools: []ToolConfig{
			{Name: "test", Endpoint: "/test", Command: "/bin/test", Input: InputConfig{Type: "invalid"}},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for invalid input type")
	}
}

func TestValidateInvalidOutputType(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/test"},
		},
		Tools: []ToolConfig{
			{Name: "test", Endpoint: "/test", Command: "/bin/test", Input: InputConfig{Type: "none"}, Output: OutputConfig{Type: "invalid"}},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for invalid output type")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := &Config{
		Security: SecurityConfig{
			AllowedCommands: []string{"/bin/cat", "/bin/date"},
		},
		Tools: []ToolConfig{
			{
				Name:     "cat",
				Endpoint: "/cat",
				Command:  "/bin/cat",
				Method:   "POST",
				Input:    InputConfig{Type: "stdin"},
				Output:   OutputConfig{Type: "stdout"},
			},
			{
				Name:     "date",
				Endpoint: "/date",
				Command:  "/bin/date",
				Method:   "GET",
				Input:    InputConfig{Type: "none"},
				Output:   OutputConfig{Type: "stdout"},
			},
		},
	}
	err := validate(cfg)
	if err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}
