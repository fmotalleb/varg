package varg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Sentinel errors returned when the built-in help/version flags are requested.
var (
	ErrHelp    = errors.New("__HELP__")
	ErrVersion = errors.New("__VERSION__")
)

// lookup resolves a flag name (long or short form) to either a user flag or a
// built-in flag. Registrations are override based: the last registration for a
// given name wins, so a flag registered after a built-in takes that name over.
type lookup struct {
	flagKey string // key inside FlagSet.flags; empty for built-in flags
	builtin error  // sentinel error for built-in flags; nil for user flags
}

// FlagSet represents a set of command-line flags.
type FlagSet struct {
	name         string
	flags        map[string]*Flag
	longs        map[string]lookup
	shorts       map[string]lookup
	globalPrefix string
	parsedValues map[string]interface{}
	version      string
}

// New creates a new FlagSet with the given name.
// The built-in help flag (--help, -h) is registered first, so any later
// registration of the same name overrides it.
func New(name string) *FlagSet {
	fs := &FlagSet{
		name:         name,
		flags:        make(map[string]*Flag),
		longs:        make(map[string]lookup),
		shorts:       make(map[string]lookup),
		parsedValues: make(map[string]interface{}),
	}
	fs.registerBuiltin("help", "h", ErrHelp)
	return fs
}

// String adds a string flag to the set.
func (fs *FlagSet) String(key, short, defaultVal, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeString)
}

// Int adds an int flag to the set.
func (fs *FlagSet) Int(key, short string, defaultVal int, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt)
}

// Int8 adds an int8 flag to the set.
func (fs *FlagSet) Int8(key, short string, defaultVal int8, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt8)
}

// Int16 adds an int16 flag to the set.
func (fs *FlagSet) Int16(key, short string, defaultVal int16, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt16)
}

// Int32 adds an int32 flag to the set.
func (fs *FlagSet) Int32(key, short string, defaultVal int32, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt32)
}

// Int64 adds an int64 flag to the set.
func (fs *FlagSet) Int64(key, short string, defaultVal int64, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeInt64)
}

// Uint adds a uint flag to the set.
func (fs *FlagSet) Uint(key, short string, defaultVal uint, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeUint)
}

// Uint8 adds a uint8 flag to the set.
func (fs *FlagSet) Uint8(key, short string, defaultVal uint8, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeUint8)
}

// Uint16 adds a uint16 flag to the set.
func (fs *FlagSet) Uint16(key, short string, defaultVal uint16, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeUint16)
}

// Uint32 adds a uint32 flag to the set.
func (fs *FlagSet) Uint32(key, short string, defaultVal uint32, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeUint32)
}

// Uint64 adds a uint64 flag to the set.
func (fs *FlagSet) Uint64(key, short string, defaultVal uint64, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeUint64)
}

// Bool adds a bool flag to the set.
func (fs *FlagSet) Bool(key, short string, defaultVal bool, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeBool)
}

// Float64 adds a float64 flag to the set.
func (fs *FlagSet) Float64(key, short string, defaultVal float64, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeFloat64)
}

// Float32 adds a float32 flag to the set.
func (fs *FlagSet) Float32(key, short string, defaultVal float32, help string) *Flag {
	return fs.addFlag(key, short, defaultVal, help, TypeFloat32)
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
	fs.longs[key] = lookup{flagKey: key}
	if short != "" {
		fs.shorts[short] = lookup{flagKey: key}
	}
	return flag
}

// registerBuiltin claims the long and short names for a built-in flag.
// Any name claimed earlier (by a previous built-in) is replaced.
func (fs *FlagSet) registerBuiltin(long, short string, err error) {
	fs.longs[long] = lookup{builtin: err}
	if short != "" {
		fs.shorts[short] = lookup{builtin: err}
	}
}

// unregisterBuiltin releases the long and short names of a built-in flag.
// Names that have been taken over by other registrations are left untouched.
func (fs *FlagSet) unregisterBuiltin(long, short string, err error) {
	if e, ok := fs.longs[long]; ok && e.builtin == err {
		delete(fs.longs, long)
	}
	if e, ok := fs.shorts[short]; ok && e.builtin == err {
		delete(fs.shorts, short)
	}
}

// GlobalEnvPrefix sets a prefix for all env vars (e.g., "MYAPP_").
// Individual flag env vars will be appended to this.
func (fs *FlagSet) GlobalEnvPrefix(prefix string) *FlagSet {
	fs.globalPrefix = prefix
	return fs
}

// Version sets the version string for --version output.
// It registers the --version and -v names; a flag registered afterwards
// overrides whichever of those names it uses.
func (fs *FlagSet) Version(v string) *FlagSet {
	fs.version = v
	fs.registerBuiltin("version", "v", ErrVersion)
	return fs
}

