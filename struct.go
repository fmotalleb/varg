package varg

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Struct registers flags for all exported fields in a struct.
// Fields must have `arg` tags to be registered as flags.
//
// Tag format:
//
//	`arg:"name" arg_short:"n" env:"ENV_VAR" default:"value" help:"description"`
//
// For nested structs, the field names are concatenated with dots (e.g., "server.addr").
// Environment variable names for nested fields use double underscores (e.g., "SERVER__ADDR").
// Default values are parsed according to the field type.
//
// Example:
//
//	type Config struct {
//	    Name string `arg:"name" arg_short:"n" env:"NAME" default:"myapp" help:"application name"`
//	    Server struct {
//	        Addr string `arg:"addr" arg_short:"a" env:"ADDR" default:"127.0.0.1" help:"server address"`
//	        Port uint16 `arg:"port" arg_short:"p" env:"PORT" default:"8080" help:"server port"`
//	    } `arg:"server"`
//	}
//
//	cfg := &Config{}
//	fs.Struct(cfg)
func (fs *FlagSet) Struct(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer {
		return fmt.Errorf("Struct requires a pointer to a struct")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("Struct requires a pointer to a struct, got %s", val.Kind())
	}

	return fs.walkStruct(val, "")
}

// walkStruct recursively processes a struct and its nested structs.
func (fs *FlagSet) walkStruct(val reflect.Value, keyPrefix string) error {
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the arg tag
		argTag := field.Tag.Get("arg")

		// If this field is a struct, handle it specially
		if fieldVal.Kind() == reflect.Struct {
			// Only process if this struct field has an arg tag
			if argTag == "" {
				continue
			}

			// Build the full key name for the nested struct
			fullKey := argTag
			if keyPrefix != "" {
				fullKey = keyPrefix + "." + argTag
			}

			// Recurse into the nested struct
			if err := fs.walkStruct(fieldVal, fullKey); err != nil {
				return err
			}
			continue
		}

		// Non-struct fields must have an arg tag to be registered
		if argTag == "" {
			continue
		}

		// Build the full key name
		fullKey := argTag
		if keyPrefix != "" {
			fullKey = keyPrefix + "." + argTag
		}

		// Get other tags
		shortTag := field.Tag.Get("arg_short")
		envTag := field.Tag.Get("env")
		defaultTag := field.Tag.Get("default")
		helpTag := field.Tag.Get("help")

		// Build the environment variable name
		var envVar string
		if envTag != "" {
			if keyPrefix != "" {
				// For nested fields: keyPrefix (with dots to __) + __ + envTag
				envVar = strings.ToUpper(strings.ReplaceAll(keyPrefix, ".", "__")) + "__" + envTag
			} else {
				envVar = envTag
			}
		}

		// Register the flag based on the field type
		flag, err := fs.registerStructField(fullKey, shortTag, fieldVal, defaultTag, helpTag)
		if err != nil {
			return fmt.Errorf("error registering field %s: %w", field.Name, err)
		}

		// Set the environment variable if specified
		if envVar != "" {
			flag.Env(envVar)
		}
	}

	return nil
}

// registerStructField registers a flag for a struct field based on its type.
func (fs *FlagSet) registerStructField(key, short string, fieldVal reflect.Value, defaultStr, help string) (*Flag, error) {
	switch fieldVal.Kind() {
	case reflect.String:
		return fs.String(key, short, defaultStr, help), nil

	case reflect.Bool:
		defaultVal := false
		if defaultStr != "" {
			parsed, err := strconv.ParseBool(defaultStr)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for bool: %q", defaultStr)
			}
			defaultVal = parsed
		}
		return fs.Bool(key, short, defaultVal, help), nil

	case reflect.Int:
		defaultVal := int(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseInt(defaultStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for int: %q", defaultStr)
			}
			defaultVal = int(parsed)
		}
		return fs.Int(key, short, defaultVal, help), nil
	case reflect.Int8:
		defaultVal := int8(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseInt(defaultStr, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for int8: %q", defaultStr)
			}
			defaultVal = int8(parsed)
		}
		return fs.Int8(key, short, defaultVal, help), nil
	case reflect.Int16:
		defaultVal := int16(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseInt(defaultStr, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for int16: %q", defaultStr)
			}
			defaultVal = int16(parsed)
		}
		return fs.Int16(key, short, defaultVal, help), nil
	case reflect.Int32:
		defaultVal := int32(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseInt(defaultStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for int32: %q", defaultStr)
			}
			defaultVal = int32(parsed)
		}
		return fs.Int32(key, short, defaultVal, help), nil
	case reflect.Int64:
		defaultVal := int64(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseInt(defaultStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for int64: %q", defaultStr)
			}
			defaultVal = parsed
		}
		return fs.Int64(key, short, defaultVal, help), nil

	case reflect.Uint:
		defaultVal := uint(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseUint(defaultStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for uint: %q", defaultStr)
			}
			defaultVal = uint(parsed)
		}
		return fs.Uint(key, short, defaultVal, help), nil
	case reflect.Uint8:
		defaultVal := uint8(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseUint(defaultStr, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for uint8: %q", defaultStr)
			}
			defaultVal = uint8(parsed)
		}
		return fs.Uint8(key, short, defaultVal, help), nil
	case reflect.Uint16:
		defaultVal := uint16(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseUint(defaultStr, 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for uint16: %q", defaultStr)
			}
			defaultVal = uint16(parsed)
		}
		return fs.Uint16(key, short, defaultVal, help), nil
	case reflect.Uint32:
		defaultVal := uint32(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseUint(defaultStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for uint32: %q", defaultStr)
			}
			defaultVal = uint32(parsed)
		}
		return fs.Uint32(key, short, defaultVal, help), nil
	case reflect.Uint64:
		defaultVal := uint64(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseUint(defaultStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for uint64: %q", defaultStr)
			}
			defaultVal = parsed
		}
		return fs.Uint64(key, short, defaultVal, help), nil

	case reflect.Float32:
		defaultVal := float32(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseFloat(defaultStr, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for float32: %q", defaultStr)
			}
			defaultVal = float32(parsed)
		}
		return fs.Float32(key, short, defaultVal, help), nil
	case reflect.Float64:
		defaultVal := float64(0)
		if defaultStr != "" {
			parsed, err := strconv.ParseFloat(defaultStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid default value for float64: %q", defaultStr)
			}
			defaultVal = parsed
		}
		return fs.Float64(key, short, defaultVal, help), nil

	case reflect.Slice:
		if fieldVal.Type().Elem().Kind() == reflect.String {
			return fs.StringSlice(key, short, help), nil
		}
		return nil, fmt.Errorf("unsupported slice type: %s", fieldVal.Type().Elem().Kind())

	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
	}
}
