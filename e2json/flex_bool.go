package e2json

import (
	"encoding/json"
	"strconv"
)

type FlexBool struct {
	Bool  bool
	Valid bool
}

func (f *FlexBool) UnmarshalJSON(data []byte) error {
	n, v, err := parse[bool](data)
	f.Bool = n
	f.Valid = v
	return err
}

func (f FlexBool) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.Bool)
}

func (f FlexBool) Value() (bool, bool) {
	return f.Bool, f.Valid
}

func (f FlexBool) String() string {
	return strconv.FormatBool(f.Bool)
}
