// Example provided with help from Jason Waldrip.
// Package work manages a pool of goroutines to perform work.
// Jason Waldrip 协助完成了这个示例
// work 包管理一个 goroutine 池来完成工作
package work

import "sync"

// Worker must be implemented by types that want to use
// the work pool.
// Worker 必须满足接口类型，
// 才能使用工作池
type Worker interface {
	Task()
}

// Pool provides a pool of goroutines that can execute any Worker
// tasks that are submitted.
// Pool 提供一个 goroutine 池， 这个池可以完成
// 任何已提交的 Worker 任务
type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

// New creates a new work pool.
// New 创建一个新工作池
func New(maxGoroutines int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}

	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				w.Task()
			}
			p.wg.Done()
		}()
	}

	return &p
}

// Run submits work to the pool.
// Run 提交工作到工作池
func (p *Pool) Run(w Worker) {
	p.work <- w
}

// Shutdown waits for all the goroutines to shutdown.
// Shutdown 等待所有 goroutine 停止工作
func (p *Pool) Shutdown() {
	close(p.work)
	p.wg.Wait()
}
