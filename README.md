# varg

A lightweight command-line flag parser for Go with first-class environment variable support, nested configuration keys, and struct unmarshaling.

`varg` is designed for small to medium-sized command-line applications that need more configuration flexibility than Go's standard `flag` package, without the complexity of a full CLI framework.

It combines simple `flag`-style initialization with a Cobra-like CLI syntax:

```text
--key value
--key=value
-k value
```

Configuration values can come from command-line arguments, environment variables, or defaults.

## Features

* Simple initialization with a small API
* Cobra-style long and short flag syntax
* `--key value` and `--key=value`
* Nested configuration keys using dot notation
* Environment variable support
* Custom environment variable names
* Environment variable prefixes
* Automatic nested-key environment variable mapping
* Struct unmarshaling using `flag:"..."` tags
* String, boolean, integer, unsigned integer, floating-point, and string-slice types
* Repeatable string-slice flags
* CLI > environment > default precedence
* Automatic `--help` / `-h`
* Optional `--version` / `-v`
* Formatted usage output
* Repeated short integer flags for counters, such as `-vvv`
* Minimal dependencies
* Chainable flag definitions

## Requirements

* Go 1.21 or newer

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

	fs.Version("1.0.0")

	fs.String(
		"server.host",
		"",
		"localhost",
		"server host",
	).EnvWithPrefix("APP")

	fs.Int(
		"server.port",
		"",
		8080,
		"server port",
	).EnvWithPrefix("APP")

	fs.Int(
		"verbose",
		"v",
		0,
		"verbosity level",
	)

	res := fs.Handle(os.Args[1:])
	if res.ShouldExit {
		fmt.Print(res.Output)

		if res.Err != nil {
			os.Exit(1)
		}

		os.Exit(0)
	}

	cfg := res.Config

	fmt.Printf(
		"Server: %s:%d\n",
		cfg.String("server.host"),
		cfg.Int("server.port"),
	)

	fmt.Printf("Verbosity: %d\n", cfg.Int("verbose"))
}
```

Run it with command-line arguments:

```bash
./myapp --server.host api.example.com --server.port 9000
```

Or use environment variables:

```bash
APP_SERVER__HOST=api.example.com \
APP_SERVER__PORT=9000 \
./myapp
```

## Configuration Precedence

When a flag has both a default value and an environment variable configured, values are resolved in this order:

```text
CLI arguments
      ↓
Environment variables
      ↓
Default values
```

For example:

```go
fs.String(
	"config",
	"c",
	"default.conf",
	"config file",
).Env("CONFIG_FILE")
```

CLI:

```bash
./app --config cli.conf
```

Result:

```text
cli.conf
```

Without a CLI argument:

```bash
CONFIG_FILE=env.conf ./app
```

Result:

```text
env.conf
```

Without either:

```text
default.conf
```

## CLI Syntax

### Long form

```bash
./app --output result.txt
```

### Long form with `=`

```bash
./app --output=result.txt
```

### Short form

```bash
./app -o result.txt
```

### Nested keys

```bash
./app \
	--server.host localhost \
	--server.port 8080
```

### Boolean flags

```bash
./app --verbose
./app --debug=true
./app --debug=false
```

### Repeatable string flags

```bash
./app \
	-t important \
	-t urgent \
	-t review
```

## Nested Configuration

Keys can use dot notation:

```go
fs.String(
	"server.host",
	"",
	"localhost",
	"server host",
)

fs.Int(
	"server.port",
	"",
	8080,
	"server port",
)

