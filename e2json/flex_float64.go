package e2json

import (
	"encoding/json"
	"strconv"
)

type FlexFloat64 struct {
	Float64 float64
	Valid   bool
}

func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	n, v, err := parse[float64](data)
	f.Float64 = n
	f.Valid = v
	return err
}

func (f FlexFloat64) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.Float64)
}

func (f FlexFloat64) Value() (float64, bool) {
	return f.Float64, f.Valid
}

func (f FlexFloat64) String() string {
	return strconv.FormatFloat(f.Float64, 'g', -1, 64)
}
