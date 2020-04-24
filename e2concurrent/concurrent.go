// 併發執行
// 使用範例
// package main
//
//import (
//	"fmt"
//	"time"
//
//	"git.panda-fintech.com/golang/e2util/e2concurrent"
//)
//
//func main() {
//	fmt.Println("start")
//
//	c := e2concurrent.New(5)
//	defer c.Close()
//	// 消費處理結果,需要先執行
//	c.RangeOutput(func(oo *e2concurrent.Output) {
//		fmt.Println(oo)
//		time.Sleep(20 * time.Millisecond)
//	})
//
//	// 並行執行處理方法
//	for i := 0; i <= 100; i++ {
//		c.Process(process, i)
//	}
//
//	// 等待最終處理結束
//	c.Wait()
//	fmt.Println("done")
//}
//
//// 處理方法
//func process(input interface{}) *e2concurrent.Output {
//	fmt.Println(">> process", input)
//	time.Sleep(2 * time.Millisecond)
//	return &e2concurrent.Output{
//		Input: input,
//		Value: "aaaa " + time.Now().Format(time.RFC3339),
//	}
//}

package e2concurrent

import (
	"sync"
)

// Output 方法處理結果
type Output struct {
	Input interface{}
	Value interface{}
	Error error
}

type Concurrent struct {
	wg            sync.WaitGroup
	maxConcurrent int           // 最大併發數
	concurrentCtl chan struct{} // 用於控制併發的 channel
	outputChannel chan *Output  // 結果 channel
}

func New(maxConcurrent int) *Concurrent {
	return &Concurrent{
		wg:            sync.WaitGroup{},
		concurrentCtl: make(chan struct{}, maxConcurrent),
		outputChannel: make(chan *Output, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// Process 併發執行傳入的方法， input 為入參,output 為返回值
func (c *Concurrent) Process(f func(interface{}) *Output, payload interface{}) {
	c.wg.Add(1)
	c.concurrentCtl <- struct{}{}
	go func(_payload interface{}) {
		c.outputChannel <- f(_payload) // 傳入的方法輸出結果放入到響應 channel 中
	}(payload)
}

// RangeOutput 獲取輸出結果,如果不把 channel 中的結果取走，則阻塞
func (c *Concurrent) RangeOutput(f func(*Output)) {
	go func() {
		for o := range c.outputChannel {
			c.wg.Done()
			<-c.concurrentCtl
			f(o)
		}
	}()
}

func (c *Concurrent) Wait() {
	c.wg.Wait()
}

func (c *Concurrent) Close() {
	close(c.outputChannel)
	close(c.concurrentCtl)
}