fs.String(
	"database.host",
	"",
	"localhost",
	"database host",
)
```

They can then be accessed directly:

```go
host := cfg.String("server.host")
port := cfg.Int("server.port")
dbHost := cfg.String("database.host")
```

This is useful for keeping application configuration organized without introducing a separate configuration hierarchy.

## Environment Variables

### Custom Environment Variable Names

Use `Env()` when you want to explicitly choose the environment variable name:

```go
fs.String(
	"database",
	"d",
	"sqlite.db",
	"database file",
).Env("DB_PATH")
```

This reads:

```text
DB_PATH
```

### Environment Variable Prefixes

Use `EnvWithPrefix()` when multiple flags belong to the same application or configuration namespace:

```go
fs.String(
	"server.host",
	"",
	"localhost",
	"server host",
).EnvWithPrefix("APP")

fs.Int(
	"server.port",
	"",
	8080,
	"server port",
).EnvWithPrefix("APP")
```

The resulting environment variables are:

```text
APP_SERVER__HOST
APP_SERVER__PORT
```

Dots in nested keys are converted to double underscores.

For example:

```text
server.host
```

becomes:

```text
APP_SERVER__HOST
```

And:

```text
database.connection.host
```

becomes:

```text
APP_DATABASE__CONNECTION__HOST
```

The generated environment variable is also shown in the help output.

## Global Environment Prefix

A prefix can be applied to an entire flag set:

```go
fs := varg.New("myapp")

fs.GlobalEnvPrefix("MYAPP")

fs.String(
	"server.host",
	"",
	"localhost",
	"server host",
)

fs.Int(
	"server.port",
	"",
	8080,
	"server port",
)
```

This allows environment configuration to follow a consistent application-wide naming scheme.

## Supported Types

| Type       | Definition      | Getter              |
| ---------- | --------------- | ------------------- |
| `string`   | `String()`      | `cfg.String()`      |
| `bool`     | `Bool()`        | `cfg.Bool()`        |
| `[]string` | `StringSlice()` | `cfg.StringSlice()` |
| `int`      | `Int()`         | `cfg.Int()`         |
| `int8`     | `Int8()`        | `cfg.Int8()`        |
| `int16`    | `Int16()`       | `cfg.Int16()`       |
| `int32`    | `Int32()`       | `cfg.Int32()`       |
| `int64`    | `Int64()`       | `cfg.Int64()`       |
| `uint`     | `Uint()`        | `cfg.Uint()`        |
| `uint8`    | `Uint8()`       | `cfg.Uint8()`       |
| `uint16`   | `Uint16()`      | `cfg.Uint16()`      |
| `uint32`   | `Uint32()`      | `cfg.Uint32()`      |
| `uint64`   | `Uint64()`      | `cfg.Uint64()`      |
| `float32`  | `Float32()`     | `cfg.Float32()`     |
| `float64`  | `Float64()`     | `cfg.Float64()`     |

Unsigned integer flags reject negative values during parsing.

Numeric getters perform the appropriate type conversion and return the zero value when the requested type cannot represent the stored value.

## String Slices

`StringSlice()` creates a repeatable string flag:

```go
fs.StringSlice(
	"tag",
	"t",
	"tags to apply",
)
```

Use it multiple times:

```bash
./app -t important -t urgent -t review
```

The resulting value is:

```go
[]string{
	"important",
	"urgent",
	"review",
}
```

Access it with:

```go
tags := cfg.StringSlice("tag")
```

## Struct Unmarshaling

For applications with larger configuration structures, parsed values can be unmarshaled into a Go struct.

```go
type Config struct {
	Server struct {
		Host string `flag:"server.host"`
		Port int    `flag:"server.port"`
	}

	Database struct {
		Host string `flag:"database.host"`
		Port int    `flag:"database.port"`
	}

	Debug bool `flag:"debug"`
}
```

Parse the arguments:

```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil {
	log.Fatal(err)
}
```

Then unmarshal:

```go
var config Config

