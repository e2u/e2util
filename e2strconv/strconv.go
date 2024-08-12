package e2strconv

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/e2u/e2util/e2exec"
)

func MustParseInt(s string) int {
	s = strings.TrimSpace(s)
	return int(MustParseInt64(s, 10, 64))
}

func MustParseInt64(s string, base, bitSize int) int64 {
	s = strings.TrimSpace(s)
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}

func MustParseInt16(s string) int16 {
	s = strings.TrimSpace(s)
	i, _ := strconv.ParseInt(s, 10, 16)
	return int16(i)
}

func MustParseUint(s string, base, bitSize int) uint64 {
	s = strings.TrimSpace(s)
	i, _ := strconv.ParseUint(s, base, bitSize)
	return i
}

func MustParseFloat(s any) float64 {
	switch v := s.(type) {
	case string:
		return e2exec.Must(strconv.ParseFloat(v, 64))
	case []byte:
		return e2exec.Must(strconv.ParseFloat(string(v), 64))
	default:
		return e2exec.Must(strconv.ParseFloat(fmt.Sprintf("%v", v), 64))
	}
}

func MustParseBool(s string) bool {
	s = strings.TrimSpace(s)
	b, _ := strconv.ParseBool(s)
	return b
}

func MustParseStringUnixTime(s string) time.Time {
	s = strings.TrimSpace(s)
	sec := s
	nsec := "0"
	if len(s) > 10 {
		sec = s[0:10]
		nsec = s[10:]
	}
	return time.Unix(MustParseInt64(sec, 10, 64), MustParseInt64(nsec, 10, 64))
}
