package varg

import (
	"fmt"
	"reflect"
	"strings"
)

// Config holds parsed flag values.
type Config struct {
	values map[string]interface{}
	flags  map[string]*Flag
}

// Get returns the raw value for a key.
func (c *Config) Get(key string) (interface{}, bool) {
	val, ok := c.values[key]
	return val, ok
}

// String returns the string value for a key.
func (c *Config) String(key string) string {
	if val, ok := c.values[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// Int returns the int value for a key.
func (c *Config) Int(key string) int {
	if val, ok := c.values[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// Bool returns the bool value for a key.
func (c *Config) Bool(key string) bool {
	if val, ok := c.values[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Float64 returns the float64 value for a key.
func (c *Config) Float64(key string) float64 {
	if val, ok := c.values[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

// StringSlice returns the string slice value for a key.
func (c *Config) StringSlice(key string) []string {
	if val, ok := c.values[key]; ok {
		if ss, ok := val.([]string); ok {
			return ss
		}
	}
	return []string{}
}

// Unmarshal populates a struct from parsed flag values.
// Uses struct tags in the format: `flag:"key"`
// For nested keys (e.g., "server.host"), the struct should have nested fields
// or the tag should specify the full path.
//
// Example:
//
//	type Config struct {
//	    Server struct {
//	        Host string `flag:"server.host"`
//	        Port int    `flag:"server.port"`
//	    }
//	}
func (c *Config) Unmarshal(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("unmarshal target must be a pointer")
	}

	val = val.Elem()
	return c.unmarshalValue(val, "")
}

func (c *Config) unmarshalValue(val reflect.Value, pathPrefix string) error {
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("unmarshal target must be a struct, got %s", val.Kind())
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check for explicit flag tag
		flagTag := field.Tag.Get("flag")
		if flagTag != "" {
			// Direct mapping with tag
			if err := c.setFieldValue(fieldVal, field.Type, flagTag); err != nil {
				return fmt.Errorf("error setting field %s: %w", field.Name, err)
			}
			continue
		}

		// Try to build nested key from struct hierarchy
		nestedKey := field.Name
		if pathPrefix != "" {
			nestedKey = pathPrefix + "." + nestedKey
		}

		// Try to find matching flag
		// Convert field name to lowercase for matching
		lookupKey := strings.ToLower(nestedKey)
		if _, ok := c.values[lookupKey]; ok {
			if err := c.setFieldValue(fieldVal, field.Type, lookupKey); err != nil {
				return fmt.Errorf("error setting field %s: %w", field.Name, err)
			}
			continue
		}

		// If field is a struct, recurse
		if fieldVal.Kind() == reflect.Struct {
			if err := c.unmarshalValue(fieldVal, nestedKey); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Config) setFieldValue(fieldVal reflect.Value, fieldType reflect.Type, key string) error {
	if !fieldVal.CanSet() {
		return nil
	}

	val, ok := c.values[key]
	if !ok {
		return nil
	}

	switch fieldType.Kind() {
	case reflect.String:
		if s, ok := val.(string); ok {
			fieldVal.SetString(s)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := val.(int); ok {
			fieldVal.SetInt(int64(i))
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			fieldVal.SetBool(b)
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := val.(float64); ok {
			fieldVal.SetFloat(f)
		}
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.String {
			if ss, ok := val.([]string); ok {
				fieldVal.Set(reflect.ValueOf(ss))
			}
		}
	default:
		return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}

	return nil
}

// All returns all parsed values as a map.
func (c *Config) All() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.values {
		result[k] = v
	}
	return result
}
