// This sample program demonstrates how to use the pool package
// to share a simulated set of database connections.
// 这个示例程序展示如何使用 pool 包
// 来共享一组模拟的数据库连接
package main

import (
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goinaction/code/chapter7/patterns/pool"
)

const (
	maxGoroutines   = 25 // the number of routines to use. // 要使用的 goroutine 的数量
	pooledResources = 2  // number of resources in the pool // 池中的资源的数量
)

// dbConnection simulates a resource to share.
// dbConnection 模拟要共享的资源
type dbConnection struct {
	ID int32
}

// Close implements the io.Closer interface so dbConnection
// can be managed by the pool. Close performs any resource
// release management.
// Close 实现了 io.Closer 接口，以便 dbConnection
// 可以被池管理。 Close 用来完成任意资源的
// 释放管理
func (dbConn *dbConnection) Close() error {
	log.Println("Close: Connection", dbConn.ID)
	return nil
}

// idCounter provides support for giving each connection a unique id.
// idCounter 用来给每个连接分配一个独一无二的 id
var idCounter int32

// createConnection is a factory method that will be called by
// the pool when a new connection is needed.
// createConnection 是一个工厂函数，
// 当需要一个新连接时，资源池会调用这个函数
func createConnection() (io.Closer, error) {
	id := atomic.AddInt32(&idCounter, 1)
	log.Println("Create: New Connection", id)

	return &dbConnection{id}, nil
}

// main is the entry point for all Go programs.
// main 是所有 Go 程序的入口
func main() {
	var wg sync.WaitGroup
	wg.Add(maxGoroutines)

	// Create the pool to manage our connections.
	// 创建用来管理连接的池
	p, err := pool.New(createConnection, pooledResources)
	if err != nil {
		log.Println(err)
	}

	// Perform queries using connections from the pool.
	// 使用池里的连接来完成查询
	for query := 0; query < maxGoroutines; query++ {
		// Each goroutine needs its own copy of the query
		// value else they will all be sharing the same query
		// variable.
		// 每个 goroutine 需要自己复制一份要
		// 查询值的副本，不然所有的查询会共享
		// 同一个查询变量
		go func(q int) {
			performQueries(q, p)
			wg.Done()
		}(query)
	}

	// Wait for the goroutines to finish.
	// 等待 goroutine 结束
	wg.Wait()

	// Close the pool.
	log.Println("Shutdown Program.")
	p.Close()
}

// performQueries tests the resource pool of connections.
// performQueries 用来测试连接的资源池
func performQueries(query int, p *pool.Pool) {
	// Acquire a connection from the pool.
	// 从池里请求一个连接
	conn, err := p.Acquire()
	if err != nil {
		log.Println(err)
		return
	}

	// Release the connection back to the pool.
	// 将该连接释放回池里
	defer p.Release(conn)

	// Wait to simulate a query response.
	// 用等待来模拟查询响应
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	log.Printf("Query: QID[%d] CID[%d]\n", query, conn.(*dbConnection).ID)
}
