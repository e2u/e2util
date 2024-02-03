package e2array

import (
	"sort"
	"strings"
)

func IncludeString(as []string, s string) bool {
	for idx := range as {
		if as[idx] == s {
			return true
		}
	}
	return false
}

func ContainString(as []string, s string) bool {
	for idx := range as {
		if strings.Contains(s, as[idx]) {
			return true
		}
	}
	return false
}

func HasPrefix(as []string, prefix string) bool {
	for idx := range as {
		if strings.HasPrefix(as[idx], prefix) {
			return true
		}
	}
	return false
}

func HasSuffix(as []string, suffix string) bool {
	for idx := range as {
		if strings.HasSuffix(as[idx], suffix) {
			return true
		}
	}
	return false
}

func ContainInt64(is []int64, i int64) bool {
	for idx := range is {
		if is[idx] == i {
			return true
		}
	}
	return false
}

func BoolSliceInclude(ba []bool, c bool) bool {
	for _, b := range ba {
		if b == c {
			return true
		}
	}
	return false
}

func CompareStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func MergeStringSlice(src1, src2 []string) []string {
	var rs []string
	tm := make(map[string]bool)
	for idx := range src1 {
		tm[src1[idx]] = true
	}
	for idx := range src2 {
		tm[src2[idx]] = true
	}
	for k := range tm {
		rs = append(rs, k)
	}
	return rs
}

func UniqStringSlice(sr []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range sr {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func GetDefault[T comparable](arr []T, index int, defaultValue T) T {
	if index >= 0 && index < len(arr) {
		return arr[index]
	}
	return defaultValue
}

func HasConsecutiveNumbers[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](nfs []T) bool {
	nums := make([]T, len(nfs))
	for idx := range nfs {
		nums[idx] = nfs[idx]
	}
	if len(nums) < 2 {
		return false
	}
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	for i := 1; i < len(nums); i++ {
		if nums[i]-nums[i-1] == 1 {
			return true
		}
	}
	return false
}
