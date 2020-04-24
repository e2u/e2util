package e2json

import (
	"strings"
	"time"
)

const (
	defaultDateFormat = "2006-01-02"
)

var (
	// 默认格式，如果要修改格式，则用 e2json.DateFormat 进行定义
	DateFormat string
)

type Date struct {
	t time.Time
}

func init() {
	if len(DateFormat) == 0 {
		DateFormat = defaultDateFormat
	}
}

// NowDate json 序列化時間 time.Time 類型，格式化成日期格式,默認 2006-01-02
// *e2json.NowDate().SetTime(timeValue) 設置時間
// 如需要其他的時間格式,需要設置時間格式
// e2json.DateFormat = "2006-01-02"
func NowDate() *Date {
	return &Date{
		t: time.Now(),
	}
}

func (t *Date) UnmarshalJSON(buf []byte) error {
	tt, err := time.Parse(DateFormat, strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.t = tt
	return nil
}

func (t Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.t.Format(DateFormat) + `"`), nil
}

func (t Date) Time() time.Time {
	return t.t
}

func (t *Date) SetTime(tt time.Time) *Date {
	t.t = tt
	return t
}

func ParseDate(val string) Date {
	tt, err := time.Parse(DateFormat, val)
	if err != nil {
		return Date{
			time.Time{},
		}
	}
	return Date{
		tt,
	}
}

func (t Date) String() string {
	return t.t.Format(DateFormat)
}
