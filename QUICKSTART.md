# varg Quick Start

## Initialization (like `flag` package)

```go
fs := varg.New("myapp")

// Set version (enables --version flag)
fs.Version("1.2.3")

// Add flags one at a time
fs.String("output", "o", "default.txt", "output file")
fs.Int("port", "p", 8080, "listen port")
fs.Int("verbose", "v", 0, "verbosity level")

// Optional: disable built-in help/version
fs.DisableHelp()
fs.DisableVersion()
```

## Environment Variables

### Custom env var per flag
```go
fs.String("database", "d", "sqlite.db", "database").Env("DB_PATH")
// Reads from DB_PATH env var
```

### Prefix for group of flags
```go
fs.String("server.host", "", "localhost", "host").EnvWithPrefix("APP")
fs.Int("server.port", "", 8080, "port").EnvWithPrefix("APP")
// Reads from: APP_SERVER__HOST, APP_SERVER__PORT
```

## Nested Keys (Dot Notation)

```go
fs.String("db.server.host", "", "localhost", "db host")
fs.Int("db.server.port", "", 5432, "db port")

cfg, _ := fs.Parse(os.Args[1:])
cfg.String("db.server.host")  // --db.server.host value
cfg.Int("db.server.port")     // --db.server.port value
```

## Parsing

```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil {
    log.Fatal(err)
}
```

## Getting Values

```go
// Type-safe getters
host := cfg.String("server.host")
port := cfg.Int("server.port")
debug := cfg.Bool("debug")
threshold := cfg.Float64("threshold")
tags := cfg.StringSlice("tag")

// Raw value
val, ok := cfg.Get("output")

// All values
all := cfg.All()
```

## Struct Tags

```go
type Config struct {
    Server struct {
        Host string `flag:"server.host"`
        Port int    `flag:"server.port"`
    }
    Debug bool `flag:"debug"`
}

var config Config
cfg.Unmarshal(&config)
```

## Repeatable Flags

```go
fs.StringSlice("tag", "t", "tags")

// Usage: -t foo -t bar -t baz
cfg, _ := fs.Parse([]string{"-t", "foo", "-t", "bar"})
tags := cfg.StringSlice("tag")  // [foo bar]
```

## CLI Usage Examples

```bash
# Long form
./app --output result.txt --port 9000

# Short form
./app -o result.txt -p 9000

# Equals syntax
./app --output=result.txt --port=9000

# Nested keys
./app --server.host localhost --server.port 8080

# Boolean
./app --verbose
./app --debug=true

# Repeatable
./app -t important -t urgent -t review
```

## Precedence

1. **CLI arguments** — Highest priority
2. **Environment variables**
3. **Default values** — Lowest priority

```go
fs.String("config", "c", "default.conf", "config").Env("CONFIG_FILE")

// CLI wins
cfg, _ := fs.Parse([]string{"--config", "cli.conf"})
// Result: "cli.conf"

// Env is used if no CLI arg
cfg, _ := fs.Parse([]string{})
// With CONFIG_FILE=env.conf → "env.conf"
// Without → "default.conf"
```

## Handling Help and Version

```go
fs := varg.New("myapp")
fs.Version("1.2.3")
fs.String("output", "o", "out.log", "output file")

cfg, err := fs.Parse(os.Args[1:])
if err != nil {
    // Handle built-in flags
    if err.Error() == "__HELP__" {
        fmt.Print(fs.Usage())
        os.Exit(0)
    }
    if err.Error() == "__VERSION__" {
        fmt.Printf("myapp %s\n", "1.2.3")
        os.Exit(0)
    }
    // Real errors
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

## Numeral Shorthand

For int flags, repeat the short flag to count:

```go
fs.Int("verbose", "v", 0, "verbosity")

// Usage:
// ./app -vvv         # verbose = 3
// ./app -vv          # verbose = 2
// ./app -v           # verbose = 1
```

## Increment/Decrement with +

Use `+` prefix to set negative value (useful for counters):

```go
fs.Int("priority", "p", 0, "priority")

// Usage:
// ./app +p            # priority = -1 (decrement)
```

## Complete Example

```go
package main

import (
    "fmt"
    "os"
    "github.com/fmotalleb/varg"
)

func main() {
    fs := varg.New("myapp")
    fs.Version("1.0.0")
    
    fs.String("input", "i", "input.txt", "input file").Env("INPUT_FILE")
    fs.String("output", "o", "output.txt", "output file").Env("OUTPUT_FILE")
    fs.String("server.host", "", "localhost", "server host").EnvWithPrefix("APP")
    fs.Int("server.port", "", 8080, "server port").EnvWithPrefix("APP")
    fs.Int("verbose", "v", 0, "verbosity level (use -vvv for level 3)")
    
    cfg, err := fs.Parse(os.Args[1:])
    if err != nil {
        // Handle built-in help/version flags
        if err.Error() == "__HELP__" {
            fmt.Print(fs.Usage())
            os.Exit(0)
        }
        if err.Error() == "__VERSION__" {
            fmt.Println("myapp version 1.0.0")
            os.Exit(0)
        }
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Input: %s\n", cfg.String("input"))
    fmt.Printf("Output: %s\n", cfg.String("output"))
    fmt.Printf("Server: %s:%d\n", cfg.String("server.host"), cfg.Int("server.port"))
    fmt.Printf("Verbosity: %d\n", cfg.Int("verbose"))
}
```

Usage examples:
```bash
./myapp --help                     # Show help
./myapp --version                  # Show version
./myapp -vvv --output result.txt   # Level 3 verbosity
./myapp -v                         # Level 1 verbosity
```

## Environment Variable Mapping

### Without `.EnvWithPrefix()`
```go
fs.String("output", "o", "out.log", "output").Env("OUTPUT_FILE")
// Reads: OUTPUT_FILE
```

### With `.EnvWithPrefix()`
```go
fs.String("output", "o", "out.log", "output").EnvWithPrefix("APP")
// Reads: APP_OUTPUT

fs.String("server.host", "", "localhost", "host").EnvWithPrefix("APP")
// Reads: APP_SERVER__HOST  (dot → double underscore)

fs.String("db.server.host", "", "localhost", "host").EnvWithPrefix("CONFIG")
// Reads: CONFIG_DB__SERVER__HOST
```

## Type Support

| Type | Flag Method | Get Method | Default Example |
|------|---|---|---|
| String | `String()` | `cfg.String()` | `"default"` |
| Int | `Int()` | `cfg.Int()` | `42` |
| Bool | `Bool()` | `cfg.Bool()` | `false` |
| Float64 | `Float64()` | `cfg.Float64()` | `3.14` |
| []String | `StringSlice()` | `cfg.StringSlice()` | `[]string{}` |

## No Subcommands

This library intentionally does not support subcommands (like `app serve`, `app config`). If you need that, use Cobra. `varg` is for single-command tools where you just want to parse flags simply.
