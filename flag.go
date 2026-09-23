package varg

import (
	"strconv"
	"strings"
)

// Type represents the type of a flag.
type Type int

const (
	TypeString Type = iota
	TypeInt
	TypeBool
	TypeFloat64
	TypeStringSlice
)

// Flag represents a single command-line flag.
type Flag struct {
	Key       string
	Short     string
	Default   interface{}
	Help      string
	Type      Type
	EnvVar    string
	envMatch  []string
}

// Env sets the environment variable name for this flag.
// Returns the flag for chaining.
func (f *Flag) Env(envVar string) *Flag {
	f.EnvVar = envVar
	f.envMatch = append(f.envMatch, envVar)
	return f
}

// EnvWithPrefix sets the environment variable with a custom prefix.
// For example, if key is "server.host" and prefix is "MYAPP",
// the env var will be "MYAPP_SERVER__HOST".
func (f *Flag) EnvWithPrefix(prefix string) *Flag {
	if prefix == "" {
		return f
	}

	// Replace dots with double underscores for nested keys
	envKey := strings.ToUpper(strings.ReplaceAll(f.Key, ".", "__"))

	f.EnvVar = prefix + "_" + envKey
	f.envMatch = append(f.envMatch, f.EnvVar)
	return f
}

// convertValue converts a string to the appropriate type.
func convertValue(s string, typ Type) (interface{}, error) {
	switch typ {
	case TypeString:
		return s, nil
	case TypeInt:
		return strconv.Atoi(s)
	case TypeBool:
		switch strings.ToLower(s) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return false, strconv.ErrSyntax
		}
	case TypeFloat64:
		return strconv.ParseFloat(s, 64)
	case TypeStringSlice:
		return []string{s}, nil
	default:
		return nil, strconv.ErrSyntax
	}
}
