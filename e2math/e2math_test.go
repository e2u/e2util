package e2math

import "testing"

func TestRoundToDecimal(t *testing.T) {
	cases := []struct {
		value  float64
		places int
		want   float64
	}{
		{1.23456, 2, 1.23},
		{1.235, 2, 1.24},
		{1.5, 0, 2},
		{-1.26, 1, -1.3},
	}
	for _, tc := range cases {
		if got := RoundToDecimal(tc.value, tc.places); got != tc.want {
			t.Errorf("RoundToDecimal(%v, %d) = %v, want %v", tc.value, tc.places, got, tc.want)
		}
	}
}
