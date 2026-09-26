package varg

import (
	"fmt"
	"reflect"
	"strings"
)

// Struct registers flags for all exported fields in a struct.
// Fields must have `arg` tags to be registered as flags.
//
// Tag format:
//
//	`arg:"name" arg_short:"n" env:"ENV_VAR" help:"description"`
//
// For nested structs, the field names are concatenated with dots (e.g., "server.addr").
// Environment variable names for nested fields use double underscores (e.g., "SERVER__ADDR").
//
// Example:
//
//	type Config struct {
//	    Name string `arg:"name" arg_short:"n" env:"NAME" help:"application name"`
//	    Server struct {
//	        Addr string `arg:"addr" arg_short:"a" env:"ADDR" help:"server address"`
//	        Port uint16 `arg:"port" arg_short:"p" env:"PORT" help:"server port"`
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
		flag, err := fs.registerStructField(fullKey, shortTag, fieldVal, helpTag)
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
func (fs *FlagSet) registerStructField(key, short string, fieldVal reflect.Value, help string) (*Flag, error) {
	switch fieldVal.Kind() {
	case reflect.String:
		return fs.String(key, short, "", help), nil

	case reflect.Bool:
		return fs.Bool(key, short, false, help), nil

	case reflect.Int:
		return fs.Int(key, short, 0, help), nil
	case reflect.Int8:
		return fs.Int8(key, short, int8(0), help), nil
	case reflect.Int16:
		return fs.Int16(key, short, int16(0), help), nil
	case reflect.Int32:
		return fs.Int32(key, short, int32(0), help), nil
	case reflect.Int64:
		return fs.Int64(key, short, int64(0), help), nil

	case reflect.Uint:
		return fs.Uint(key, short, uint(0), help), nil
	case reflect.Uint8:
		return fs.Uint8(key, short, uint8(0), help), nil
	case reflect.Uint16:
		return fs.Uint16(key, short, uint16(0), help), nil
	case reflect.Uint32:
		return fs.Uint32(key, short, uint32(0), help), nil
	case reflect.Uint64:
		return fs.Uint64(key, short, uint64(0), help), nil

	case reflect.Float32:
		return fs.Float32(key, short, float32(0), help), nil
	case reflect.Float64:
		return fs.Float64(key, short, float64(0), help), nil

	case reflect.Slice:
		if fieldVal.Type().Elem().Kind() == reflect.String {
			return fs.StringSlice(key, short, help), nil
		}
		return nil, fmt.Errorf("unsupported slice type: %s", fieldVal.Type().Elem().Kind())

	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
	}
}
