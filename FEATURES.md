# New Features Added to varg

## 1. Version Configuration

Set a version string for your application:

```go
fs := varg.New("myapp")
fs.Version("1.2.3")

// Now --version and -v flags are automatically available
```

Access the version:
```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil && err.Error() == "__VERSION__" {
    fmt.Printf("myapp %s\n", fs.version)
    os.Exit(0)
}
```

## 2. Default Help Flag

Automatically available via `--help` or `-h`:

```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil && err.Error() == "__HELP__" {
    fmt.Print(fs.Usage())
    os.Exit(0)
}
```

The `Usage()` method returns a formatted help string:

```go
fmt.Print(fs.Usage())
// Output:
// Usage: myapp [options]
//
// Options:
//   -h, --help            show this help message
//   -v, --version         show version
//   -o, --output          output file
//   -p, --port            listen port
```

## 3. Disable Help and Version

Override built-in behavior:

```go
fs.DisableHelp()       // Disable --help(-h)
fs.DisableVersion()    // Disable --version(-v)
```

## 4. Numeral Shorthand (Counting)

Repeat short flags to count:

```go
fs.Int("verbose", "v", 0, "verbosity level")

// Usage:
// ./app -vvv          # verbose = 3
// ./app -vv           # verbose = 2
// ./app -v            # verbose = 1
```

Useful for verbosity levels, debug depth, etc.

## 5. Increment/Decrement Shorthand

Use `+` prefix to set negative value:

```go
fs.Int("priority", "p", 0, "priority")

// Usage:
// ./app +p             # priority = -1
```

Useful for decrements or negative adjustments.

## API Changes

### New Methods on FlagSet

- `Version(v string) *FlagSet` — Set version string
- `DisableHelp() *FlagSet` — Disable automatic --help flag
- `DisableVersion() *FlagSet` — Disable automatic --version flag
- `Usage() string` — Get formatted help text

### Parse Error Handling

`Parse()` now returns special error values:

- `err.Error() == "__HELP__"` — User requested help
- `err.Error() == "__VERSION__"` — User requested version

Example handler:
```go
cfg, err := fs.Parse(os.Args[1:])
if err != nil {
    if err.Error() == "__HELP__" {
        fmt.Print(fs.Usage())
        os.Exit(0)
    }
    if err.Error() == "__VERSION__" {
        fmt.Printf("myapp %s\n", fs.version)
        os.Exit(0)
    }
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

## Examples

### Basic App with Version and Help

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
    
    fs.Int("verbose", "v", 0, "verbosity level")
    fs.String("output", "o", "output.txt", "output file")
    
    cfg, err := fs.Parse(os.Args[1:])
    if err != nil {
        if err.Error() == "__HELP__" {
            fmt.Print(fs.Usage())
            os.Exit(0)
        }
        if err.Error() == "__VERSION__" {
            fmt.Println("myapp 1.0.0")
            os.Exit(0)
        }
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Verbosity: %d\n", cfg.Int("verbose"))
    fmt.Printf("Output: %s\n", cfg.String("output"))
}
```

Usage:
```bash
$ ./myapp --help
Usage: myapp [options]

Options:
  -h, --help            show this help message
  -v, --version         show version
  -v, --verbose         verbosity level
  -o, --output          output file

$ ./myapp --version
myapp 1.0.0

$ ./myapp -vvv --output result.txt
Verbosity: 3
Output: result.txt
```

## Notes

- Built-in `-h, --help` and `-v, --version` flags are enabled by default
- If you override with your own `-v` or `-h` flag, the built-in will take precedence unless disabled
- Numeral shorthand only works for repeated identical characters: `-vvv` counts, but `-vd` doesn't
- Increment/decrement requires the `+` prefix: `+flag` sets value to -1
- Empty env vars now correctly skip to default (don't override defaults with empty strings)
