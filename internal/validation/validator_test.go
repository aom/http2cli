package validation

import (
	"testing"

	"http2cli/internal/config"
)

func intPtr(i int) *int {
	return &i
}

func TestValidateBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", true, false},
		{"1", true, false},
		{"yes", true, false},
		{"on", true, false},
		{"false", false, false},
		{"0", false, false},
		{"no", false, false},
		{"off", false, false},
		{"", false, false},
		{"invalid", false, true},
		{"TRUE", false, true}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := validateBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("validateBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateInt(t *testing.T) {
	tests := []struct {
		name     string
		arg      *config.MappedArgument
		value    string
		expected int
		wantErr  bool
	}{
		{
			name:     "valid integer",
			arg:      &config.MappedArgument{Name: "test"},
			value:    "42",
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "negative integer",
			arg:      &config.MappedArgument{Name: "test"},
			value:    "-10",
			expected: -10,
			wantErr:  false,
		},
		{
			name:    "invalid integer",
			arg:     &config.MappedArgument{Name: "test"},
			value:   "abc",
			wantErr: true,
		},
		{
			name:     "within bounds",
			arg:      &config.MappedArgument{Name: "test", Min: intPtr(1), Max: intPtr(100)},
			value:    "50",
			expected: 50,
			wantErr:  false,
		},
		{
			name:    "below minimum",
			arg:     &config.MappedArgument{Name: "test", Min: intPtr(10)},
			value:   "5",
			wantErr: true,
		},
		{
			name:    "above maximum",
			arg:     &config.MappedArgument{Name: "test", Max: intPtr(10)},
			value:   "15",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateInt(tt.arg, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("validateInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	arg := &config.MappedArgument{
		Name:   "size",
		Values: []string{"A4", "Letter", "Legal"},
	}

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"A4", false},
		{"Letter", false},
		{"Legal", false},
		{"a4", true},      // case sensitive
		{"A5", true},      // not in list
		{"", true},        // empty
		{"Invalid", true}, // not in list
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result, err := validateEnum(arg, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEnum(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.value {
				t.Errorf("validateEnum(%q) = %v, want %v", tt.value, result, tt.value)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	tests := []struct {
		name    string
		arg     *config.MappedArgument
		value   string
		wantErr bool
	}{
		{
			name:    "no pattern",
			arg:     &config.MappedArgument{Name: "test"},
			value:   "anything",
			wantErr: false,
		},
		{
			name:    "pattern matches",
			arg:     &config.MappedArgument{Name: "test", Pattern: "^[0-9]+$"},
			value:   "12345",
			wantErr: false,
		},
		{
			name:    "pattern does not match",
			arg:     &config.MappedArgument{Name: "test", Pattern: "^[0-9]+$"},
			value:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid regex pattern",
			arg:     &config.MappedArgument{Name: "test", Pattern: "[invalid"},
			value:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateString(tt.arg, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndParse(t *testing.T) {
	tests := []struct {
		name     string
		arg      *config.MappedArgument
		value    string
		expected any
		wantErr  bool
	}{
		{
			name:     "required missing",
			arg:      &config.MappedArgument{Name: "test", Required: true},
			value:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "optional missing returns nil",
			arg:      &config.MappedArgument{Name: "test", Required: false},
			value:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "uses default when empty",
			arg:      &config.MappedArgument{Name: "test", Type: "string", Default: "default_value"},
			value:    "",
			expected: "default_value",
			wantErr:  false,
		},
		{
			name:     "bool type",
			arg:      &config.MappedArgument{Name: "test", Type: "bool"},
			value:    "true",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "int type",
			arg:      &config.MappedArgument{Name: "test", Type: "int"},
			value:    "42",
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "enum type",
			arg:      &config.MappedArgument{Name: "test", Type: "enum", Values: []string{"a", "b"}},
			value:    "a",
			expected: "a",
			wantErr:  false,
		},
		{
			name:     "unknown type returns value as-is",
			arg:      &config.MappedArgument{Name: "test", Type: "unknown"},
			value:    "anything",
			expected: "anything",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndParse(tt.arg, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndParse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ValidateAndParse() = %v, want %v", result, tt.expected)
			}
		})
	}
}
