package client

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"
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

	separatorRegex := regexp.MustCompile(sep)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := separatorRegex.Split(line, 2)
		if len(parts) < 2 {
			return &out, fmt.Errorf("kv: malformed line %q", line)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		fieldSettings, ok := fieldByTag[key]
		if !ok {
			// If direct indexing fails, try regexes
			regexFound := false
			for k, v := range fieldByTag {
				if !v.regex {
					continue
				}

				r, _ := regexp.Compile(k)
				if r.MatchString(key) {
					fieldSettings = v
					regexFound = true
					break
				}
			}

			if !regexFound{
				continue // unknown key: ignore
			}
		}

		if err := setField(v.Field(fieldSettings.index), val); err != nil {
			return &out, fmt.Errorf("kv: field %q: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return &out, fmt.Errorf("kv: reading body: %w", err)
	}

	return &out, nil
}

type fieldIndexSettings struct {
	index int
	csv bool
	regex bool
}

// buildFieldIndex maps tag name -> field index for direct field lookup.
func buildFieldIndex(t reflect.Type) map[string]fieldIndexSettings {
	m := make(map[string]fieldIndexSettings, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}

		components := strings.Split(tag, ",")
		name := components[0]
		rest := components[1:]

		settings := fieldIndexSettings {
			index: i,
			csv: false,
			regex: false,
		}
		if slices.Contains(rest, "csv") {
			settings.csv = true
		} else if slices.Contains(rest, "regex") {
			settings.regex = true
		}

		m[name] = settings
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

	case reflect.Slice:
		elemType := field.Type().Elem()
		if elemType.Kind() != reflect.String {
			return fmt.Errorf("unsupported slice element kind: %s", elemType.Kind())
		}

		items := strings.Split(val, ",")
		sliceVal := reflect.ValueOf(items).Convert(field.Type())
		field.Set(sliceVal)

	default:
		return fmt.Errorf("unsupported field kind %s", field.Kind())
	}

	return nil
}
