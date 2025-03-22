package e2sync

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func Test_Add(t *testing.T) {
	var m sync.Map
	for i := 0; i < 10; i++ {
		v, _ := Add[int](&m, "abc", -1)
		//t.Log(v)
		_ = v
	}

}

func Benchmark_Add(b *testing.B) {
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Add[int](&m, "abc", -1)
		}
	})
	//b.Log(m.Load("abc"))
}

func Benchmark_Add_Parallel(b *testing.B) {
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			Add[int](&m, fmt.Sprintf("key%d", i), -1)
			i++
		}
	})
}

func Benchmark_AtomicAdd(b *testing.B) {
	var value atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			value.Add(-1)
		}
	})
	//b.Log(value.Load())
}

func Benchmark_Add_Parallel_Optimized(b *testing.B) {
	var m sync.Map
	var counter int64
	b.RunParallel(func(pb *testing.PB) {
		id := atomic.AddInt64(&counter, 1) // 每個 goroutine 獲取唯一 ID
		for pb.Next() {
			Add[int](&m, fmt.Sprintf("key%d", id), -1)
		}
	})
}
