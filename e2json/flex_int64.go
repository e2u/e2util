package e2json

import (
	"encoding/json"
	"strconv"
)

type FlexInt64 struct {
	Int64 int64
	Valid bool
}

func (f *FlexInt64) UnmarshalJSON(data []byte) error {
	n, v, err := parse[int64](data)
	f.Int64 = n
	f.Valid = v
	return err
}

func (f FlexInt64) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.Int64)
}

func (f FlexInt64) Value() (int64, bool) {
	return f.Int64, f.Valid
}

func (f FlexInt64) String() string {
	return strconv.FormatInt(f.Int64, 10)
}
