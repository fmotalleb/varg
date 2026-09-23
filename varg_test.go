package varg

import (
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
