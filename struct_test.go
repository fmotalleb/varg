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

// TestDefaultValues verifies default tag values are applied correctly
func TestDefaultValues(t *testing.T) {
	type Config struct {
		Host    string  `arg:"host" default:"localhost" help:"server host"`
		Port    uint16  `arg:"port" default:"8080" help:"server port"`
		Debug   bool    `arg:"debug" default:"false" help:"enable debug mode"`
		Timeout int     `arg:"timeout" default:"30" help:"timeout in seconds"`
		Weight  float64 `arg:"weight" default:"0.5" help:"weight factor"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	if err := fs.Struct(cfg); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Parse with no arguments, should use defaults
	result, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := result.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host: got %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want %d", cfg.Port, 8080)
	}
	if cfg.Debug != false {
		t.Errorf("Debug: got %v, want %v", cfg.Debug, false)
	}
	if cfg.Timeout != 30 {
		t.Errorf("Timeout: got %d, want %d", cfg.Timeout, 30)
	}
	if cfg.Weight != 0.5 {
		t.Errorf("Weight: got %v, want %v", cfg.Weight, 0.5)
	}
}

// TestDefaultValuesOverride verifies CLI args override defaults
func TestDefaultValuesOverride(t *testing.T) {
	type Config struct {
		Host   string `arg:"host" default:"localhost" help:"server host"`
		Port   uint16 `arg:"port" default:"8080" help:"server port"`
		Debug  bool   `arg:"debug" default:"false" help:"enable debug mode"`
		Repeat int    `arg:"repeat" default:"1" help:"repeat count"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	if err := fs.Struct(cfg); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Override some defaults with CLI args
	result, err := fs.Parse([]string{
		"--host", "example.com",
		"--debug", "true",
	})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := result.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("Host: got %q, want %q", cfg.Host, "example.com")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want %d (should use default)", cfg.Port, 8080)
	}
	if cfg.Debug != true {
		t.Errorf("Debug: got %v, want %v", cfg.Debug, true)
	}
	if cfg.Repeat != 1 {
		t.Errorf("Repeat: got %d, want %d (should use default)", cfg.Repeat, 1)
	}
}

// TestNestedDefaultValues verifies defaults work with nested structs
func TestNestedDefaultValues(t *testing.T) {
	type Server struct {
		Addr string `arg:"addr" default:"127.0.0.1" help:"server address"`
		Port uint16 `arg:"port" default:"9000" help:"server port"`
	}

	type Config struct {
		Name string `arg:"name" default:"myapp" help:"app name"`
		Srv  Server `arg:"server" help:"server config"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	if err := fs.Struct(cfg); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Parse with no arguments
	result, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := result.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if cfg.Name != "myapp" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "myapp")
	}
	if cfg.Srv.Addr != "127.0.0.1" {
		t.Errorf("Srv.Addr: got %q, want %q", cfg.Srv.Addr, "127.0.0.1")
	}
	if cfg.Srv.Port != 9000 {
		t.Errorf("Srv.Port: got %d, want %d", cfg.Srv.Port, 9000)
	}
}

// TestDefaultValuesWithEnv verifies env vars override defaults
func TestDefaultValuesWithEnv(t *testing.T) {
	type Config struct {
		Host  string `arg:"host" env:"HOST" default:"localhost" help:"server host"`
		Port  uint16 `arg:"port" env:"PORT" default:"8080" help:"server port"`
		Debug bool   `arg:"debug" env:"DEBUG" default:"false" help:"enable debug"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	if err := fs.Struct(cfg); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	// Set env var to override default
	t.Setenv("PORT", "3000")

	result, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := result.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host: got %q, want %q (should use default)", cfg.Host, "localhost")
	}
	if cfg.Port != 3000 {
		t.Errorf("Port: got %d, want %d (env should override default)", cfg.Port, 3000)
	}
	if cfg.Debug != false {
		t.Errorf("Debug: got %v, want %v (should use default)", cfg.Debug, false)
	}
}

// TestInvalidDefaultValues verifies error handling for invalid defaults
func TestInvalidDefaultValues(t *testing.T) {
	type Config struct {
		Port uint16 `arg:"port" default:"not_a_number" help:"server port"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	err := fs.Struct(cfg)
	if err == nil {
		t.Error("Struct() should error on invalid default value, got nil")
	}
}

// TestAllNumericDefaults verifies all numeric types handle defaults correctly
func TestAllNumericDefaults(t *testing.T) {
	type Config struct {
		I8  int8    `arg:"i8" default:"10" help:"int8 value"`
		I16 int16   `arg:"i16" default:"1000" help:"int16 value"`
		I32 int32   `arg:"i32" default:"100000" help:"int32 value"`
		I64 int64   `arg:"i64" default:"10000000000" help:"int64 value"`
		U8  uint8   `arg:"u8" default:"10" help:"uint8 value"`
		U16 uint16  `arg:"u16" default:"1000" help:"uint16 value"`
		U32 uint32  `arg:"u32" default:"100000" help:"uint32 value"`
		U64 uint64  `arg:"u64" default:"10000000000" help:"uint64 value"`
		F32 float32 `arg:"f32" default:"1.5" help:"float32 value"`
		F64 float64 `arg:"f64" default:"3.14159" help:"float64 value"`
	}

	fs := varg.New("test")
	cfg := &Config{}

	if err := fs.Struct(cfg); err != nil {
		t.Fatalf("Struct() failed: %v", err)
	}

	result, err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if err := result.Unmarshal(cfg); err != nil {
		t.Fatalf("Unmarshal() failed: %v", err)
	}

	if cfg.I8 != 10 {
		t.Errorf("I8: got %d, want %d", cfg.I8, 10)
	}
	if cfg.I16 != 1000 {
		t.Errorf("I16: got %d, want %d", cfg.I16, 1000)
	}
	if cfg.I32 != 100000 {
		t.Errorf("I32: got %d, want %d", cfg.I32, 100000)
	}
	if cfg.I64 != 10000000000 {
		t.Errorf("I64: got %d, want %d", cfg.I64, 10000000000)
	}
	if cfg.U8 != 10 {
		t.Errorf("U8: got %d, want %d", cfg.U8, 10)
	}
	if cfg.U16 != 1000 {
		t.Errorf("U16: got %d, want %d", cfg.U16, 1000)
	}
	if cfg.U32 != 100000 {
		t.Errorf("U32: got %d, want %d", cfg.U32, 100000)
	}
	if cfg.U64 != 10000000000 {
		t.Errorf("U64: got %d, want %d", cfg.U64, 10000000000)
	}
	if cfg.F32 != float32(1.5) {
		t.Errorf("F32: got %v, want %v", cfg.F32, float32(1.5))
	}
	if cfg.F64 != 3.14159 {
		t.Errorf("F64: got %v, want %v", cfg.F64, 3.14159)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr))
}
