package e2time

import "time"

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

func TimePointer(t time.Time) *time.Time {
	return &t
}
