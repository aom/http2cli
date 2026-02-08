package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30_000_000_000 // 30s
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 300_000_000_000 // 300s
	}
	if cfg.Security.CommandTimeout == 0 {
		cfg.Security.CommandTimeout = 30_000_000_000 // 30s
	}
	if cfg.Security.MaxConcurrentCommands == 0 {
		cfg.Security.MaxConcurrentCommands = 10
	}

	for i := range cfg.Tools {
		if cfg.Tools[i].Timeout == 0 {
			cfg.Tools[i].Timeout = cfg.Security.CommandTimeout
		}
		if cfg.Tools[i].Method == "" {
			cfg.Tools[i].Method = "POST"
		}
		if cfg.Tools[i].Input.Type == "" {
			cfg.Tools[i].Input.Type = "none"
		}
		if cfg.Tools[i].Output.Type == "" {
			cfg.Tools[i].Output.Type = "stdout"
		}
		if cfg.Tools[i].Output.ContentType == "" {
			cfg.Tools[i].Output.ContentType = "application/octet-stream"
		}
	}
}

func validate(cfg *Config) error {
	if len(cfg.Tools) == 0 {
		return fmt.Errorf("no tools configured")
	}

	allowedCommands := make(map[string]bool)
	for _, cmd := range cfg.Security.AllowedCommands {
		allowedCommands[cmd] = true
	}

	endpoints := make(map[string]bool)
	for _, tool := range cfg.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tool missing name")
		}
		if tool.Endpoint == "" {
			return fmt.Errorf("tool %q missing endpoint", tool.Name)
		}
		if tool.Command == "" {
			return fmt.Errorf("tool %q missing command", tool.Name)
		}

		// Check command is in allowlist
		if !allowedCommands[tool.Command] {
			return fmt.Errorf("tool %q uses command %q which is not in allowed_commands", tool.Name, tool.Command)
		}

		// Check for duplicate endpoints
		key := tool.Method + " " + tool.Endpoint
		if endpoints[key] {
			return fmt.Errorf("duplicate endpoint: %s %s", tool.Method, tool.Endpoint)
		}
		endpoints[key] = true

		// Validate input type
		switch tool.Input.Type {
		case "stdin", "file", "none":
			// valid
		default:
			return fmt.Errorf("tool %q has invalid input type %q", tool.Name, tool.Input.Type)
		}

		// Validate output type
		switch tool.Output.Type {
		case "stdout", "file":
			// valid
		default:
			return fmt.Errorf("tool %q has invalid output type %q", tool.Name, tool.Output.Type)
		}
	}

	return nil
}
