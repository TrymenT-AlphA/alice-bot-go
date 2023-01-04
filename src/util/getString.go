package util

import (
	"encoding/json"
	"strconv"
)

func GetString(i interface{}) (string, error) {
	var key string

	if i == nil {
		return key, nil
	}

	switch i.(type) {
	case float64:
		ft := i.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := i.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := i.(int)
		key = strconv.Itoa(it)
	case uint:
		it := i.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := i.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := i.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := i.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := i.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := i.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := i.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := i.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := i.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = i.(string)
	case []byte:
		key = string(i.([]byte))
	default:
		bytes, err := json.Marshal(i)
		if err != nil {
			return "", err
		}
		key = string(bytes)
	}

	return key, nil
}
