package caelum

import "time"

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int8(key string, value int8) Field {
	return Field{Key: key, Value: value}
}

func Int16(key string, value int16) Field {
	return Field{Key: key, Value: value}
}

func Int32(key string, value int32) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Uint(key string, value uint) Field {
	return Field{Key: key, Value: value}
}

func Uint8(key string, value uint8) Field {
	return Field{Key: key, Value: value}
}

func Uint16(key string, value uint16) Field {
	return Field{Key: key, Value: value}
}

func Uint32(key string, value uint32) Field {
	return Field{Key: key, Value: value}
}

func Uint64(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

func Float32(key string, value float32) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

func NamedError(key string, err error) Field {
	return Field{Key: key, Value: err}
}
