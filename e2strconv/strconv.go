package e2strconv

import (
	"strconv"
	"time"
)

func MustParseInt64(s string, base, bitSize int) int64 {
	i, _ := strconv.ParseInt(s, base, bitSize)
	return i
}

func MustParseInt16(s string) int16 {
	i, _ := strconv.ParseInt(s, 10, 16)
	return int16(i)
}

func MustParseUint(s string, base, bitSize int) uint64 {
	i, _ := strconv.ParseUint(s, base, bitSize)
	return i
}

func MustParseFloat(s string, bitSize int) float64 {
	f, _ := strconv.ParseFloat(s, bitSize)
	return f
}

func MustParseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func MustParseStringUnixTime(s string) time.Time {
	sec := s
	nsec := "0"
	if len(s) > 10 {
		sec = s[0:10]
		nsec = s[10:]
	}
	return time.Unix(MustParseInt64(sec, 10, 64), MustParseInt64(nsec, 10, 64))
}
