package e2sync

import (
	"fmt"
	"testing"
)

func Test_SortedMap(t *testing.T) {
	rbt := NewRBTreeMap[string, string](StringComparer{})
	rbt.Store("A", "seven")
	rbt.Store("D", "three")
	rbt.Store("E", "eighteen")
	rbt.Store("Z", "ten")
	rbt.Store("EFE", "twenty-two")
	rbt.Store("abc", "eight")
	rbt.Store("zzz", "eleven")

	// InOrderTraversal（固定正序）
	fmt.Println("InOrderTraversal (Ascending):")
	rbt.InOrderTraversal(func(key, value string) {
		fmt.Printf("Key: %v, Value: %v\n", key, value)
	})

	// RangeAscending 正序
	fmt.Println("\nRangeAscending:")
	rbt.RangeAscending(func(key, value string) bool {
		fmt.Printf("Key: %v, Value: %v\n", key, value)
		return true
	})

	// RangeDescending 逆序
	fmt.Println("\nRangeDescending:")
	rbt.RangeDescending(func(key, value string) bool {
		fmt.Printf("Key: %v, Value: %v\n", key, value)
		return true
	})

	// 測試查找
	if val, ok := rbt.Load("Z"); ok {
		fmt.Printf("\nFound key Z with value: %v\n", val)
	}
}

func Benchmark_RBTreeMap(b *testing.B) {
	for b.Loop() {
		rbt := NewRBTreeMap[string, string](StringComparer{})
		rbt.Store("A", "seven")
		rbt.Store("D", "three")
		rbt.Store("E", "eighteen")
		rbt.Store("Z", "ten")
		rbt.Store("EFE", "twenty-two")
		rbt.Store("abc", "eight")
		rbt.Store("zzz", "eleven")

		rbt.InOrderTraversal(func(key, value string) {})
		rbt.RangeAscending(func(key, value string) bool { return true })
		rbt.RangeDescending(func(key, value string) bool { return true })
		if val, ok := rbt.Load("Z"); ok {
			_ = val
		}
	}
}

func Benchmark_RBTreeMap_Store(b *testing.B) {
	rbt := NewRBTreeMap[string, string](StringComparer{})
	b.ResetTimer()
	for b.Loop() {
		rbt.Store(fmt.Sprintf("key%d", b.N), "value")
	}
}

func Benchmark_RBTreeMap_Load(b *testing.B) {
	rbt := NewRBTreeMap[string, string](StringComparer{})
	for i := range 1000 {
		rbt.Store(fmt.Sprintf("key%d", i), "value")
	}
	b.ResetTimer()
	for b.Loop() {
		rbt.Load("key500")
	}
}

func Benchmark_RBTreeMap_RangeAscending(b *testing.B) {
	rbt := NewRBTreeMap[string, string](StringComparer{})
	for i := range 1000 {
		rbt.Store(fmt.Sprintf("key%d", i), "value")
	}
	b.ResetTimer()
	for b.Loop() {
		rbt.RangeAscending(func(key, value string) bool { return true })
	}
}
