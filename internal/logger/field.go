package logger

import (
	"encoding"
	"encoding/json"
	"maps"
	"reflect"
	"time"
)

type Field struct {
	Key   string
	Value any
}

func normalizeFields(fields map[string]any) map[string]any {
	var normalized map[string]any

	for key, value := range fields {
		var replacement any

		switch value := value.(type) {
		case json.Marshaler, encoding.TextMarshaler:
			continue
		case time.Duration:
			replacement = value.String()
		case error:
			reflected := reflect.ValueOf(value)

			kind := reflected.Kind()
			if (kind == reflect.Chan || kind == reflect.Func || kind == reflect.Interface ||
				kind == reflect.Map || kind == reflect.Pointer || kind == reflect.Slice) && reflected.IsNil() {
				replacement = nil
			} else {
				replacement = value.Error()
			}
		default:
			continue
		}

		if normalized == nil {
			normalized = maps.Clone(fields)
		}

		normalized[key] = replacement
	}

	if normalized != nil {
		return normalized
	}

	return fields
}
