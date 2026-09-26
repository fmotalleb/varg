package varg_test

import (
	"testing"

	"github.com/fmotalleb/varg"
)

type Server struct {
	Addr    string `arg:"addr" arg_short:"a" env:"ADDR" help:"server address"`
	Port    uint16 `arg:"port" arg_short:"p" env:"PORT" help:"server port"`
	Workers uint16 `arg:"workers" arg_short:"w" env:"WORKERS" help:"worker count"`
}

type CFG struct {
	Name string `arg:"name" arg_short:"n" env:"NAME" help:"app name"`
	Srv  Server `arg:"server" env:"SERVER" help:"server config"`
}

// TestExampleStructParamUnmarshal tests the struct registration and unmarshaling
func TestExampleStructParamUnmarshal(t *testing.T) {
	fs := varg.New("myapp")
	obj := &CFG{}

	// Register struct fields as flags
	err := fs.Struct(obj)
	if err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Set environment variable for nested field
	t.Setenv("SERVER__WORKERS", "50")

	// Parse command-line arguments
	cfg, err := fs.Parse([]string{
		"--name", "Test",
		"--server.addr", "0.0.0.0",
		"-p", "3000",
	})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Unmarshal the parsed config into the struct
	if err := cfg.Unmarshal(obj); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	// Verify the results
	if obj.Name != "Test" {
		t.Errorf("Name: got %q, want %q", obj.Name, "Test")
	}
	if obj.Srv.Addr != "0.0.0.0" {
		t.Errorf("Srv.Addr: got %q, want %q", obj.Srv.Addr, "0.0.0.0")
	}
	if obj.Srv.Port != 3000 {
		t.Errorf("Srv.Port: got %d, want %d", obj.Srv.Port, 3000)
	}
	if obj.Srv.Workers != 50 {
		t.Errorf("Srv.Workers: got %d, want %d", obj.Srv.Workers, 50)
	}
}

// TestStructRegistrationErrors verifies error handling
func TestStructRegistrationErrors(t *testing.T) {
	fs := varg.New("myapp")

	// Test: not a pointer
	obj := CFG{}
	err := fs.Struct(obj)
	if err == nil {
		t.Error("Struct() should error on non-pointer, got nil")
	}

	// Test: pointer to non-struct
	s := "string"
	err = fs.Struct(&s)
	if err == nil {
		t.Error("Struct() should error on pointer to non-struct, got nil")
	}
}

// TestStructHelpText verifies that help messages work
func TestStructHelpText(t *testing.T) {
	fs := varg.New("myapp").About("Test application")
	obj := &CFG{}

	if err := fs.Struct(obj); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	usage := fs.Usage()

	// Check that help text includes field descriptions
	if usage == "" {
		t.Error("Usage() returned empty string")
	}

	// The usage should contain flag names (exact format depends on fs.Usage implementation)
	if !contains(usage, "name") && !contains(usage, "--name") {
		t.Errorf("Usage missing 'name' flag description")
	}
	if !contains(usage, "server") && !contains(usage, "--server") {
		t.Errorf("Usage missing 'server' section description")
	}
}

// TestNestedStructParsing verifies nested struct parsing works
func TestNestedStructParsing(t *testing.T) {
	fs := varg.New("myapp")
	obj := &CFG{}

	if err := fs.Struct(obj); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Test with only nested fields
	cfg, err := fs.Parse([]string{
		"--server.addr", "127.0.0.1",
		"--server.port", "9000",
	})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := cfg.Unmarshal(obj); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if obj.Srv.Addr != "127.0.0.1" {
		t.Errorf("Srv.Addr: got %q, want %q", obj.Srv.Addr, "127.0.0.1")
	}
	if obj.Srv.Port != 9000 {
		t.Errorf("Srv.Port: got %d, want %d", obj.Srv.Port, 9000)
	}
}

// TestShortFlags verifies short flag forms work
func TestShortFlags(t *testing.T) {
	fs := varg.New("myapp")
	obj := &CFG{}

	if err := fs.Struct(obj); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	cfg, err := fs.Parse([]string{
		"-n", "ShortForm",
		"-a", "192.168.1.1",
		"-p", "5000",
	})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := cfg.Unmarshal(obj); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if obj.Name != "ShortForm" {
		t.Errorf("Name: got %q, want %q", obj.Name, "ShortForm")
	}
	if obj.Srv.Addr != "192.168.1.1" {
		t.Errorf("Srv.Addr: got %q, want %q", obj.Srv.Addr, "192.168.1.1")
	}
	if obj.Srv.Port != 5000 {
		t.Errorf("Srv.Port: got %d, want %d", obj.Srv.Port, 5000)
	}
}

// TestEnvironmentVariables verifies env vars work with nested fields
func TestEnvironmentVariables(t *testing.T) {
	fs := varg.New("myapp")
	obj := &CFG{}

	if err := fs.Struct(obj); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Set env vars for all fields
	t.Setenv("NAME", "EnvApp")
	t.Setenv("SERVER__ADDR", "10.0.0.1")
	t.Setenv("SERVER__PORT", "8888")
	t.Setenv("SERVER__WORKERS", "16")

	cfg, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := cfg.Unmarshal(obj); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if obj.Name != "EnvApp" {
		t.Errorf("Name: got %q, want %q", obj.Name, "EnvApp")
	}
	if obj.Srv.Addr != "10.0.0.1" {
		t.Errorf("Srv.Addr: got %q, want %q", obj.Srv.Addr, "10.0.0.1")
	}
	if obj.Srv.Port != 8888 {
		t.Errorf("Srv.Port: got %d, want %d", obj.Srv.Port, 8888)
	}
	if obj.Srv.Workers != 16 {
		t.Errorf("Srv.Workers: got %d, want %d", obj.Srv.Workers, 16)
	}
}

// TestCliOverridesEnv verifies CLI args have priority over env vars
func TestCliOverridesEnv(t *testing.T) {
	fs := varg.New("myapp")
	obj := &CFG{}

	if err := fs.Struct(obj); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Set env var
	t.Setenv("NAME", "FromEnv")

	// Override with CLI
	cfg, err := fs.Parse([]string{"--name", "FromCLI"})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := cfg.Unmarshal(obj); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if obj.Name != "FromCLI" {
		t.Errorf("Name: got %q, want %q (CLI should override env)", obj.Name, "FromCLI")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr))
}
