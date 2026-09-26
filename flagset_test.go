package varg

import (
	"errors"
	"strings"
	"testing"
)

func TestBasicStringFlag(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output file")

	cfg, err := fs.Parse([]string{"--output", "custom.txt"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "custom.txt" {
		t.Errorf("expected 'custom.txt', got '%s'", got)
	}
}

func TestShortFlag(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output file")

	cfg, err := fs.Parse([]string{"-o", "short.txt"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "short.txt" {
		t.Errorf("expected 'short.txt', got '%s'", got)
	}
}

func TestDefaultValue(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output file")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "default.txt" {
		t.Errorf("expected 'default.txt', got '%s'", got)
	}
}

func TestIntFlag(t *testing.T) {
	fs := New("test")
	fs.Int("port", "p", 8080, "listen port")

	cfg, err := fs.Parse([]string{"--port", "9000"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Int("port"); got != 9000 {
		t.Errorf("expected 9000, got %d", got)
	}
}

func TestBoolFlag(t *testing.T) {
	fs := New("test")
	fs.Bool("verbose", "v", false, "verbose output")

	cfg, err := fs.Parse([]string{"--verbose"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Bool("verbose"); !got {
		t.Errorf("expected true, got %v", got)
	}
}

func TestBoolFlagExplicit(t *testing.T) {
	fs := New("test")
	fs.Bool("debug", "d", false, "debug mode")

	cfg, err := fs.Parse([]string{"--debug=false"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Bool("debug"); got {
		t.Errorf("expected false, got %v", got)
	}
}

func TestFloat64Flag(t *testing.T) {
	fs := New("test")
	fs.Float64("threshold", "t", 0.5, "threshold value")

	cfg, err := fs.Parse([]string{"--threshold", "0.75"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Float64("threshold"); got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}
}

func TestStringSliceFlag(t *testing.T) {
	fs := New("test")
	fs.StringSlice("tag", "t", "tags")

	cfg, err := fs.Parse([]string{"-t", "foo", "-t", "bar", "-t", "baz"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	tags := cfg.StringSlice("tag")
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "foo" || tags[1] != "bar" || tags[2] != "baz" {
		t.Errorf("unexpected tag values: %v", tags)
	}
}

func TestEnvironmentVariable(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output").Env("OUTPUT_FILE")

	t.Setenv("OUTPUT_FILE", "env.txt")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "env.txt" {
		t.Errorf("expected 'env.txt', got '%s'", got)
	}
}

func TestCLIPrecedenceOverEnv(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output").Env("OUTPUT_FILE")

	t.Setenv("OUTPUT_FILE", "env.txt")

	cfg, err := fs.Parse([]string{"--output", "cli.txt"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "cli.txt" {
		t.Errorf("expected 'cli.txt', got '%s'", got)
	}
}

func TestNestedKeyWithEnvPrefix(t *testing.T) {
	fs := New("test")
	fs.String("server.host", "", "localhost", "server host").EnvWithPrefix("APP")

	t.Setenv("APP_SERVER__HOST", "example.com")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("server.host"); got != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", got)
	}
}

func TestNestedKeyMultiLevel(t *testing.T) {
	fs := New("test")
	fs.String("db.server.host", "", "localhost", "db host").EnvWithPrefix("CONFIG")
	fs.Int("db.server.port", "", 5432, "db port").EnvWithPrefix("CONFIG")

	t.Setenv("CONFIG_DB__SERVER__HOST", "postgres.example.com")
	t.Setenv("CONFIG_DB__SERVER__PORT", "5433")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("db.server.host"); got != "postgres.example.com" {
		t.Errorf("expected 'postgres.example.com', got '%s'", got)
	}

	if got := cfg.Int("db.server.port"); got != 5433 {
		t.Errorf("expected 5433, got %d", got)
	}
}

func TestUnmarshalSimpleStruct(t *testing.T) {
	fs := New("test")
	fs.String("host", "", "localhost", "host").Env("HOST")
	fs.Int("port", "", 8080, "port").Env("PORT")

	cfg, err := fs.Parse([]string{"--host", "example.com", "--port", "9000"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var config struct {
		Host string `flag:"host"`
		Port int    `flag:"port"`
	}

	if err := cfg.Unmarshal(&config); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if config.Host != "example.com" {
		t.Errorf("expected 'example.com', got '%s'", config.Host)
	}

	if config.Port != 9000 {
		t.Errorf("expected 9000, got %d", config.Port)
	}
}

func TestUnmarshalNestedStruct(t *testing.T) {
	fs := New("test")
	fs.String("server.host", "", "localhost", "server host").Env("SERVER_HOST")
	fs.Int("server.port", "", 8080, "server port").Env("SERVER_PORT")
	fs.String("db.host", "", "localhost", "db host").Env("DB_HOST")
	fs.Int("db.port", "", 5432, "db port").Env("DB_PORT")

	cfg, err := fs.Parse([]string{
		"--server.host", "api.example.com",
		"--server.port", "9000",
		"--db.host", "db.example.com",
		"--db.port", "5433",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var config struct {
		Server struct {
			Host string `flag:"server.host"`
			Port int    `flag:"server.port"`
		}
		DB struct {
			Host string `flag:"db.host"`
			Port int    `flag:"db.port"`
		}
	}

	if err := cfg.Unmarshal(&config); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if config.Server.Host != "api.example.com" {
		t.Errorf("expected 'api.example.com', got '%s'", config.Server.Host)
	}

	if config.Server.Port != 9000 {
		t.Errorf("expected 9000, got %d", config.Server.Port)
	}

	if config.DB.Host != "db.example.com" {
		t.Errorf("expected 'db.example.com', got '%s'", config.DB.Host)
	}

	if config.DB.Port != 5433 {
		t.Errorf("expected 5433, got %d", config.DB.Port)
	}
}

func TestUnknownFlag(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output")

	_, err := fs.Parse([]string{"--unknown", "value"})
	if err == nil {
		t.Errorf("expected error for unknown flag, got nil")
	}
}

func TestInvalidIntValue(t *testing.T) {
	fs := New("test")
	fs.Int("port", "p", 8080, "port")

	_, err := fs.Parse([]string{"--port", "not-a-number"})
	if err == nil {
		t.Errorf("expected error for invalid int, got nil")
	}
}

func TestEqualsSyntax(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output")

	cfg, err := fs.Parse([]string{"--output=equals.txt"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("output"); got != "equals.txt" {
		t.Errorf("expected 'equals.txt', got '%s'", got)
	}
}

func TestMultipleFlags(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "out.txt", "output")
	fs.String("input", "i", "in.txt", "input")
	fs.Int("count", "c", 1, "count")
	fs.Bool("verbose", "v", false, "verbose")

	cfg, err := fs.Parse([]string{
		"-i", "custom-in.txt",
		"--output", "custom-out.txt",
		"--count", "42",
		"-v",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.String("input"); got != "custom-in.txt" {
		t.Errorf("input: expected 'custom-in.txt', got '%s'", got)
	}

	if got := cfg.String("output"); got != "custom-out.txt" {
		t.Errorf("output: expected 'custom-out.txt', got '%s'", got)
	}

	if got := cfg.Int("count"); got != 42 {
		t.Errorf("count: expected 42, got %d", got)
	}

	if got := cfg.Bool("verbose"); !got {
		t.Errorf("verbose: expected true, got %v", got)
	}
}

func TestAllValues(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output")
	fs.Int("port", "p", 8080, "port")

	cfg, err := fs.Parse([]string{"--output", "test.txt", "--port", "9000"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	all := cfg.All()
	if len(all) != 2 {
		t.Errorf("expected 2 values, got %d", len(all))
	}

	if all["output"] != "test.txt" {
		t.Errorf("output mismatch in All()")
	}

	if all["port"] != 9000 {
		t.Errorf("port mismatch in All()")
	}
}

func TestGetMethod(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output")

	cfg, err := fs.Parse([]string{"--output", "custom.txt"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	val, ok := cfg.Get("output")
	if !ok {
		t.Errorf("Get: key not found")
	}

	if val != "custom.txt" {
		t.Errorf("Get: expected 'custom.txt', got '%v'", val)
	}
}

func TestNonExistentKey(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	val, ok := cfg.Get("nonexistent")
	if ok {
		t.Errorf("Get: expected not found for nonexistent key, but got: %v", val)
	}

	if got := cfg.String("nonexistent"); got != "" {
		t.Errorf("String: expected empty string, got '%s'", got)
	}
}

func TestHelpFlag(t *testing.T) {
	fs := New("test")
	fs.String("output", "o", "default.txt", "output file")

	_, err := fs.Parse([]string{"--help"})
	if err == nil || err.Error() != "__HELP__" {
		t.Errorf("expected __HELP__ error, got %v", err)
	}

	_, err = fs.Parse([]string{"-h"})
	if err == nil || err.Error() != "__HELP__" {
		t.Errorf("expected __HELP__ error with -h, got %v", err)
	}
}

func TestVersionFlag(t *testing.T) {
	fs := New("test")
	fs.Version("1.0.0")

	_, err := fs.Parse([]string{"--version"})
	if err == nil || err.Error() != "__VERSION__" {
		t.Errorf("expected __VERSION__ error, got %v", err)
	}

	_, err = fs.Parse([]string{"-v"})
	if err == nil || err.Error() != "__VERSION__" {
		t.Errorf("expected __VERSION__ error with -v, got %v", err)
	}
}

func TestDisableHelp(t *testing.T) {
	fs := New("test")
	fs.DisableHelp()
	fs.String("output", "o", "default.txt", "output file")

	_, err := fs.Parse([]string{"--help"})
	if err == nil {
		t.Errorf("expected error for --help when disabled, got nil")
	}
}

func TestDisableVersion(t *testing.T) {
	fs := New("test")
	fs.DisableVersion()
	fs.Version("1.0.0")

	_, err := fs.Parse([]string{"--version"})
	if err == nil {
		t.Errorf("expected error for --version when disabled, got nil")
	}
}

func TestNumeralShorthand(t *testing.T) {
	fs := New("test")
	fs.Int("verbose", "v", 0, "verbosity level")
	fs.DisableHelp().DisableVersion() // Disable built-in flags to test -vvv

	cfg, err := fs.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Int("verbose"); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestNumeralShorthandDouble(t *testing.T) {
	fs := New("test")
	fs.Int("debug", "d", 0, "debug level")

	cfg, err := fs.Parse([]string{"-dd"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Int("debug"); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestIncrementDecrement(t *testing.T) {
	fs := New("test")
	fs.Int("count", "c", 0, "counter")

	cfg, err := fs.Parse([]string{"+c"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if got := cfg.Int("count"); got != -1 {
		t.Errorf("expected -1 (decrement), got %d", got)
	}
}

func TestShortFlagOverridesBuiltInVersion(t *testing.T) {
	fs := New("test")
	fs.Version("1.0.0")
	fs.Int("verbose", "v", 0, "verbosity level")

	cfg, err := fs.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := cfg.Int("verbose"); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}

	// The long name still belongs to the built-in version flag.
	_, err = fs.Parse([]string{"--version"})
	if err == nil || err.Error() != "__VERSION__" {
		t.Errorf("expected __VERSION__ error, got %v", err)
	}
}

func TestLongFlagOverridesBuiltInVersion(t *testing.T) {
	fs := New("test")
	fs.Version("1.0.0")
	fs.String("version", "", "app", "version source")

	cfg, err := fs.Parse([]string{"--version", "cli"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := cfg.String("version"); got != "cli" {
		t.Errorf("expected 'cli', got '%s'", got)
	}

	// The short name still belongs to the built-in version flag.
	_, err = fs.Parse([]string{"-v"})
	if err == nil || err.Error() != "__VERSION__" {
		t.Errorf("expected __VERSION__ error, got %v", err)
	}
}

func TestShortFlagOverridesBuiltInHelp(t *testing.T) {
	fs := New("test")
	fs.Bool("human", "h", false, "human readable output")

	cfg, err := fs.Parse([]string{"-h"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !cfg.Bool("human") {
		t.Errorf("expected human readable output to be enabled")
	}

	_, err = fs.Parse([]string{"--help"})
	if err == nil || err.Error() != "__HELP__" {
		t.Errorf("expected __HELP__ error, got %v", err)
	}
}

func TestDisableHelpKeepsOverriddenShort(t *testing.T) {
	fs := New("test")
	fs.Bool("human", "h", false, "human readable output")
	fs.DisableHelp()

	cfg, err := fs.Parse([]string{"-h"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !cfg.Bool("human") {
		t.Errorf("expected human readable output to be enabled")
	}

	if _, err = fs.Parse([]string{"--help"}); err == nil {
		t.Errorf("expected error for --help when disabled")
	}
}

func TestRepeatedShortMatchesNumeralShorthand(t *testing.T) {
	fs := New("test")
	fs.Version("1.0.0")
	fs.Int("verbose", "v", 0, "verbosity level")

	repeated, err := fs.Parse([]string{"-v", "-v", "-v"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	numeral, err := fs.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if repeated.Int("verbose") != numeral.Int("verbose") || repeated.Int("verbose") != 3 {
		t.Errorf("expected both forms to give 3, got %d and %d", repeated.Int("verbose"), numeral.Int("verbose"))
	}
}

func TestUsage(t *testing.T) {
	fs := New("myapp")
	fs.Version("1.0.0")
	fs.String("output", "o", "output.txt", "output file")
	fs.Int("port", "p", 8080, "listen port")

	usage := fs.Usage()
	if !strings.Contains(usage, "myapp") {
		t.Errorf("usage should contain app name")
	}
	if !strings.Contains(usage, "--help") {
		t.Errorf("usage should contain help flag")
	}
	if !strings.Contains(usage, "--version") {
		t.Errorf("usage should contain version flag")
	}
	if !strings.Contains(usage, "output file") {
		t.Errorf("usage should contain flag help text")
	}
}

func TestNumericFlagTypes(t *testing.T) {
	tests := []struct {
		name string
		add  func(fs *FlagSet)
		get  func(cfg *Config) interface{}
		args []string
		want interface{}
	}{
		{"int", func(fs *FlagSet) { fs.Int("v", "", 0, "") }, func(c *Config) interface{} { return c.Int("v") }, []string{"--v", "-42"}, int(-42)},
		{"int8", func(fs *FlagSet) { fs.Int8("v", "", 0, "") }, func(c *Config) interface{} { return c.Int8("v") }, []string{"--v", "-42"}, int8(-42)},
		{"int16", func(fs *FlagSet) { fs.Int16("v", "", 0, "") }, func(c *Config) interface{} { return c.Int16("v") }, []string{"--v", "-30000"}, int16(-30000)},
		{"int32", func(fs *FlagSet) { fs.Int32("v", "", 0, "") }, func(c *Config) interface{} { return c.Int32("v") }, []string{"--v", "-70000"}, int32(-70000)},
		{"int64", func(fs *FlagSet) { fs.Int64("v", "", 0, "") }, func(c *Config) interface{} { return c.Int64("v") }, []string{"--v", "1099511627776"}, int64(1099511627776)},
		{"uint", func(fs *FlagSet) { fs.Uint("v", "", 0, "") }, func(c *Config) interface{} { return c.Uint("v") }, []string{"--v", "42"}, uint(42)},
		{"uint8", func(fs *FlagSet) { fs.Uint8("v", "", 0, "") }, func(c *Config) interface{} { return c.Uint8("v") }, []string{"--v", "255"}, uint8(255)},
		{"uint16", func(fs *FlagSet) { fs.Uint16("v", "", 0, "") }, func(c *Config) interface{} { return c.Uint16("v") }, []string{"--v", "65535"}, uint16(65535)},
		{"uint32", func(fs *FlagSet) { fs.Uint32("v", "", 0, "") }, func(c *Config) interface{} { return c.Uint32("v") }, []string{"--v", "4000000000"}, uint32(4000000000)},
		{"uint64", func(fs *FlagSet) { fs.Uint64("v", "", 0, "") }, func(c *Config) interface{} { return c.Uint64("v") }, []string{"--v", "1152921504606846976"}, uint64(1152921504606846976)},
		{"float32", func(fs *FlagSet) { fs.Float32("v", "", 0, "") }, func(c *Config) interface{} { return c.Float32("v") }, []string{"--v", "1.5"}, float32(1.5)},
		{"float64", func(fs *FlagSet) { fs.Float64("v", "", 0, "") }, func(c *Config) interface{} { return c.Float64("v") }, []string{"--v", "2.25"}, float64(2.25)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := New("test")
			tc.add(fs)

			cfg, err := fs.Parse(tc.args)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if got, ok := cfg.Get("v"); !ok || got != tc.want {
				t.Errorf("Get: got %v (%T), want %v (%T)", got, got, tc.want, tc.want)
			}
			if got := tc.get(cfg); got != tc.want {
				t.Errorf("getter: got %v (%T), want %v (%T)", got, got, tc.want, tc.want)
			}
		})
	}
}

func TestNumericOutOfRange(t *testing.T) {
	fs := New("test")
	fs.Int8("small", "", 0, "")
	if _, err := fs.Parse([]string{"--small", "300"}); err == nil {
		t.Errorf("expected error for int8 out of range")
	}
	if _, err := fs.Parse([]string{"--small", "42"}); err != nil {
		t.Errorf("parse failed: %v", err)
	}

	fs = New("test")
	fs.Uint8("byte", "", 0, "")
	if _, err := fs.Parse([]string{"--byte", "256"}); err == nil {
		t.Errorf("expected error for uint8 out of range")
	}
	if _, err := fs.Parse([]string{"--byte", "-1"}); err == nil {
		t.Errorf("expected error for negative uint8")
	}
}

func TestNumericNegativeValue(t *testing.T) {
	fs := New("test")
	fs.Int("port", "p", 8080, "")
	fs.Int64("delta", "", 0, "")

	cfg, err := fs.Parse([]string{"--port", "-1", "--delta", "-5"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if cfg.Int("port") != -1 {
		t.Errorf("port: expected -1, got %d", cfg.Int("port"))
	}
	if cfg.Int64("delta") != -5 {
		t.Errorf("delta: expected -5, got %d", cfg.Int64("delta"))
	}
}

func TestNumericShorthandAcrossTypes(t *testing.T) {
	fs := New("test")
	fs.Int64("verbose", "v", 0, "")
	fs.Uint8("debug", "d", 0, "")
	fs.Float32("factor", "f", 0, "")
	fs.Uint16("retries", "r", 10, "")

	cfg, err := fs.Parse([]string{"-v", "-v", "-vv", "-d", "-d", "-f", "-f", "-r"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if cfg.Int64("verbose") != 4 {
		t.Errorf("verbose: expected 4, got %d", cfg.Int64("verbose"))
	}
	if cfg.Uint8("debug") != 2 {
		t.Errorf("debug: expected 2, got %d", cfg.Uint8("debug"))
	}
	if cfg.Float32("factor") != 2 {
		t.Errorf("factor: expected 2, got %v", cfg.Float32("factor"))
	}
	if cfg.Uint16("retries") != 11 {
		t.Errorf("retries: expected 11 (default 10 + 1), got %d", cfg.Uint16("retries"))
	}
}

func TestUnsignedCannotBeDecremented(t *testing.T) {
	fs := New("test")
	fs.Uint("count", "c", 0, "")

	if _, err := fs.Parse([]string{"+c"}); err == nil {
		t.Errorf("expected error when decrementing a uint flag")
	}
}

func TestUnmarshalNumericTypes(t *testing.T) {
	fs := New("test")
	fs.Int64("id", "", int64(0), "")
	fs.Uint16("mask", "", uint16(0), "")
	fs.Float32("ratio", "", float32(0), "")
	fs.Int("port", "", 0, "")

	cfg, err := fs.Parse([]string{
		"--id", "1099511627776",
		"--mask", "65535",
		"--ratio", "0.25",
		"--port", "8080",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	var s struct {
		ID    int64   `flag:"id"`
		Mask  uint16  `flag:"mask"`
		Ratio float32 `flag:"ratio"`
		Port  int64   `flag:"port"`
	}
	if err := cfg.Unmarshal(&s); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if s.ID != 1099511627776 {
		t.Errorf("id: expected 1099511627776, got %d", s.ID)
	}
	if s.Mask != 65535 {
		t.Errorf("mask: expected 65535, got %d", s.Mask)
	}
	if s.Ratio != 0.25 {
		t.Errorf("ratio: expected 0.25, got %v", s.Ratio)
	}
	if s.Port != 8080 {
		t.Errorf("port: expected 8080, got %d", s.Port)
	}

	var tooSmall struct {
		Port int8 `flag:"port"`
	}
	if err := cfg.Unmarshal(&tooSmall); err == nil {
		t.Errorf("expected out of range error for int8 field")
	}
}

func TestHandleHelp(t *testing.T) {
	fs := New("myapp")
	fs.String("output", "o", "out.txt", "output file")

	res := fs.Handle([]string{"--help"})
	if !res.ShouldExit {
		t.Errorf("expected ShouldExit")
	}
	if res.Err != nil {
		t.Errorf("expected no error for help, got %v", res.Err)
	}
	if res.Config != nil {
		t.Errorf("expected nil config for help")
	}
	if res.Output != res.Help {
		t.Errorf("expected output to be the help text")
	}
	if !strings.Contains(res.Output, "Usage: myapp") {
		t.Errorf("help output missing usage line: %q", res.Output)
	}
}

func TestHandleVersion(t *testing.T) {
	fs := New("myapp")
	fs.Version("1.2.3")

	res := fs.Handle([]string{"-v"})
	if !res.ShouldExit {
		t.Errorf("expected ShouldExit")
	}
	if res.Err != nil {
		t.Errorf("expected no error for version, got %v", res.Err)
	}
	if res.Output != "myapp 1.2.3\n" {
		t.Errorf("version output: got %q", res.Output)
	}
	if res.Version != "1.2.3" {
		t.Errorf("expected retrievable version 1.2.3, got %q", res.Version)
	}
}

func TestHandleSuccess(t *testing.T) {
	fs := New("myapp")
	fs.Version("1.2.3")
	fs.Int("port", "p", 8080, "listen port")

	res := fs.Handle([]string{"--port", "9000"})
	if res.ShouldExit {
		t.Errorf("expected ShouldExit to be false")
	}
	if res.Err != nil {
		t.Errorf("expected no error, got %v", res.Err)
	}
	if res.Config == nil {
		t.Fatalf("expected a config")
	}
	if res.Config.Int("port") != 9000 {
		t.Errorf("port: expected 9000, got %d", res.Config.Int("port"))
	}
	if res.Version != "1.2.3" {
		t.Errorf("expected retrievable version, got %q", res.Version)
	}
	if res.Help == "" {
		t.Errorf("expected help text to be available")
	}
}

func TestHandleError(t *testing.T) {
	fs := New("myapp")
	fs.String("output", "o", "out.txt", "output file")

	res := fs.Handle([]string{"--nope", "x"})
	if !res.ShouldExit {
		t.Errorf("expected ShouldExit")
	}
	if res.Err == nil {
		t.Errorf("expected an error")
	}
	if res.Config != nil {
		t.Errorf("expected nil config on failure")
	}
	if !strings.Contains(res.Output, "unknown flag") {
		t.Errorf("output should carry the error message, got %q", res.Output)
	}
}

func TestSentinelErrors(t *testing.T) {
	fs := New("test")
	fs.Version("1.0.0")

	if _, err := fs.Parse([]string{"--help"}); !errors.Is(err, ErrHelp) {
		t.Errorf("expected ErrHelp, got %v", err)
	}
	if _, err := fs.Parse([]string{"-v"}); !errors.Is(err, ErrVersion) {
		t.Errorf("expected ErrVersion, got %v", err)
	}
}

func TestUsageShowsEnvVars(t *testing.T) {
	fs := New("myapp")
	fs.String("output", "o", "out.txt", "output file").Env("OUTPUT_FILE")
	fs.Int("server.port", "", 8080, "server port").EnvWithPrefix("APP")
	fs.String("plain", "", "x", "no env here")

	usage := fs.Usage()
	if !strings.Contains(usage, "output file [$OUTPUT_FILE]") {
		t.Errorf("usage should show the env var of a flag, got:\n%s", usage)
	}
	if !strings.Contains(usage, "server port [$APP_SERVER__PORT]") {
		t.Errorf("usage should show the prefixed env var, got:\n%s", usage)
	}
	if !strings.Contains(usage, "no env here\n") {
		t.Errorf("usage should keep help text of a flag without env var, got:\n%s", usage)
	}
	if strings.Contains(usage, "no env here [$") {
		t.Errorf("flag without env var must not show an env suffix, got:\n%s", usage)
	}
}
