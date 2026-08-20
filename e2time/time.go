package e2time

import (
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2exec"
)

func MustParse(format, value string) time.Time {
	t, err := time.Parse(format, value)
	if err != nil {
		t = time.Time{}
	}
	return t
}

func ToDay() time.Time {
	return MustParse("2006-01-02", time.Now().Format("2006-01-02"))
}

func AddDay(t time.Time, days int) time.Time {
	return t.Add(time.Hour * 24 * time.Duration(days))
}

//go:fix inline
func TimePointer(t time.Time) *time.Time {
	return new(t)
}

func SleepRandom(min time.Duration, max time.Duration) time.Duration {
	rn := time.Duration(e2exec.Must(e2crypto.RandomNumber(int64(min), int64(max))))
	time.Sleep(rn)
	return rn
}
