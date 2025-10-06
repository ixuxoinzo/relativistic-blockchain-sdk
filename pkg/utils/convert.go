package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type ConvertUtils struct{}

func NewConvertUtils() *ConvertUtils {
	return &ConvertUtils{}
}

func (cu *ConvertUtils) StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func (cu *ConvertUtils) StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (cu *ConvertUtils) StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func (cu *ConvertUtils) StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func (cu *ConvertUtils) IntToString(i int) string {
	return strconv.Itoa(i)
}

func (cu *ConvertUtils) Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func (cu *ConvertUtils) Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (cu *ConvertUtils) BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

func (cu *ConvertUtils) DurationToString(d time.Duration) string {
	return d.String()
}

func (cu *ConvertUtils) StringToDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func (cu *ConvertUtils) TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (cu *ConvertUtils) StringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func (cu *ConvertUtils) BytesToString(b []byte) string {
	return string(b)
}

func (cu *ConvertUtils) StringToBytes(s string) []byte {
	return []byte(s)
}

func (cu *ConvertUtils) InterfaceToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", val), nil
	case []byte:
		return string(val), nil
	case time.Time:
		return val.Format(time.RFC3339), nil
	case time.Duration:
		return val.String(), nil
	default:
		bytes, err := json.Marshal(val)
		if err != nil {
			return "", fmt.Errorf("failed to convert interface to string: %w", err)
		}
		return string(bytes), nil
	}
}

func (cu *ConvertUtils) StringToInterface(s string, target interface{}) error {
	return json.Unmarshal([]byte(s), target)
}

func (cu *ConvertUtils) MapToJSON(m map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return string(bytes), nil
}

func (cu *ConvertUtils) JSONToMap(s string) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}
	return m, nil
}

func (cu *ConvertUtils) StructToMap(v interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}
	
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}
	return m, nil
}

func (cu *ConvertUtils) MapToStruct(m map[string]interface{}, v interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	
	err = json.Unmarshal(bytes, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}
	return nil
}

func (cu *ConvertUtils) ConvertSlice(slice interface{}, convertFunc func(interface{}) (interface{}, error)) ([]interface{}, error) {
	var result []interface{}
	
	switch s := slice.(type) {
	case []string:
		for _, item := range s {
			converted, err := convertFunc(item)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
	case []int:
		for _, item := range s {
			converted, err := convertFunc(item)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
	case []float64:
		for _, item := range s {
			converted, err := convertFunc(item)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
	case []interface{}:
		for _, item := range s {
			converted, err := convertFunc(item)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
	default:
		return nil, fmt.Errorf("unsupported slice type: %T", slice)
	}
	
	return result, nil
}

func (cu *ConvertUtils) ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func (cu *ConvertUtils) ParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (cu *ConvertUtils) ParseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}