# varg

A lightweight, focused command-line flag parser for Go with environment variable support, nested configuration keys, and struct unmarshaling. Inspired by Go's standard `flag` package for initialization simplicity, but with a Cobra-like CLI syntax.

## Features

- **Simple initialization** — Like Go's `flag` package, no boilerplate
- **Cobra-style CLI syntax** — `--key value`, `-k value`, `--key=value`
- **Nested keys** — Dot notation: `--server.host localhost`
- **Environment variables** — Custom env var names per flag, with prefix support
- **Env var translation** — Nested keys use `__` instead of `.` in env vars
- **Struct tags** — Unmarshal into structs with `flag:"key"` tags
- **Type support** — String, Bool, StringSlice, `Int`/`Int8`/`Int16`/`Int32`/`Int64`, `Uint`/`Uint8`/`Uint16`/`Uint32`/`Uint64`, `Float32`/`Float64`
- **Precedence** — CLI args > environment variables > defaults
- **Built-in help & version** — `Handle()` returns `ShouldExit`, `Output` and the version string

## Installation

```bash
go get github.com/fmotalleb/varg
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"
    "github.com/fmotalleb/varg"
)

func main() {
    fs := varg.New("myapp")
    fs.Version("1.0.0")  // Enable --version flag
    
    // Define flags with dot notation for nesting
    fs.String("server.host", "", "localhost", "server host").EnvWithPrefix("APP")
    fs.Int("server.port", "", 8080, "server port").EnvWithPrefix("APP")
    fs.String("db.host", "", "localhost", "db host").Env("DB_HOST")
    fs.Int("verbose", "v", 0, "verbosity level")
    
    // Parse command-line arguments; Handle resolves --help and --version
    res := fs.Handle(os.Args[1:])
    if res.ShouldExit {
        fmt.Print(res.Output)     // usage text, "myapp 1.0.0" or the error
        if res.Err != nil {
            os.Exit(1)
        }
        os.Exit(0)
    }
    cfg := res.Config
    
    // Access values
    fmt.Printf("Server: %s:%d\n", cfg.String("server.host"), cfg.Int("server.port"))
    fmt.Printf("Verbosity: %d\n", cfg.Int("verbose"))
}
```

Usage:
```bash
# CLI arguments
./myapp --server.host api.example.com --server.port 9000

# Environment variables (with APP_ prefix from EnvWithPrefix)
APP_SERVER__HOST=api.example.com APP_SERVER__PORT=9000 ./myapp

# Custom env var names
DB_HOST=db.example.com ./myapp
```

## API Reference

### FlagSet

#### `New(name string) *FlagSet`
Creates a new FlagSet.

```go
fs := varg.New("myapp")
```

#### `String(key, short, defaultVal, help string) *Flag`
Adds a string flag.

```go
fs.String("output", "o", "output.log", "output file")
```

#### `Int(key, short string, defaultVal int, help string) *Flag`
Adds an int flag.

```go
fs.Int("port", "p", 8080, "listen port")
```

#### `Int8/Int16/Int32/Int64(key, short string, defaultVal <type>, help string) *Flag`
Adds a signed integer flag of the given width.

```go
fs.Int64("id", "i", int64(0), "unique id")
fs.Int16("temp", "", int16(0), "temperature")
```

#### `Uint/Uint8/Uint16/Uint32/Uint64(key, short string, defaultVal <type>, help string) *Flag`
Adds an unsigned integer flag. Negative values are rejected at parse time.

```go
fs.Uint16("mask", "", uint16(0), "netmask")
fs.Uint64("size", "", uint64(0), "size in bytes")
```

#### `Bool(key, short string, defaultVal bool, help string) *Flag`
Adds a bool flag.

```go
fs.Bool("verbose", "v", false, "verbose output")
```

#### `Float64(key, short string, defaultVal float64, help string) *Flag`
Adds a float64 flag.

```go
fs.Float64("threshold", "t", 0.5, "threshold")
```

#### `Float32(key, short string, defaultVal float32, help string) *Flag`
Adds a float32 flag.

```go
fs.Float32("ratio", "", float32(0), "ratio")
```

#### `StringSlice(key, short, help string) *Flag`
Adds a repeatable string slice flag.

```go
fs.StringSlice("tag", "t", "tags")
// Usage: -t foo -t bar -t baz
```

#### `GlobalEnvPrefix(prefix string) *FlagSet`
Sets a global prefix for environment variables (chaining).

```go
fs.GlobalEnvPrefix("MYAPP").String("output", "o", "out.log", "output")
// Looks for MYAPP_OUTPUT env var
```

#### `Version(v string) *FlagSet`
Sets the version string for the automatic `--version` flag (chaining).

```go
fs.Version("1.2.3")
```

#### `DisableHelp() *FlagSet`
Disables the automatic `--help(-h)` flag (chaining).

