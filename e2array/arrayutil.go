package e2array

import (
	"sort"
	"strings"
)

// IncludeString 检查 as 数组中是否包含 s
// 例如: as = {aaa,bbb,ccc},s = aaa 则结果为 true
//  例如: as = {a,b,c},s=aa 则结果为 false
func IncludeString(as []string, s string) bool {
	for idx := range as {
		if as[idx] == s {
			return true
		}
	}
	return false
}

// ContainString 检查 s 是否包含 as 中的词
// 例如: as = {aaa,bbb,ccc} ,s = 45678aaabbb123 则结果为 true
// 例如: as = {aaa,bbb,ccc} ,s = 45678aabb123 则结果为 false
func ContainString(as []string, s string) bool {
	for idx := range as {
		if strings.Contains(s, as[idx]) {
			return true
		}
	}
	return false
}

// HasPrefix 判断是否 as 数组是否包含字符串 prefix 开头的元素
func HasPrefix(as []string, prefix string) bool {
	for idx := range as {
		if strings.HasPrefix(as[idx], prefix) {
			return true
		}
	}
	return false
}

// HasPrefix 判断是否 as 数组是否包含字符串 suffix 结尾的元素
func HasSuffix(as []string, suffix string) bool {
	for idx := range as {
		if strings.HasSuffix(as[idx], suffix) {
			return true
		}
	}
	return false
}

// ContainInt64 检查 as 是否包含 s
func ContainInt64(is []int64, i int64) bool {
	for idx := range is {
		if is[idx] == i {
			return true
		}
	}
	return false
}

// BoolSliceInclude 检查一个bool值是否在 bool 类型的数组中
func BoolSliceInclude(ba []bool, c bool) bool {
	for _, b := range ba {
		if b == c {
			return true
		}
	}
	return false
}

// CompareStringSlice 比较两个字符串数组是否相同
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

// MergeStringSlice 合并两个字符串数组返回
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

// UniqStringSlice 返回不重复的字符串数组
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
