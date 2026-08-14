package e2json // 或者你自己的 package

import (
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"
)

// FlexDecimal 可以接受：
// - 正常數字字串（"13.5600"）
// - 空字串 ""
// - null
// - 真正的 JSON number
type FlexDecimal struct {
	Decimal decimal.Decimal
	Valid   bool
}

func (f *FlexDecimal) UnmarshalJSON(data []byte) error {
	// 處理 null
	if string(data) == "null" {
		f.Valid = false
		return nil
	}

	// 先試直接 Unmarshal 成 decimal
	var d decimal.Decimal
	if err := json.Unmarshal(data, &d); err == nil {
		f.Decimal = d
		f.Valid = true
		return nil
	}

	// 再試字串
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		f.Valid = false
		return nil
	}

	d, err := decimal.NewFromString(s)
	if err != nil {
		return err
	}

	f.Decimal = d
	f.Valid = true
	return nil
}

func (f FlexDecimal) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.Decimal.String())
}

// 方便取值
func (f FlexDecimal) Value() (decimal.Decimal, bool) {
	return f.Decimal, f.Valid
}
