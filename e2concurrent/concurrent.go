// 併發執行
// 使用範例
//package main
//
//import (
//	"fmt"
//	"github.com/e2u/e2util/e2crypto"
//	"sync"
//	"time"
//)
//
//func main() {
//	c := New(10)
//
//	go c.RangeResult(func(o *Output) {
//		fmt.Printf("args: %v, result: %v\n", o.Args, o.Result)
//	})
//
//	for i := int64(0); i <= 100; i++ {
//		fmt.Println("result: >>>>>", i)
//		go c.Run(i, func(args interface{}) *Output {
//			ri := e2crypto.RandomUint(1, 8)
//			time.Sleep(time.Duration(ri) * time.Second)
//			rs := &Output{
//				Args:   args,
//				Result: args.(int64) + ri,
//				Error:  nil,
//			}
//			return rs
//		})
//	}
//
//	c.Wait()
//}

package e2concurrent

import (
	"sync"
)

type Output struct {
	Args   interface{}
	Result interface{}
	Error  error
}

type Concurrent struct {
	wg            sync.WaitGroup
	maxConcurrent int
	concurrentCtl chan struct{} // 用於控制併發的 channel
	outputChannel chan *Output  // 結果 channel
}

func New(mc int) *Concurrent {
	c := &Concurrent{
		wg:            sync.WaitGroup{},
		concurrentCtl: make(chan struct{}, mc),
		outputChannel: make(chan *Output, mc),
		maxConcurrent: mc,
	}
	return c
}

func (c *Concurrent) Wait() {
	c.wg.Wait()
	close(c.outputChannel)
	close(c.concurrentCtl)
}

func (c *Concurrent) RangeResult(fc func(*Output)) {
	for o := range c.outputChannel {
		c.wg.Done()
		<-c.concurrentCtl
		fc(o)
	}
}

func (c *Concurrent) Run(args interface{}, fc func(interface{}) *Output) {
	c.wg.Add(1)
	c.concurrentCtl <- struct{}{}
	go func(args interface{}) {
		c.outputChannel <- fc(args)
	}(args)
}