// DisableHelp disables the automatic --help(-h) flag.
// Names already taken over by user flags are kept.
func (fs *FlagSet) DisableHelp() *FlagSet {
	fs.unregisterBuiltin("help", "h", ErrHelp)
	return fs
}

// DisableVersion disables the automatic --version(-v) flag.
// Names already taken over by user flags are kept.
func (fs *FlagSet) DisableVersion() *FlagSet {
	fs.unregisterBuiltin("version", "v", ErrVersion)
	return fs
}

// builtinNames returns the usage display name of a built-in flag and whether
// it is still registered. The short form is only shown while it owns it.
func (fs *FlagSet) builtinNames(long, short string, err error) (string, bool) {
	if e, ok := fs.longs[long]; !ok || e.builtin != err {
		return "", false
	}
	if e, ok := fs.shorts[short]; ok && e.builtin == err {
		return fmt.Sprintf("  -%s, --%s", short, long), true
	}
	return fmt.Sprintf("      --%s", long), true
}

// Usage returns a formatted usage string.
func (fs *FlagSet) Usage() string {
	var buf strings.Builder
	buf.WriteString("Usage: " + fs.name + " [options]\n\n")
	buf.WriteString("Options:\n")

	if name, ok := fs.builtinNames("help", "h", ErrHelp); ok {
		fmt.Fprintf(&buf, "%-30s %s\n", name, "show this help message")
	}
	if fs.version != "" {
		if name, ok := fs.builtinNames("version", "v", ErrVersion); ok {
			fmt.Fprintf(&buf, "%-30s %s\n", name, "show version")
		}
	}

	for _, flag := range fs.flags {
		var optStr string
		if flag.Short != "" {
			optStr = fmt.Sprintf("  -%s, --%s", flag.Short, flag.Key)
		} else {
			optStr = fmt.Sprintf("  --%s", flag.Key)
		}
		fmt.Fprintf(&buf, "%-30s %s\n", optStr, flag.helpText())
	}

	return buf.String()
}

// parseShortNumeral counts repeated short flags (e.g., -ddd -> 3).
// Returns the count and whether it was a numeral pattern.
func parseShortNumeral(short string) (int, bool) {
	if len(short) < 2 {
		return 0, false
	}

	first := short[0]
	for i := 1; i < len(short); i++ {
		if short[i] != first {
			return 0, false
		}
	}

	return len(short), true
}

// takesNextArg reports whether args[i+1] must be read as the value of flag.
func takesNextArg(args []string, i int, flag *Flag) bool {
	if i+1 >= len(args) {
		return false
	}
	next := args[i+1]
	if !strings.HasPrefix(next, "-") && !strings.HasPrefix(next, "+") {
		return true
	}
	// Signed values (-1, +5) only qualify for numeric flags. Everything else
	// stays an argument of its own, so -v -v and --out -h keep working.
	return flag.Type.isNumeric() && isNumericLiteral(next)
}

