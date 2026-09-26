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

// asInt64 converts val to an int64 when it holds an integer value.
// Unsigned values that do not fit in an int64 report false.
func asInt64(val interface{}) (int64, bool) {
	switch rv := reflect.ValueOf(val); rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u := rv.Uint(); u <= 1<<63-1 {
			return int64(u), true
		}
	}
	return 0, false
}

// asUint64 converts val to a uint64 when it holds a non-negative integer.
func asUint64(val interface{}) (uint64, bool) {
	switch rv := reflect.ValueOf(val); rv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v := rv.Int(); v >= 0 {
			return uint64(v), true
		}
	}
	return 0, false
}

// asFloat64 converts val to a float64 when it holds a number.
func asFloat64(val interface{}) (float64, bool) {
	switch rv := reflect.ValueOf(val); rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	}
	return 0, false
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
	v, _ := asInt64(c.values[key])
	return int(v)
}

// Int8 returns the int8 value for a key, or 0 when it does not fit.
func (c *Config) Int8(key string) int8 {
	v, ok := asInt64(c.values[key])
	if !ok || int64(int8(v)) != v {
		return 0
	}
	return int8(v)
}

// Int16 returns the int16 value for a key, or 0 when it does not fit.
func (c *Config) Int16(key string) int16 {
	v, ok := asInt64(c.values[key])
	if !ok || int64(int16(v)) != v {
		return 0
	}
	return int16(v)
}

// Int32 returns the int32 value for a key, or 0 when it does not fit.
func (c *Config) Int32(key string) int32 {
	v, ok := asInt64(c.values[key])
	if !ok || int64(int32(v)) != v {
		return 0
	}
	return int32(v)
}

// Int64 returns the int64 value for a key.
func (c *Config) Int64(key string) int64 {
	v, _ := asInt64(c.values[key])
	return v
}

// Uint returns the uint value for a key.
func (c *Config) Uint(key string) uint {
	v, _ := asUint64(c.values[key])
	return uint(v)
}

// Uint8 returns the uint8 value for a key, or 0 when it does not fit.
func (c *Config) Uint8(key string) uint8 {
	v, ok := asUint64(c.values[key])
	if !ok || uint64(uint8(v)) != v {
		return 0
	}
	return uint8(v)
}

// Uint16 returns the uint16 value for a key, or 0 when it does not fit.
func (c *Config) Uint16(key string) uint16 {
	v, ok := asUint64(c.values[key])
	if !ok || uint64(uint16(v)) != v {
		return 0
	}
	return uint16(v)
}

// Uint32 returns the uint32 value for a key, or 0 when it does not fit.
func (c *Config) Uint32(key string) uint32 {
	v, ok := asUint64(c.values[key])
	if !ok || uint64(uint32(v)) != v {
		return 0
	}
	return uint32(v)
}

// Uint64 returns the uint64 value for a key.
func (c *Config) Uint64(key string) uint64 {
	v, _ := asUint64(c.values[key])
	return v
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

// Float32 returns the float32 value for a key.
func (c *Config) Float32(key string) float32 {
	v, ok := asFloat64(c.values[key])
	if !ok {
		return 0
	}
	return float32(v)
}

// Float64 returns the float64 value for a key.
func (c *Config) Float64(key string) float64 {
	v, ok := asFloat64(c.values[key])
	if !ok {
		return 0
	}
	return v
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
	if val.Kind() != reflect.Pointer {
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
		if v, ok := asInt64(val); ok {
			if fieldVal.OverflowInt(v) {
				return fmt.Errorf("value %v is out of range for %s", val, fieldType)
			}
			fieldVal.SetInt(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, ok := asUint64(val); ok {
			if fieldVal.OverflowUint(v) {
				return fmt.Errorf("value %v is out of range for %s", val, fieldType)
			}
			fieldVal.SetUint(v)
		} else if _, isInt := asInt64(val); isInt {
			return fmt.Errorf("value %v is out of range for %s", val, fieldType)
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			fieldVal.SetBool(b)
		}
	case reflect.Float32, reflect.Float64:
		if v, ok := asFloat64(val); ok {
			if fieldVal.OverflowFloat(v) {
				return fmt.Errorf("value %v is out of range for %s", val, fieldType)
			}
			fieldVal.SetFloat(v)
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
