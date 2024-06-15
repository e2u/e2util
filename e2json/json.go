package e2json

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultDateFormat = "2006-01-02"
)

var (
	// DateFormat 默认格式，如果要修改格式，则用 e2json.DateFormat 进行定义
	DateFormat string = defaultDateFormat
)

type Date struct {
	t time.Time
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

func MustToJSONByte(v interface{}, indent ...bool) []byte {
	if len(indent) > 0 && indent[0] {
		b, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			logrus.Errorf("marshal to json error=%v", err)
			return nil
		}
		return b
	}

	b, err := json.Marshal(v)
	if err != nil {
		logrus.Errorf("marshal to json error=%v", err)
		return nil
	}
	return b
}

func MustToJSONString(v interface{}, indent ...bool) string {
	return string(MustToJSONByte(v, indent...))
}

func MustFromJSONByte(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

func MustFromJSONString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

func MustIndentJSONByte(b []byte) []byte {
	var v interface{}
	_ = MustFromJSONByte(b, &v)
	if v != nil {
		return MustToJSONByte(v, true)
	}
	return nil
}

func MustIndentJSONString(s string) string {
	return string(MustIndentJSONByte([]byte(s)))
}

func MustFromReader(r io.Reader, v interface{}) error {
	b, err := io.ReadAll(r)
	if err != nil {
		logrus.Errorf("read error=%v", err)
		return err
	}
	return MustFromJSONByte(b, v)
}

func MustToJSONPString(v interface{}, callback ...string) string {
	j := MustToJSONString(v)

	if len(callback) == 0 {
		return fmt.Sprintf(`callback(%s)`, j)
	}
	return j
}

func MustToSecureJSONString(v interface{}, prefix ...string) string {
	j := MustToJSONString(v)
	if len(prefix) == 0 {
		return fmt.Sprintf(`while(1);%s`, j)
	}
	return j
}
