package client

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const tagName = "kv"

// KVDeserializer returns a Deserializer[T] that parses "key=value" lines
// from the body and populates a struct of type T based on `kv` tags.
func KVDeserializer[T any](r io.Reader, sep string) (*T, error) {
	var out T
	v := reflect.ValueOf(&out).Elem()
	if v.Kind() != reflect.Struct {
		return &out, fmt.Errorf("kv: type %T is not a struct", out)
	}

	fieldByTag := buildFieldIndex(v.Type())

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, sep)
		if !ok {
			return &out, fmt.Errorf("kv: malformed line %q", line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		idx, ok := fieldByTag[key]
		if !ok {
			continue // unknown key: ignore
		}

		if err := setField(v.Field(idx), val); err != nil {
			return &out, fmt.Errorf("kv: field %q: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return &out, fmt.Errorf("kv: reading body: %w", err)
	}

	return &out, nil
}

// buildFieldIndex maps tag name -> field index for direct field lookup.
func buildFieldIndex(t reflect.Type) map[string]int {
	m := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}
		// TODO: support `kv:"name,omitempty"`-style tags if needed later;
		// for now just take the first comma-separated part
		name, _, _ := strings.Cut(tag, ",")
		m[name] = i
	}
	return m
}

// setField converts the string value into the field's type and sets it.
func setField(field reflect.Value, val string) error {
	if !field.CanSet() {
		return fmt.Errorf("cannot set field")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	default:
		return fmt.Errorf("unsupported field kind %s", field.Kind())
	}

	return nil
}
