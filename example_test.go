package varg

import (
	"testing"
)

// TestExampleBasicUsage demonstrates simple flag parsing.
func TestExampleBasicUsage(t *testing.T) {
	fs := New("myapp")
	fs.String("output", "o", "output.log", "output file")
	fs.Int("port", "p", 8080, "listen port")
	fs.Bool("verbose", "v", false, "verbose output")

	cfg, err := fs.Parse([]string{
		"--output", "custom.log",
		"--port", "9000",
		"-v",
	})
	if err != nil {
		t.Error(err)
	}
	if cfg.String("output") != "custom.log" {
		t.Errorf("output mismatch")
	}
	if cfg.Int("port") != 9000 {
		t.Errorf("port mismatch")
	}
	if !cfg.Bool("verbose") {
		t.Errorf("verbose mismatch")
	}
}

// TestExampleEnvironmentVariables demonstrates env var support with custom names.
func TestExampleEnvironmentVariables(t *testing.T) {
	fs := New("myapp")
	fs.String("output", "o", "output.log", "output file").Env("OUTPUT_FILE")
	fs.Int("port", "p", 8080, "listen port").Env("PORT")

	t.Setenv("OUTPUT_FILE", "env-output.log")
	t.Setenv("PORT", "3000")

	// No CLI args; values come from env
	cfg, _ := fs.Parse([]string{})

	if cfg.String("output") != "env-output.log" {
		t.Errorf("output from env mismatch")
	}
	if cfg.Int("port") != 3000 {
		t.Errorf("port from env mismatch")
	}
}

// TestExampleNestedKeys demonstrates dot notation and env var translation.
func TestExampleNestedKeys(t *testing.T) {
	fs := New("myapp")

	// Nested keys use dots; env vars use double underscores
	fs.String("server.host", "", "localhost", "server host").EnvWithPrefix("APP")
	fs.Int("server.port", "", 8080, "server port").EnvWithPrefix("APP")
	fs.String("db.host", "", "localhost", "db host").EnvWithPrefix("APP")
	fs.Int("db.port", "", 5432, "db port").EnvWithPrefix("APP")

	t.Setenv("APP_SERVER__HOST", "api.example.com")
	t.Setenv("APP_SERVER__PORT", "9000")

	cfg, _ := fs.Parse([]string{
		"--db.host", "postgres.example.com",
		"--db.port", "5433",
	})

	if cfg.String("server.host") != "api.example.com" {
		t.Errorf("server.host mismatch")
	}
	if cfg.Int("server.port") != 9000 {
		t.Errorf("server.port mismatch")
	}
	if cfg.String("db.host") != "postgres.example.com" {
		t.Errorf("db.host mismatch")
	}
	if cfg.Int("db.port") != 5433 {
		t.Errorf("db.port mismatch")
	}
}

// TestExampleStructUnmarshal demonstrates unmarshaling into structs with tags.
func TestExampleStructUnmarshal(t *testing.T) {
	fs := New("myapp")
	fs.String("server.host", "", "localhost", "server host")
	fs.Int("server.port", "", 8080, "server port")
	fs.String("db.host", "", "localhost", "db host")
	fs.Int("db.port", "", 5432, "db port")
	fs.Bool("debug", "", false, "debug mode")

	cfg, _ := fs.Parse([]string{
		"--server.host", "api.dev.local",
		"--server.port", "3000",
		"--db.host", "db.dev.local",
		"--debug",
	})

	var config struct {
		Server struct {
			Host string `flag:"server.host"`
			Port int    `flag:"server.port"`
		}
		DB struct {
			Host string `flag:"db.host"`
			Port int    `flag:"db.port"`
		}
		Debug bool `flag:"debug"`
	}

	if err := cfg.Unmarshal(&config); err != nil {
		t.Error(err)
	}

	if config.Server.Host != "api.dev.local" {
		t.Errorf("server.host mismatch")
	}
	if config.Server.Port != 3000 {
		t.Errorf("server.port mismatch")
	}
	if config.DB.Host != "db.dev.local" {
		t.Errorf("db.host mismatch")
	}
	if !config.Debug {
		t.Errorf("debug mismatch")
	}
}

// TestExampleStringSlice demonstrates repeatable flags.
func TestExampleStringSlice(t *testing.T) {
	fs := New("myapp")
	fs.StringSlice("tag", "t", "tags to apply")

	cfg, _ := fs.Parse([]string{
		"-t", "important",
		"-t", "urgent",
		"-t", "review",
	})

	tags := cfg.StringSlice("tag")
	if len(tags) != 3 || tags[0] != "important" {
		t.Errorf("tags mismatch")
	}
}

// TestExamplePrecedence demonstrates CLI > env > default precedence.
func TestExamplePrecedence(t *testing.T) {
	fs := New("myapp")
	fs.String("config", "c", "default.conf", "config file").Env("CONFIG_FILE")

	t.Setenv("CONFIG_FILE", "env.conf")

	// CLI argument takes precedence
	cfg, _ := fs.Parse([]string{"--config", "cli.conf"})
	if cfg.String("config") != "cli.conf" {
		t.Errorf("CLI precedence failed")
	}

	// Env var is used when no CLI arg
	cfg, _ = fs.Parse([]string{})
	if cfg.String("config") != "env.conf" {
		t.Errorf("env precedence failed")
	}

	// Default is used when nothing else
	t.Setenv("CONFIG_FILE", "")
	cfg, _ = fs.Parse([]string{})
	if cfg.String("config") != "default.conf" {
		t.Errorf("default precedence failed")
	}
}

// TestExampleVersion demonstrates version flag.
func TestExampleVersion(t *testing.T) {
	fs := New("myapp")
	fs.Version("1.2.3")
	fs.String("output", "o", "output.log", "output file")

	_, err := fs.Parse([]string{"--version"})
	if err == nil || err.Error() != "__VERSION__" {
		t.Errorf("expected __VERSION__ error, got %v", err)
	}
}

// TestExampleNumeralShorthand demonstrates -vvv shorthand.
func TestExampleNumeralShorthand(t *testing.T) {
	fs := New("myapp")
	fs.Int("verbose", "v", 0, "verbosity level")
	fs.DisableHelp().DisableVersion()

	cfg, err := fs.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if cfg.Int("verbose") != 3 {
		t.Errorf("expected 3, got %d", cfg.Int("verbose"))
	}
}

// TestExampleIncrement demonstrates increment/decrement with +flag.
func TestExampleIncrement(t *testing.T) {
	fs := New("myapp")
	fs.Int("priority", "p", 0, "priority level")

	cfg, err := fs.Parse([]string{"+p"})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if cfg.Int("priority") != -1 {
		t.Errorf("expected -1, got %d", cfg.Int("priority"))
	}
}