```go
fs.DisableHelp()
```

#### `DisableVersion() *FlagSet`
Disables the automatic `--version(-v)` flag (chaining).

```go
fs.DisableVersion()
```

#### `Usage() string`
Returns a formatted usage string (help text).

```go
fmt.Print(fs.Usage())
```

#### `Parse(args []string) (*Config, error)`
Parses command-line arguments and returns a Config with resolved values.

Returns special errors for built-in flags — match them with `errors.Is`:
- `errors.Is(err, varg.ErrHelp)` when `--help` or `-h` is used
- `errors.Is(err, varg.ErrVersion)` when `--version` or `-v` is used

```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil {
    if errors.Is(err, varg.ErrHelp) {
        fmt.Print(fs.Usage())
        os.Exit(0)
    }
    log.Fatal(err)
}
```

#### `Handle(args []string) Result`
Parses args like `Parse`, but resolves the built-in help and version flags itself and
returns everything needed to print and exit.

| Field | Description |
| --- | --- |
| `Config` | Parsed values, `nil` unless parsing succeeded |
| `ShouldExit` | Print `Output` before the program stops |
| `Output` | Usage text for `--help`, `myapp 1.2.3` for `--version`, the error message otherwise |
| `Err` | `nil` on success, help and version; set when parsing failed (use it as exit code) |
| `Version` | The version string configured with `Version()`, empty when unset |
| `Help` | The usage text of this flag set |

```go
res := fs.Handle(os.Args[1:])
if res.ShouldExit {
    fmt.Print(res.Output)
    if res.Err != nil {
        os.Exit(1)
    }
    os.Exit(0)
}
cfg := res.Config
fmt.Println("version:", res.Version)
```

### Flag

#### `Env(envVar string) *Flag`
Sets a custom environment variable name for this flag (chaining).

```go
fs.String("output", "o", "out.log", "output").Env("OUTPUT_FILE")
// Reads from OUTPUT_FILE env var
```

#### `EnvWithPrefix(prefix string) *Flag`
Sets environment variable with a prefix, translating nested keys (chaining).

```go
fs.String("server.host", "", "localhost", "host").EnvWithPrefix("APP")
// Reads from APP_SERVER__HOST env var
// (dots become double underscores)
```

### Config

#### `Get(key string) (interface{}, bool)`
Gets the raw value for a key.

```go
val, ok := cfg.Get("server.host")
```

#### `String(key string) string`
Gets a string value.

```go
host := cfg.String("server.host")
```

#### `Int(key string) int`
Gets an int value.

```go
port := cfg.Int("server.port")
```

#### `Int8/Int16/Int32/Int64(key string) <type>` · `Uint/Uint8/Uint16/Uint32/Uint64(key string) <type>`
Width specific integer getters. They convert between numeric types and return
the zero value when the value does not fit (a negative value in a `Uint`
getter, an out of range value in a sized getter).

```go
id := cfg.Int64("id")
mask := cfg.Uint16("mask")
```

#### `Bool(key string) bool`
Gets a bool value.

```go
debug := cfg.Bool("debug")
```

#### `Float64(key string) float64`
Gets a float64 value.

```go
threshold := cfg.Float64("threshold")
```

#### `Float32(key string) float32`
Gets a float32 value (integer values are converted).

```go
ratio := cfg.Float32("ratio")
```

#### `StringSlice(key string) []string`
Gets a string slice value.

```go
tags := cfg.StringSlice("tag")
```

#### `All() map[string]interface{}`
Returns all parsed values as a map.

```go
values := cfg.All()
```

#### `Unmarshal(v interface{}) error`
Populates a struct from parsed flag values using `flag:"key"` struct tags.

```go
var config struct {
    Server struct {
        Host string `flag:"server.host"`
        Port int    `flag:"server.port"`
    }
    Debug bool `flag:"debug"`
}
cfg.Unmarshal(&config)
```

## Examples

### Basic Usage

```go
fs := varg.New("app")
fs.String("input", "i", "input.txt", "input file")
fs.String("output", "o", "output.txt", "output file")
fs.Int("threads", "t", 4, "thread count")

cfg, _ := fs.Parse(os.Args[1:])

input := cfg.String("input")
output := cfg.String("output")
threads := cfg.Int("threads")
```

### Environment Variables

```go
fs := varg.New("app")
fs.String("db.host", "", "localhost", "db host").Env("DATABASE_HOST")
fs.Int("db.port", "", 5432, "db port").Env("DATABASE_PORT")

// Usage:
// DATABASE_HOST=postgres.example.com DATABASE_PORT=5433 ./app
```

### Nested Configuration with Prefixes

