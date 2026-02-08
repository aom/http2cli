package config

import "time"

// Config is the root configuration structure.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Security SecurityConfig `yaml:"security"`
	Tools    []ToolConfig   `yaml:"tools"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Host          string        `yaml:"host"`
	Port          int           `yaml:"port"`
	ReadTimeout   time.Duration `yaml:"read_timeout"`
	WriteTimeout  time.Duration `yaml:"write_timeout"`
	MaxUploadSize string        `yaml:"max_upload_size"`
}

// SecurityConfig contains security-related settings.
type SecurityConfig struct {
	APIKey                string        `yaml:"api_key"`
	AllowedCommands       []string      `yaml:"allowed_commands"`
	CommandTimeout        time.Duration `yaml:"command_timeout"`
	MaxConcurrentCommands int           `yaml:"max_concurrent_commands"`
}

// ToolConfig defines a CLI tool exposed via HTTP.
type ToolConfig struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Endpoint    string          `yaml:"endpoint"`
	Method      string          `yaml:"method"`
	Command     string          `yaml:"command"`
	Timeout     time.Duration   `yaml:"timeout,omitempty"`
	Input       InputConfig     `yaml:"input"`
	Output      OutputConfig    `yaml:"output"`
	Arguments   ArgumentsConfig `yaml:"arguments,omitempty"`
}

// InputConfig defines how input is provided to the command.
type InputConfig struct {
	Type              string   `yaml:"type"` // stdin, file, none
	FormField         string   `yaml:"form_field,omitempty"`
	Required          bool     `yaml:"required,omitempty"`
	AllowedExtensions []string `yaml:"allowed_extensions,omitempty"`
	MaxSize           string   `yaml:"max_size,omitempty"`
}

// OutputConfig defines how output is returned.
type OutputConfig struct {
	Type        string `yaml:"type"` // stdout, file
	ContentType string `yaml:"content_type"`
	Filename    string `yaml:"filename,omitempty"`
}

// ArgumentsConfig defines how CLI arguments are constructed.
type ArgumentsConfig struct {
	Static []string        `yaml:"static,omitempty"`
	Mapped []MappedArgument `yaml:"mapped,omitempty"`
}

// MappedArgument maps an HTTP parameter to a CLI flag.
type MappedArgument struct {
	Name        string   `yaml:"name"`
	Flag        string   `yaml:"flag"`
	Source      string   `yaml:"source"` // query, form, header
	Type        string   `yaml:"type"`   // string, int, bool, enum
	Values      []string `yaml:"values,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Min         *int     `yaml:"min,omitempty"`
	Max         *int     `yaml:"max,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
	Description string   `yaml:"description,omitempty"`
}
