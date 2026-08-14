package e2json

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

func parse[T int | int64 | float64 | bool](data []byte) (T, bool, error) {
	var n T
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		return n, false, nil
	}

	if err := json.Unmarshal(data, &n); err == nil {
		return n, true, nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return n, false, err
	}

	switch any(n).(type) {
	case float64:
		if pv, err := strconv.ParseFloat(s, 64); err == nil {
			return any(pv).(T), true, nil
		}
	case int:
		if pv, err := strconv.ParseInt(s, 10, 64); err == nil {
			return any(pv).(T), true, nil
		}
	case int64:
		if pv, err := strconv.ParseInt(s, 10, 64); err == nil {
			return any(pv).(T), true, nil
		}
	case bool:
		if pv, err := strconv.ParseBool(s); err == nil {
			return any(pv).(T), true, nil
		}

		switch strings.ToUpper(s) {
		case "1", "T", "TRUE", "YES", "Y":
			return any(true).(T), true, nil
		case "0", "F", "FALSE", "NO", "N":
			return any(false).(T), true, nil
		}
	}
	return n, false, nil
}
