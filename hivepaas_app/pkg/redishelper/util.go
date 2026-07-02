package redishelper

import (
	"encoding/json"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

func ParseString(val any) string {
	if val == nil {
		return ""
	}
	str, ok := val.(string)
	if ok {
		return str
	}
	bytes, ok := val.([]byte)
	if ok {
		return string(bytes)
	}
	return fmt.Sprintf("%v", val)
}

func ParseBytes(val any) []byte {
	if val == nil {
		return nil
	}
	bytes, ok := val.([]byte)
	if ok {
		return bytes
	}
	str, ok := val.(string)
	if ok {
		return []byte(str)
	}
	return fmt.Appendf(nil, "%v", val)
}

var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal

func marshalSlice[T any](values []T) ([]any, error) {
	result := make([]any, 0, len(values))
	for i := range values {
		value, err := jsonMarshal(values[i])
		if err != nil {
			return nil, apperrors.New(err)
		}
		result = append(result, reflectutil.UnsafeBytesToStr(value))
	}
	return result, nil
}

func marshalKVSlices[T any](keys []string, values []T) ([]any, error) {
	length := len(keys)
	if length != len(values) {
		return nil, apperrors.New(apperrors.ErrArgumentInvalid)
	}
	data := make([]any, 0, length*2) //nolint:mnd
	for i := range length {
		value, err := jsonMarshal(values[i])
		if err != nil {
			return nil, apperrors.New(err)
		}
		data = append(data, keys[i], reflectutil.UnsafeBytesToStr(value))
	}
	return data, nil
}

func unmarshalStr[T any](data string) (value T, err error) {
	if data == "" {
		return value, nil
	}
	err = jsonUnmarshal(reflectutil.UnsafeStrToBytes(data), &value)
	if err != nil {
		return value, apperrors.New(err)
	}
	return value, nil
}

func unmarshalSlice[T any](data ...any) ([]T, error) {
	result := make([]T, 0, len(data))
	for _, item := range data {
		var value T
		if item == nil {
			result = append(result, value)
			continue
		}
		err := jsonUnmarshal(ParseBytes(item), &value)
		if err != nil {
			return nil, apperrors.New(err)
		}
		result = append(result, value)
	}
	return result, nil
}

func unmarshalStrSlice[T any](data ...string) ([]T, error) {
	result := make([]T, 0, len(data))
	for _, item := range data {
		var value T
		if item == "" {
			result = append(result, value)
			continue
		}
		err := jsonUnmarshal(reflectutil.UnsafeStrToBytes(item), &value)
		if err != nil {
			return nil, apperrors.New(err)
		}
		result = append(result, value)
	}
	return result, nil
}

func unmarshalStrMap[T any](data map[string]string) (map[string]T, error) {
	result := make(map[string]T, len(data))
	for k, item := range data {
		var value T
		err := jsonUnmarshal(reflectutil.UnsafeStrToBytes(item), &value)
		if err != nil {
			return nil, apperrors.New(err)
		}
		result[k] = value
	}
	return result, nil
}
