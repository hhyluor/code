// Example provided with help from Fatih Arslan and Gabriel Aszalos. // Fatih Arslan 和 Gabriel Aszalos 协助完成了这个示例
// Package pool manages a user defined set of resources. // 包 pool 管理用户定义的一组资源
package pool

import (
	"errors"
	"io"
	"log"
	"sync"
)

// Pool manages a set of resources that can be shared safely by
// multiple goroutines. The resource being managed must implement
// the io.Closer interface.
// Pool 管理一组可以安全地在多个 goroutine 间
// 共享的资源。被管理的资源必须
// 实现 io.Closer 接口
type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   func() (io.Closer, error)
	closed    bool
}

// ErrPoolClosed is returned when an Acquire returns on a
// closed pool.
// ErrPoolClosed 表示请求（Acquire）了一个
// 已经关闭的池
var ErrPoolClosed = errors.New("Pool has been closed.")

// New creates a pool that manages resources. A pool requires a
// function that can allocate a new resource and the size of
// the pool.
// New 创建一个用来管理资源的池。
// 这个池需要一个可以分配新资源的函数，
// 并规定池的大小
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size value too small.")
	}

	return &Pool{
		factory:   fn,
		resources: make(chan io.Closer, size),
	}, nil
}

// Acquire retrieves a resource	from the pool.
// Acquire 从池中获取一个资源
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	// Check for a free resource.
	// 检查是否有空闲的资源
	case r, ok := <-p.resources:
		log.Println("Acquire:", "Shared Resource")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil

	// Provide a new resource since there are none available.
	// 因为没有空闲资源可用，所以提供一个新资源
	default:
		log.Println("Acquire:", "New Resource")
		return p.factory()
	}
}

// Release places a new resource onto the pool.
// Release 将一个使用后的资源放回池里
func (p *Pool) Release(r io.Closer) {
	// Secure this operation with the Close operation.
	// 保证本操作和 Close 操作的安全
	p.m.Lock()
	defer p.m.Unlock()

	// If the pool is closed, discard the resource.
	// 如果池已经被关闭，销毁这个资源
	if p.closed {
		r.Close()
		return
	}

	select {
	// Attempt to place the new resource on the queue.
	// 试图将这个资源放入队列
	case p.resources <- r:
		log.Println("Release:", "In Queue")

	// If the queue is already at cap we close the resource.
	// 如果队列已满，则关闭这个资源
	default:
		log.Println("Release:", "Closing")
		r.Close()
	}
}

// Close will shutdown the pool and close all existing resources.
// Close 会让资源池停止工作，并关闭所有现有的资源
func (p *Pool) Close() {
	// Secure this operation with the Release operation.
	// 保证本操作与 Release 操作的安全
	p.m.Lock()
	defer p.m.Unlock()

	// If the pool is already close, don't do anything.
	// 如果 pool 已经被关闭，什么也不做
	if p.closed {
		return
	}

	// Set the pool as closed.
	// 将池关闭
	p.closed = true

	// Close the channel before we drain the channel of its
	// resources. If we don't do this, we will have a deadlock.
	// 在清空通道里的资源之前，将通道关闭
	// 如果不这样做，会发生死锁
	close(p.resources)

	// Close the resources
	// 关闭资源
	for r := range p.resources {
		r.Close()
	}
}
