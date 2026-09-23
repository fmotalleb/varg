package varg

import (
	"fmt"
	"os"
	"strings"
)

// varg represents a set of command-line flags.
type varg struct {
	name         string
	flags        map[string]*Flag
	globalPrefix string
	parsedValues map[string]interface{}
}

// New creates a new varg with the given name.
func New(name string) *varg {
	return &varg{
		name:         name,
		flags:        make(map[string]*Flag),
		parsedValues: make(map[string]interface{}),
	}
}

// String adds a string flag to the set.
func (fs *varg) String(key, short, defaultVal, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeString)
}

// Int adds an int flag to the set.
func (fs *varg) Int(key, short string, defaultVal int, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt)
}

// Bool adds a bool flag to the set.
func (fs *varg) Bool(key, short string, defaultVal bool, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeBool)
}

// Float64 adds a float64 flag to the set.
func (fs *varg) Float64(key, short string, defaultVal float64, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeFloat64)
}

// StringSlice adds a string slice flag to the set (repeatable: -tag foo -tag bar).
func (fs *varg) StringSlice(key, short string, help string) *Flag {
	return fs.addFlag(key, short, []string{}, help, TypeStringSlice)
}

func (fs *varg) addFlag(key, short string, defaultVal interface{}, help string, typ Type) *Flag {
	flag := &Flag{
		Key:      key,
		Short:    short,
		Default:  defaultVal,
		Help:     help,
		Type:     typ,
		EnvVar:   "",
		envMatch: make([]string, 0),
	}
	fs.flags[key] = flag
	return flag
}

// GlobalEnvPrefix sets a prefix for all env vars (e.g., "MYAPP_").
// Individual flag env vars will be appended to this.
func (fs *varg) GlobalEnvPrefix(prefix string) *varg {
	fs.globalPrefix = prefix
	return fs
}

// Parse parses command-line arguments and environment variables.
// Returns a Config with resolved values following precedence: CLI > env > default.
func (fs *varg) Parse(args []string) (*Config, error) {
	fs.parsedValues = make(map[string]interface{})

	// First pass: collect all CLI arguments
	cliValues := make(map[string]interface{})
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			break
		}

		// Handle long form: --key=value or --key value
		if strings.HasPrefix(arg, "--") {
			key, val, ok := strings.Cut(arg[2:], "=")
			if !ok {
				// Check if next arg is the value
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					val = args[i+1]
					i++
				} else {
					// Boolean flag or flag needing value
					val = "true"
				}
			}

			flag, exists := fs.flags[key]
			if !exists {
				return nil, fmt.Errorf("unknown flag: --%s", key)
			}

			if err := fs.setFlagValue(cliValues, flag, key, val); err != nil {
				return nil, err
			}
			continue
		}

		// Handle short form: -k value or -kv
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]

			// Find flag by short form
			var flag *Flag
			var flagKey string
			for k, f := range fs.flags {
				if f.Short == short {
					flag = f
					flagKey = k
					break
				}
			}

			if flag == nil {
				return nil, fmt.Errorf("unknown flag: -%s", short)
			}

			var val string
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				val = args[i+1]
				i++
			} else {
				val = "true"
			}

			if err := fs.setFlagValue(cliValues, flag, flagKey, val); err != nil {
				return nil, err
			}
			continue
		}
	}

	// Second pass: resolve values from CLI, env, or default
	for key, flag := range fs.flags {
		var value interface{}
		var found bool

		// Priority 1: CLI argument
		if cliVal, ok := cliValues[key]; ok {
			value = cliVal
			found = true
		}

		// Priority 2: Environment variable
		if !found {
			if flag.EnvVar != "" {
				if envVal, ok := os.LookupEnv(flag.EnvVar); ok {
					var err error
					value, err = convertValue(envVal, flag.Type)
					if err != nil {
						return nil, fmt.Errorf("error parsing env var %s: %w", flag.EnvVar, err)
					}
					found = true
				}
			}
		}

		// Priority 3: Default value
		if !found {
			value = flag.Default
		}

		fs.parsedValues[key] = value
	}

	return &Config{
		values: fs.parsedValues,
		flags:  fs.flags,
	}, nil
}

func (fs *varg) setFlagValue(target map[string]interface{}, flag *Flag, key, val string) error {
	converted, err := convertValue(val, flag.Type)
	if err != nil {
		return fmt.Errorf("error parsing flag %s: %w", key, err)
	}

	// For slice types, append to existing slice
	if flag.Type == TypeStringSlice {
		if existing, ok := target[key]; ok {
			if slice, ok := existing.([]string); ok {
				target[key] = append(slice, val)
				return nil
			}
		}
		target[key] = []string{val}
		return nil
	}

	target[key] = converted
	return nil
}