if err := cfg.Unmarshal(&config); err != nil {
	log.Fatal(err)
}
```

The resulting configuration can be used normally:

```go
fmt.Println(config.Server.Host)
fmt.Println(config.Server.Port)
fmt.Println(config.Database.Host)
fmt.Println(config.Debug)
```

## Built-in Help and Version

Help is enabled by default:

```bash
./app --help
```

or:

```bash
./app -h
```

A version can be configured with:

```go
fs.Version("1.2.3")
```

This enables:

```bash
./app --version
```

and:

```bash
./app -v
```

### Handling Help and Version

For applications that want explicit control over the exit behavior, use `Parse()`:

```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil {
	if errors.Is(err, varg.ErrHelp) {
		fmt.Print(fs.Usage())
		os.Exit(0)
	}

	if errors.Is(err, varg.ErrVersion) {
		fmt.Println("myapp 1.2.3")
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
```

Alternatively, `Handle()` packages this behavior into a single result:

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
```

`Result` contains:

| Field        | Description                         |
| ------------ | ----------------------------------- |
| `Config`     | Parsed configuration                |
| `ShouldExit` | Whether the caller should terminate |
| `Output`     | Help, version, or error output      |
| `Err`        | Parsing error, if any               |
| `Version`    | Configured version string           |
| `Help`       | Generated help text                 |

## Disabling Built-in Flags

If an application does not want the automatic help flag:

```go
fs.DisableHelp()
```

To disable version handling:

```go
fs.DisableVersion()
```

Both methods are chainable:

```go
fs.
	DisableHelp().
	DisableVersion()
```

## Counting Flags

Integer flags support repeated short options as a counter.

Given:

```go
fs.Int(
	"verbose",
	"v",
	0,
	"verbosity level",
)
```

These are equivalent:

```bash
./app -v
```

```text
verbose = 1
```

```bash
./app -vv
```

```text
verbose = 2
```

```bash
./app -vvv
```

```text
verbose = 3
```

This is useful for verbosity or debug levels:

```bash
./app -vvv
```

## Increment and Decrement Shorthand

A `+` prefix can be used with integer flags to represent a negative adjustment:

```go
fs.Int(
	"priority",
	"p",
	0,
	"priority",
)
```

Then:

```bash
./app +p
```

sets the value to:

```text
-1
```

This can be useful for counter-style options.

## Reading Configuration

After parsing, `Config` provides typed getters:

```go
host := cfg.String("server.host")
port := cfg.Int("server.port")
debug := cfg.Bool("debug")
threshold := cfg.Float64("threshold")
tags := cfg.StringSlice("tag")
```

For a raw value:

```go
value, ok := cfg.Get("server.host")
```

To retrieve all parsed values:

```go
values := cfg.All()
```

## API Overview

### FlagSet

Create a flag set:

```go
fs := varg.New("myapp")
```

Define flags:

```go
fs.String("output", "o", "output.txt", "output file")
fs.Int("port", "p", 8080, "listen port")
fs.Bool("debug", "d", false, "enable debug mode")
fs.Float64("threshold", "t", 0.5, "threshold")
fs.StringSlice("tag", "", "tags")
```

Integer variants are also available:

```go
fs.Int8(...)
fs.Int16(...)
fs.Int32(...)
fs.Int64(...)

fs.Uint(...)
fs.Uint8(...)
fs.Uint16(...)
fs.Uint32(...)
fs.Uint64(...)
```

Floating-point variants:

```go
fs.Float32(...)
fs.Float64(...)
```

Configuration helpers:

```go
fs.GlobalEnvPrefix("APP")
fs.Version("1.0.0")
fs.DisableHelp()
fs.DisableVersion()
```

Generate help text:

```go
help := fs.Usage()
```

Parse arguments:

```go
cfg, err := fs.Parse(os.Args[1:])
```

Or use the higher-level handler:

```go
res := fs.Handle(os.Args[1:])
```

### Flag

Every flag definition returns a `*Flag`, allowing environment configuration to be chained:

```go
fs.String(
	"output",
	"o",
	"output.log",
	"output file",
).Env("OUTPUT_FILE")
```

Or:

```go
fs.String(
	"server.host",
	"",
	"localhost",
	"server host",
).EnvWithPrefix("APP")
```

### Config

Available getters include:

```go
cfg.Get(...)
cfg.String(...)
cfg.Bool(...)
cfg.StringSlice(...)

cfg.Int(...)
cfg.Int8(...)
cfg.Int16(...)
cfg.Int32(...)
cfg.Int64(...)

cfg.Uint(...)
cfg.Uint8(...)
cfg.Uint16(...)
cfg.Uint32(...)
cfg.Uint64(...)

cfg.Float32(...)
cfg.Float64(...)

cfg.All()
cfg.Unmarshal(...)
```

## Complete Example

```go
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/fmotalleb/varg"
)

type Config struct {
	Server struct {
		Host string `flag:"server.host"`
		Port int    `flag:"server.port"`
	}

	Database struct {
		Host string `flag:"database.host"`
		Port int    `flag:"database.port"`
	}

	Debug   bool     `flag:"debug"`
	Verbose int      `flag:"verbose"`
	Tags    []string `flag:"tag"`
}

func main() {
	fs := varg.New("myapp")

	fs.Version("1.0.0")

	fs.String(
		"server.host",
		"",
		"localhost",
		"server host",
	).EnvWithPrefix("APP")

	fs.Int(
		"server.port",
		"",
		8080,
		"server port",
	).EnvWithPrefix("APP")

	fs.String(
		"database.host",
		"",
		"localhost",
		"database host",
	).EnvWithPrefix("APP")

	fs.Int(
		"database.port",
		"",
		5432,
		"database port",
	).EnvWithPrefix("APP")

	fs.Bool(
		"debug",
		"d",
		false,
		"enable debug mode",
	)

	fs.Int(
		"verbose",
		"v",
		0,
		"verbosity level",
	)

	fs.StringSlice(
		"tag",
		"t",
		"application tag",
	)

	cfg, err := fs.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, varg.ErrHelp) {
			fmt.Print(fs.Usage())
			return
		}

		if errors.Is(err, varg.ErrVersion) {
			fmt.Println("myapp 1.0.0")
			return
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var config Config

	if err := cfg.Unmarshal(&config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"server: %s:%d\n",
		config.Server.Host,
		config.Server.Port,
	)

	fmt.Printf(
		"database: %s:%d\n",
		config.Database.Host,
		config.Database.Port,
	)

	fmt.Printf("debug: %v\n", config.Debug)
	fmt.Printf("verbosity: %d\n", config.Verbose)
	fmt.Printf("tags: %v\n", config.Tags)
}
```

Example:

```bash
./myapp \
	--server.host 127.0.0.1 \
	--server.port 9000 \
	--database.host postgres \
	--database.port 5432 \
	--debug \
	-vvv \
	-t production \
	-t api
```

Equivalent environment configuration:

```bash
APP_SERVER__HOST=127.0.0.1 \
APP_SERVER__PORT=9000 \
APP_DATABASE__HOST=postgres \
APP_DATABASE__PORT=5432 \
./myapp --debug -vvv
```

## Design Philosophy

`varg` intentionally focuses on configuration parsing rather than becoming a complete CLI framework.

The core design goals are:

* Minimal API
* Small dependency footprint
* Simple initialization
* Type-specific accessors
* First-class environment configuration
* Nested configuration without a separate schema
* CLI/environment/default precedence
* Chainable configuration
* Easy integration with application-specific configuration structs

If your application needs subcommands such as:

```text
app serve
app migrate
app config
```

`varg` is not intended to provide that functionality. A command framework such as Cobra is a better fit for that use case.

`varg` is intended for applications where the primary concern is:

```text
CLI flags + environment variables + defaults
```

## Testing

Run the test suite with:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

The test suite covers flag parsing, CLI syntax, environment variables, precedence, nested keys, struct unmarshaling, repeatable flags, and error handling.

## License

MIT License

See [LICENSE](LICENSE) for the complete license text.
