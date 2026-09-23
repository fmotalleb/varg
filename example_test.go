package varg

import (
	"fmt"
	"os"
)

// ExampleBasicUsage demonstrates simple flag parsing.
func ExampleBasicUsage() {
	fs := New("myapp")
	fs.String("output", "o", "output.log", "output file")
	fs.Int("port", "p", 8080, "listen port")
	fs.Bool("verbose", "v", false, "verbose output")

	cfg, _ := fs.Parse([]string{
		"--output", "custom.log",
		"--port", "9000",
		"-v",
	})

	fmt.Println("Output:", cfg.String("output"))
	fmt.Println("Port:", cfg.Int("port"))
	fmt.Println("Verbose:", cfg.Bool("verbose"))
	// Output:
	// Output: custom.log
	// Port: 9000
	// Verbose: true
}

// ExampleEnvironmentVariables demonstrates env var support with custom names.
func ExampleEnvironmentVariables() {
	fs := New("myapp")
	fs.String("output", "o", "output.log", "output file").Env("OUTPUT_FILE")
	fs.Int("port", "p", 8080, "listen port").Env("PORT")

	os.Setenv("OUTPUT_FILE", "env-output.log")
	os.Setenv("PORT", "3000")

	// No CLI args; values come from env
	cfg, _ := fs.Parse([]string{})

	fmt.Println("Output:", cfg.String("output"))
	fmt.Println("Port:", cfg.Int("port"))
	// Output:
	// Output: env-output.log
	// Port: 3000
}

// ExampleNestedKeys demonstrates dot notation and env var translation.
func ExampleNestedKeys() {
	fs := New("myapp")

	// Nested keys use dots; env vars use double underscores
	fs.String("server.host", "", "localhost", "server host").EnvWithPrefix("APP")
	fs.Int("server.port", "", 8080, "server port").EnvWithPrefix("APP")
	fs.String("db.host", "", "localhost", "db host").EnvWithPrefix("APP")
	fs.Int("db.port", "", 5432, "db port").EnvWithPrefix("APP")

	os.Setenv("APP_SERVER__HOST", "api.example.com")
	os.Setenv("APP_SERVER__PORT", "9000")

	cfg, _ := fs.Parse([]string{
		"--db.host", "postgres.example.com",
		"--db.port", "5433",
	})

	fmt.Println("Server Host:", cfg.String("server.host"))
	fmt.Println("Server Port:", cfg.Int("server.port"))
	fmt.Println("DB Host:", cfg.String("db.host"))
	fmt.Println("DB Port:", cfg.Int("db.port"))
	// Output:
	// Server Host: api.example.com
	// Server Port: 9000
	// DB Host: postgres.example.com
	// DB Port: 5433
}

// ExampleStructUnmarshal demonstrates unmarshaling into structs with tags.
func ExampleStructUnmarshal() {
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

	cfg.Unmarshal(&config)

	fmt.Printf("Server: %s:%d\n", config.Server.Host, config.Server.Port)
	fmt.Printf("DB: %s:%d\n", config.DB.Host, config.DB.Port)
	fmt.Printf("Debug: %v\n", config.Debug)
	// Output:
	// Server: api.dev.local:3000
	// DB: db.dev.local:5432
	// Debug: true
}

// ExampleStringSlice demonstrates repeatable flags.
func ExampleStringSlice() {
	fs := New("myapp")
	fs.StringSlice("tag", "t", "tags to apply")

	cfg, _ := fs.Parse([]string{
		"-t", "important",
		"-t", "urgent",
		"-t", "review",
	})

	tags := cfg.StringSlice("tag")
	fmt.Println("Tags:", tags)
	// Output:
	// Tags: [important urgent review]
}

// ExamplePrecedence demonstrates CLI > env > default precedence.
func ExamplePrecedence() {
	fs := New("myapp")
	fs.String("config", "c", "default.conf", "config file").Env("CONFIG_FILE")

	os.Setenv("CONFIG_FILE", "env.conf")

	// CLI argument takes precedence
	cfg, _ := fs.Parse([]string{"--config", "cli.conf"})
	fmt.Println("With CLI arg:", cfg.String("config"))

	// Env var is used when no CLI arg
	cfg, _ = fs.Parse([]string{})
	fmt.Println("With env var:", cfg.String("config"))

	// Default is used when nothing else
	os.Unsetenv("CONFIG_FILE")
	cfg, _ = fs.Parse([]string{})
	fmt.Println("With default:", cfg.String("config"))
	// Output:
	// With CLI arg: cli.conf
	// With env var: env.conf
	// With default: default.conf
}