```go
fs := varg.New("app")

// All env vars for these flags will use APP_ prefix
fs.String("server.host", "", "localhost", "host").EnvWithPrefix("APP")
fs.Int("server.port", "", 8080, "port").EnvWithPrefix("APP")
fs.String("db.host", "", "localhost", "db host").EnvWithPrefix("APP")
fs.Int("db.port", "", 5432, "db port").EnvWithPrefix("APP")

cfg, _ := fs.Parse(os.Args[1:])

// Env vars: APP_SERVER__HOST, APP_SERVER__PORT, APP_DB__HOST, APP_DB__PORT
```

### Struct Unmarshaling

```go
fs := varg.New("app")
fs.String("server.host", "", "localhost", "host")
fs.Int("server.port", "", 8080, "port")
fs.String("db.host", "", "localhost", "db host")
fs.Int("db.port", "", 5432, "db port")

cfg, _ := fs.Parse(os.Args[1:])

type Config struct {
    Server struct {
        Host string `flag:"server.host"`
        Port int    `flag:"server.port"`
    }
    DB struct {
        Host string `flag:"db.host"`
        Port int    `flag:"db.port"`
    }
}

var config Config
cfg.Unmarshal(&config)

// Now use config.Server.Host, config.DB.Port, etc.
```

### Repeatable Flags

```go
fs := varg.New("app")
fs.StringSlice("tag", "t", "tags to apply")

cfg, _ := fs.Parse([]string{"-t", "important", "-t", "urgent", "-t", "review"})

tags := cfg.StringSlice("tag")  // [important urgent review]
```

### Precedence: CLI > Env > Default

```go
fs := varg.New("app")
fs.String("config", "c", "default.conf", "config file").Env("CONFIG_FILE")

// Priority 1: CLI argument
cfg, _ := fs.Parse([]string{"--config", "cli.conf"})
// Result: "cli.conf"

// Priority 2: Environment variable
os.Setenv("CONFIG_FILE", "env.conf")
cfg, _ = fs.Parse([]string{})
// Result: "env.conf"

// Priority 3: Default
os.Unsetenv("CONFIG_FILE")
cfg, _ = fs.Parse([]string{})
// Result: "default.conf"
```

## CLI Syntax

### Long form with space
```bash
./app --output result.txt
```

### Long form with equals
```bash
./app --output=result.txt
```

### Short form with space
```bash
./app -o result.txt
```

### Nested keys with dots
```bash
./app --server.host localhost --server.port 9000
./app --db.server.host postgres.example.com
```

### Boolean flags
```bash
./app --verbose
./app --debug=true
./app --debug=false
```

### Numeral shorthand (counting)
For int flags, repeated short flags count automatically:
```bash
./app -vvv          # Same as -v 3
./app -v -v -v      # Same as -vvv
./app -dd           # Same as -d 2
```

Useful for verbosity levels or debug depth.

### Increment/decrement with +
Use `+` prefix to pass negative value (decrement):
```bash
./app +v            # Sets flag to -1
./app +count        # Decrements counter
```

### Built-in flags

Automatically available (unless disabled):
```bash
./app --help        # Show usage and available flags
./app -h            # Short form
./app --version     # Show version (if set with .Version())
./app -v            # Short form
```

Registrations are override based — the last registration of a name wins.
Declaring your own flag after a built-in takes over its name:
```go
fs.Version("1.0.0")                    // claims --version and -v
fs.Int("verbose", "v", 0, "verbosity") // -v now counts, --version still prints
```

## Environment Variable Naming

- Custom env vars: Use `.Env("EXACT_VAR_NAME")`
- With prefix: Use `.EnvWithPrefix("PREFIX")` — nested keys convert `.` to `__`

Examples:
- `flag:"output"` with `EnvWithPrefix("APP")` → reads `APP_OUTPUT`
- `flag:"server.host"` with `EnvWithPrefix("APP")` → reads `APP_SERVER__HOST`
- `flag:"db.server.port"` with `EnvWithPrefix("CONFIG")` → reads `CONFIG_DB__SERVER__PORT`

The help output shows the environment variable of every flag that has one:

```text
$ ./myapp --help
Options:
  -h, --help            show this help message
  -o, --output          output file [$OUTPUT_FILE]
      --server.host     server host [$APP_SERVER__HOST]
```

## Testing

Run the test suite:

```bash
go test -v ./...
```

The package includes comprehensive tests covering:
- Basic flag parsing (string, int, bool, float64, slice)
- CLI argument syntax variants
- Environment variable resolution
- Precedence (CLI > env > default)
- Nested keys with prefix translation
- Struct tag unmarshaling
- Error cases

## Design Philosophy

- **Minimal** — Only core flag parsing, no complex features
- **Simple initialization** — Like Go's `flag` package
- **Type-safe** — Explicit getters for each type
- **Chainable** — Fluent API for flag definition
- **Environment-first** — First-class env var support
- **Composable** — Works well in layered configs (file + env + CLI)

## License

MIT
