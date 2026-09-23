package varg

import (
	"fmt"
	"os"
	"strings"
)

// FlagSet represents a set of command-line flags.
type FlagSet struct {
	name          string
	flags         map[string]*Flag
	globalPrefix  string
	parsedValues  map[string]interface{}
	version       string
	helpFlag      bool
	versionFlag   bool
}

// New creates a new FlagSet with the given name.
func New(name string) *FlagSet {
	return &FlagSet{
		name:         name,
		flags:        make(map[string]*Flag),
		parsedValues: make(map[string]interface{}),
		helpFlag:     true,
		versionFlag:  true,
	}
}

// String adds a string flag to the set.
func (fs *FlagSet) String(key, short, defaultVal, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeString)
}

// Int adds an int flag to the set.
func (fs *FlagSet) Int(key, short string, defaultVal int, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt)
}

// Bool adds a bool flag to the set.
func (fs *FlagSet) Bool(key, short string, defaultVal bool, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeBool)
}

// Float64 adds a float64 flag to the set.
func (fs *FlagSet) Float64(key, short string, defaultVal float64, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeFloat64)
}

// StringSlice adds a string slice flag to the set (repeatable: -tag foo -tag bar).
func (fs *FlagSet) StringSlice(key, short string, help string) *Flag {
	return fs.addFlag(key, short, []string{}, help, TypeStringSlice)
}

func (fs *FlagSet) addFlag(key, short string, defaultVal interface{}, help string, typ Type) *Flag {
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
func (fs *FlagSet) GlobalEnvPrefix(prefix string) *FlagSet {
	fs.globalPrefix = prefix
	return fs
}

// Version sets the version string for --version output.
func (fs *FlagSet) Version(v string) *FlagSet {
	fs.version = v
	return fs
}

// DisableHelp disables the automatic --help(-h) flag.
func (fs *FlagSet) DisableHelp() *FlagSet {
	fs.helpFlag = false
	return fs
}

// DisableVersion disables the automatic --version(-v) flag.
func (fs *FlagSet) DisableVersion() *FlagSet {
	fs.versionFlag = false
	return fs
}

// Usage returns a formatted usage string.
func (fs *FlagSet) Usage() string {
	var buf strings.Builder
	buf.WriteString("Usage: " + fs.name + " [options]\n\n")
	buf.WriteString("Options:\n")
	
	if fs.helpFlag {
		buf.WriteString("  -h, --help            show this help message\n")
	}
	if fs.versionFlag && fs.version != "" {
		buf.WriteString("  -v, --version         show version\n")
	}
	
	for _, flag := range fs.flags {
		var optStr string
		if flag.Short != "" {
			optStr = fmt.Sprintf("  -%s, --%s", flag.Short, flag.Key)
		} else {
			optStr = fmt.Sprintf("  --%s", flag.Key)
		}
		buf.WriteString(fmt.Sprintf("%-30s %s\n", optStr, flag.Help))
	}
	
	return buf.String()
}

// parseShortNumeral counts repeated short flags (e.g., -ddd -> 3).
// Returns the count and whether it was a numeral pattern.
func parseShortNumeral(short string) (int, bool) {
	if len(short) < 1 {
		return 0, false
	}
	
	// Check for all same character
	first := short[0]
	for _, ch := range short {
		if ch != rune(first) {
			return 0, false
		}
	}
	
	return len(short), true
}

// Parse parses command-line arguments and environment variables.
// Returns a Config with resolved values following precedence: CLI > env > default.
// Returns an error if --help or --version is requested (caller should handle display).
func (fs *FlagSet) Parse(args []string) (*Config, error) {
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
				val = ""
			}

			// Check for built-in flags first
			if key == "help" && fs.helpFlag {
				return nil, fmt.Errorf("__HELP__")
			}
			if key == "version" && fs.versionFlag {
				return nil, fmt.Errorf("__VERSION__")
			}

			flag, exists := fs.flags[key]
			if !exists {
				return nil, fmt.Errorf("unknown flag: --%s", key)
			}

			// If no value provided and not from =, try to get next arg
			if val == "" && !ok {
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.HasPrefix(args[i+1], "+") {
					val = args[i+1]
					i++
				} else {
					val = "true"
				}
			}

			if err := fs.setFlagValue(cliValues, flag, key, val); err != nil {
				return nil, err
			}
			continue
		}

		// Handle increment/decrement form: +k or -k
		if strings.HasPrefix(arg, "+") && len(arg) > 1 {
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
				return nil, fmt.Errorf("unknown flag: +%s", short)
			}

			// +k means decrement (negative value)
			val := "-1"
			if err := fs.setFlagValue(cliValues, flag, flagKey, val); err != nil {
				return nil, err
			}
			continue
		}

		// Handle short form: -k value or -kkk (numeral) or -k
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]

			// Check for numeral shorthand: -ddd means -d 3
			if count, isNumeral := parseShortNumeral(short); isNumeral {
				shortChar := string(short[0])
				
				// Check for built-in flags first
				if shortChar == "h" && fs.helpFlag {
					return nil, fmt.Errorf("__HELP__")
				}
				if shortChar == "v" && fs.versionFlag {
					return nil, fmt.Errorf("__VERSION__")
				}
				
				var flag *Flag
				var flagKey string
				for k, f := range fs.flags {
					if f.Short == shortChar {
						flag = f
						flagKey = k
						break
					}
				}

				if flag == nil {
					return nil, fmt.Errorf("unknown flag: -%s", shortChar)
				}

				val := fmt.Sprintf("%d", count)
				if err := fs.setFlagValue(cliValues, flag, flagKey, val); err != nil {
					return nil, err
				}
				continue
			}

			// Check for built-in flags first (before looking up user flags)
			if short == "h" && fs.helpFlag {
				return nil, fmt.Errorf("__HELP__")
			}
			if short == "v" && fs.versionFlag {
				return nil, fmt.Errorf("__VERSION__")
			}

			// Regular short form
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
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.HasPrefix(args[i+1], "+") {
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
				if envVal, ok := os.LookupEnv(flag.EnvVar); ok && envVal != "" {
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

func (fs *FlagSet) setFlagValue(target map[string]interface{}, flag *Flag, key, val string) error {
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
