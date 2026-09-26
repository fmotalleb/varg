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
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
)

// isNumeric reports whether the type supports the valueless forms of a flag:
// counting shorthand (-vvv) and incrementing (-v).
func (t Type) isNumeric() bool {
	switch t {
	case TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64,
		TypeFloat32, TypeFloat64:
		return true
	default:
		return false
	}
}

// isUnsigned reports whether the type holds unsigned integers.
func (t Type) isUnsigned() bool {
	switch t {
	case TypeUint, TypeUint8, TypeUint16, TypeUint32, TypeUint64:
		return true
	default:
		return false
	}
}

// Flag represents a single command-line flag.
type Flag struct {
	Key      string
	Short    string
	Default  interface{}
	Help     string
	Type     Type
	EnvVar   string
	envMatch []string
}

// Env sets the environment variable name for this flag.
// Returns the flag for chaining.
func (f *Flag) Env(envVar string) *Flag {
	f.EnvVar = envVar
	f.envMatch = append(f.envMatch, envVar)
	return f
}

// helpText returns the help text of the flag with its environment variable
// appended, e.g. "output file [$OUTPUT_FILE]".
func (f *Flag) helpText() string {
	if f.EnvVar == "" {
		return f.Help
	}
	if f.Help == "" {
		return "[$" + f.EnvVar + "]"
	}
	return f.Help + " [$" + f.EnvVar + "]"
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
	case TypeBool:
		switch strings.ToLower(s) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return false, strconv.ErrSyntax
		}
	case TypeStringSlice:
		return []string{s}, nil
	case TypeInt:
		return strconv.Atoi(s)
	case TypeInt8:
		v, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return nil, err
		}
		return int8(v), nil
	case TypeInt16:
		v, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return nil, err
		}
		return int16(v), nil
	case TypeInt32:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(v), nil
	case TypeInt64:
		return strconv.ParseInt(s, 10, 64)
	case TypeUint:
		v, err := strconv.ParseUint(s, 10, strconv.IntSize)
		if err != nil {
			return nil, err
		}
		return uint(v), nil
	case TypeUint8:
		v, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return nil, err
		}
		return uint8(v), nil
	case TypeUint16:
		v, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return nil, err
		}
		return uint16(v), nil
	case TypeUint32:
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return uint32(v), nil
	case TypeUint64:
		return strconv.ParseUint(s, 10, 64)
	case TypeFloat64:
		return strconv.ParseFloat(s, 64)
	case TypeFloat32:
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return nil, err
		}
		return float32(v), nil
	default:
		return nil, strconv.ErrSyntax
	}
}
