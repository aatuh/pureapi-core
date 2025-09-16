package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// AnyToJSONString converts data to a JSON string. It supports converting a
// large variety of data types. It will panic if the data cannot be converted.
//
// Parameters:
//   - data: The data to convert to JSON.
//
// Returns:
//   - string: A JSON string representation of the data.
func AnyToJSONString(data ...any) string {
	jsonBytes, err := AnyToJSONBytes(data[0])
	if err != nil {
		panic(fmt.Errorf("AnyToJSONString: error converting to JSON: %v", err))
	}
	intended, err := indentJSON(jsonBytes)
	if err != nil {
		panic(fmt.Errorf("AnyToJSONString: error indenting JSON: %w", err))
	}
	return intended
}

// AnyToJSONString converts data to a JSON byte array. It supports converting a
// large variety of data types.
//
// Parameters:
//   - data: The data to convert to JSON.
//
// Returns:
//   - string: A JSON string representation of the data.
//   - error: An error if the data cannot be converted.
func AnyToJSONBytes(data any) ([]byte, error) {
	convertedData, err := convertToSerializable(data)
	if err != nil {
		return nil, fmt.Errorf(
			"AnyToJSONBytes: error converting to serializable: %w", err,
		)
	}
	return json.Marshal(convertedData)
}

// convertToSerializable converts multiple data types to a serializable type.
func convertToSerializable(data any) (any, error) {
	val := reflect.ValueOf(data)
	kind := val.Kind()

	switch kind {
	case reflect.Ptr:
		if val.IsNil() {
			return nil, nil
		}
		return convertToSerializable(val.Elem().Interface())
	case reflect.Func:
		// Simplified function representation.
		// Real function names are not directly accessible via reflection.
		return "func()", nil
	case reflect.Slice, reflect.Array:
		a := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			elem, err := convertToSerializable(val.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			a[i] = elem
		}
		return a, nil
	case reflect.Map:
		// Ensure all map keys are strings.
		m := make(map[string]any)
		for _, key := range val.MapKeys() {
			// Convert key to a string representation.
			strKey := fmt.Sprintf("%v", key.Interface())
			convertedValue, err := convertToSerializable(
				val.MapIndex(key).Interface(),
			)
			if err != nil {
				return nil, err
			}
			m[strKey] = convertedValue
		}
		return m, nil
	case reflect.Struct:
		// Only convert exported fields of structs.
		s := make(map[string]any)
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			// Field is exported.
			if field.PkgPath == "" {
				convertedValue, err := convertToSerializable(
					val.Field(i).Interface(),
				)
				if err != nil {
					return nil, err
				}
				s[field.Name] = convertedValue
			}
		}
		return s, nil
	case reflect.Chan,
		reflect.Complex64,
		reflect.Complex128,
		reflect.UnsafePointer:
		// Return the kind instead of the value.
		return kind, nil
	default:
		// Basic types are directly serializable.
		return data, nil
	}
}

// indentJSON indents a JSON byte array.
func indentJSON(jsonBytes []byte) (string, error) {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, jsonBytes, "", "\t")
	if err != nil {
		return "", fmt.Errorf("indentJSON: error indenting JSON: %w", err)
	}
	return buffer.String(), nil
}
