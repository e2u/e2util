package e2crypto

import (
	"math"
	"slices"
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"normal case", 10, false},
		{"zero length", 0, true},
		{"negative length", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomString(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.n {
					t.Errorf("RandomString() length = %v, want %v", len(got), tt.n)
				}
				// 檢查是否只包含 encoder 中的字符
				for _, c := range got {
					if !strings.ContainsRune(string(encoder), c) {
						t.Errorf("RandomString() contains invalid character: %v", c)
					}
				}
			}
		})
	}
}

func TestRandomBytes(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"normal case", 16, false},
		{"zero length", 0, true},
		{"negative length", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomBytes(tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.n {
				t.Errorf("RandomBytes() length = %v, want %v", len(got), tt.n)
			}
		})
	}
}

func TestRandomNumber(t *testing.T) {
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
	}{
		{"normal case", 1, 10, false},
		{"same min max", 5, 5, false},
		{"min > max", 10, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomNumber(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got < tt.min || got > tt.max {
					t.Errorf("RandomNumber() = %v, out of range [%v, %v]", got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestRandomFloat(t *testing.T) {
	tests := []struct {
		name    string
		min     float64
		max     float64
		wantErr bool
	}{
		{"normal case", 0.0, 1.0, false},
		{"negative range", -1.0, 1.0, false},
		{"min > max", 1.0, 0.0, false}, // 函數會自動交換
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomFloat(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				actualMin := math.Min(tt.min, tt.max)
				actualMax := math.Max(tt.min, tt.max)
				if got < actualMin || got > actualMax {
					t.Errorf("RandomFloat() = %v, out of range [%v, %v]", got, actualMin, actualMax)
				}
			}
		})
	}
}

func TestRandomElement(t *testing.T) {
	intSlice := []int{1, 2, 3, 4, 5}
	stringSlice := []string{"a", "b", "c"}
	emptySlice := []int{}

	tests := []struct {
		name    string
		slice   any
		wantErr bool
	}{
		{"int slice", intSlice, false},
		{"string slice", stringSlice, false},
		{"empty slice", emptySlice, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch s := tt.slice.(type) {
			case []int:
				got, err := RandomElement(s)
				if (err != nil) != tt.wantErr {
					t.Errorf("RandomElement() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					found := slices.Contains(s, got)
					if !found {
						t.Errorf("RandomElement() = %v, not in slice %v", got, s)
					}
				}
			case []string:
				got, err := RandomElement(s)
				if (err != nil) != tt.wantErr {
					t.Errorf("RandomElement() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					found := slices.Contains(s, got)
					if !found {
						t.Errorf("RandomElement() = %v, not in slice %v", got, s)
					}
				}
			}
		})
	}
}