// isNumericLiteral reports whether s looks like a number (-1, +5, 1.5e3).
func isNumericLiteral(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
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
			key, val, hasValue := strings.Cut(arg[2:], "=")

			entry, exists := fs.longs[key]
			if !exists {
				return nil, fmt.Errorf("unknown flag: --%s", key)
			}
			if entry.builtin != nil {
				return nil, entry.builtin
			}
			flag := fs.flags[entry.flagKey]

			// If no value provided and not from =, try to get next arg
			if !hasValue {
				if takesNextArg(args, i, flag) {
					val = args[i+1]
					i++
				} else if flag.Type.isNumeric() {
					// Valueless numeric flag acts as a single increment.
					if err := fs.adjustFlagValue(cliValues, flag, key, 1); err != nil {
						return nil, err
					}
					continue
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

			entry, exists := fs.shorts[short]
			if !exists || entry.builtin != nil {
				return nil, fmt.Errorf("unknown flag: +%s", short)
			}

			flag := fs.flags[entry.flagKey]
			if !flag.Type.isNumeric() || flag.Type.isUnsigned() {
				return nil, fmt.Errorf("flag %s cannot be decremented", entry.flagKey)
			}

			// +k means decrement (negative value)
			if err := fs.setFlagValue(cliValues, flag, entry.flagKey, "-1"); err != nil {
				return nil, err
			}
			continue
		}

		// Handle short form: -k value or -kkk (numeral) or -k
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			short := arg[1:]

			// Check for numeral shorthand: -ddd means -d three times.
			// Only numeric flags support this syntax.
			if count, isNumeral := parseShortNumeral(short); isNumeral {
				shortChar := string(short[0])

				entry, exists := fs.shorts[shortChar]
				if !exists {
					return nil, fmt.Errorf("unknown flag: -%s", shortChar)
				}
				if entry.builtin != nil {
					return nil, entry.builtin
				}

				flag := fs.flags[entry.flagKey]
				switch {
				case flag.Type.isNumeric():
					// Counting adds to the value, so -vvv, -v -v -v and
					// mixed forms all agree.
					if err := fs.adjustFlagValue(cliValues, flag, entry.flagKey, count); err != nil {
						return nil, err
					}
					continue
				case flag.Type == TypeBool:
					if err := fs.setFlagValue(cliValues, flag, entry.flagKey, "true"); err != nil {
						return nil, err
					}
					continue
				}

				// Not a numeric flag, so interpret it as a normal short flag.
			}

			entry, exists := fs.shorts[short]
			if !exists {
				return nil, fmt.Errorf("unknown flag: -%s", short)
			}
			if entry.builtin != nil {
				return nil, entry.builtin
			}
			flag := fs.flags[entry.flagKey]

			var val string
			if takesNextArg(args, i, flag) {
				val = args[i+1]
				i++
			} else if flag.Type.isNumeric() {
				// Valueless numeric flag acts as a single increment, so
				// -v -v -v matches -vvv.
				if err := fs.adjustFlagValue(cliValues, flag, entry.flagKey, 1); err != nil {
					return nil, err
				}
				continue
			} else {
				val = "true"
			}

			if err := fs.setFlagValue(cliValues, flag, entry.flagKey, val); err != nil {
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

// adjustFlagValue increments a numeric flag relative to its current CLI value
// (or its default when it has not been set yet).
func (fs *FlagSet) adjustFlagValue(target map[string]interface{}, flag *Flag, key string, delta int) error {
	cur, ok := target[key]
	if !ok {
		cur = flag.Default
	}

	var next string
	switch rv := reflect.ValueOf(cur); rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		next = strconv.FormatInt(rv.Int()+int64(delta), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		next = strconv.FormatUint(rv.Uint()+uint64(delta), 10)
	case reflect.Float32, reflect.Float64:
		next = strconv.FormatFloat(rv.Float()+float64(delta), 'g', -1, 64)
	default:
		return fmt.Errorf("flag %s cannot be incremented", key)
	}

	converted, err := convertValue(next, flag.Type)
	if err != nil {
		return fmt.Errorf("error parsing flag %s: %w", key, err)
	}
	target[key] = converted
	return nil
}

func (fs *FlagSet) setFlagValue(target map[string]interface{}, flag *Flag, key, val string) error {
	converted, err := convertValue(val, flag.Type)
	if err != nil {
		return fmt.Errorf("error parsing flag %s: %w", key, err)
	}

	if flag.Type == TypeStringSlice {
		slice, _ := target[key].([]string)
		target[key] = append(slice, converted.([]string)...)
		return nil
	}

	target[key] = converted
	return nil
}

// Result is the outcome of Handle. It carries the parsed config together with
// everything the built-in help and version flags produced.
type Result struct {
	// Config holds the parsed values. It is nil unless parsing succeeded.
	Config *Config

	// ShouldExit reports that Output has to be printed before the program
	// stops: help or version was requested, or parsing failed.
	ShouldExit bool

	// Output is the text to print when ShouldExit is set: the usage text for
	// --help, "<name> <version>" for --version, or the error message.
	Output string

	// Err is nil for a successful parse and for help/version requests, and is
	// set when parsing failed. Use it to choose the exit code (0 or 1).
	Err error

	// Version is the version string configured with Version(), empty when unset.
	Version string

	// Help is the usage text of this flag set.
	Help string
}

// Handle parses args like Parse, but resolves the built-in help and version
// flags itself, so callers never have to compare internal error messages:
//
//	res := fs.Handle(os.Args[1:])
//	if res.ShouldExit {
//		fmt.Print(res.Output)
//		if res.Err != nil {
//			os.Exit(1)
//		}
//		os.Exit(0)
//	}
//	cfg := res.Config
func (fs *FlagSet) Handle(args []string) Result {
	res := Result{
		Version: fs.version,
		Help:    fs.Usage(),
	}

	cfg, err := fs.Parse(args)
	switch {
	case err == nil:
		res.Config = cfg
	case errors.Is(err, ErrHelp):
		res.ShouldExit = true
		res.Output = res.Help
	case errors.Is(err, ErrVersion):
		res.ShouldExit = true
		res.Output = strings.TrimSpace(fs.name+" "+fs.version) + "\n"
	default:
		res.ShouldExit = true
		res.Err = err
		res.Output = err.Error() + "\n"
	}

	return res
}
