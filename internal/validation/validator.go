package validation

import (
	"fmt"
	"regexp"
	"strconv"

	"http2cli/internal/config"
)

// ValidateAndParse validates an argument value and returns the parsed result.
func ValidateAndParse(arg *config.MappedArgument, value string) (any, error) {
	// Handle empty value
	if value == "" {
		if arg.Required {
			return nil, fmt.Errorf("required parameter %q is missing", arg.Name)
		}
		if arg.Default != "" {
			value = arg.Default
		} else {
			return nil, nil // No value, no default, not required
		}
	}

	switch arg.Type {
	case "string":
		return validateString(arg, value)
	case "int":
		return validateInt(arg, value)
	case "bool":
		return validateBool(value)
	case "enum":
		return validateEnum(arg, value)
	default:
		return value, nil
	}
}

func validateString(arg *config.MappedArgument, value string) (string, error) {
	if arg.Pattern != "" {
		re, err := regexp.Compile(arg.Pattern)
		if err != nil {
			return "", fmt.Errorf("invalid pattern for %q: %w", arg.Name, err)
		}
		if !re.MatchString(value) {
			return "", fmt.Errorf("value %q does not match pattern for %q", value, arg.Name)
		}
	}
	return value, nil
}

func validateInt(arg *config.MappedArgument, value string) (int, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %q: %s", arg.Name, value)
	}

	if arg.Min != nil && i < *arg.Min {
		return 0, fmt.Errorf("value for %q must be >= %d", arg.Name, *arg.Min)
	}
	if arg.Max != nil && i > *arg.Max {
		return 0, fmt.Errorf("value for %q must be <= %d", arg.Name, *arg.Max)
	}

	return i, nil
}

func validateBool(value string) (bool, error) {
	switch value {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

func validateEnum(arg *config.MappedArgument, value string) (string, error) {
	for _, v := range arg.Values {
		if v == value {
			return value, nil
		}
	}
	return "", fmt.Errorf("value %q for %q must be one of: %v", value, arg.Name, arg.Values)
}
