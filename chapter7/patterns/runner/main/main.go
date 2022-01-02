// This sample program demonstrates how to use a channel to // 这个示例程序演示如何使用通道来监视
// monitor the amount of time the program is running and terminate // 程序运行的时间，以在程序运行时间过长
// the program if it runs too long. // 时如何终止程序
package main

import (
	"log"
	"os"
	"time"

	"github.com/goinaction/code/chapter7/patterns/runner"
)

// timeout is the number of second the program has to finish. // timeout 规定了必须在多少秒内处理完成
const timeout = 3 * time.Second

// main is the entry point for the program.  // main 是程序的入口
func main() {
	log.Println("Starting work.")

	// Create a new timer value for this run. // 为本次执行分配超时时间
	r := runner.New(timeout)

	// Add the tasks to be run. // 加入要执行的任务
	r.Add(createTask(), createTask(), createTask())

	// Run the tasks and handle the result. // 执行任务并处理结果
	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Println("Terminating due to timeout.")
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Println("Terminating due to interrupt.")
			os.Exit(2)
		}
	}

	log.Println("Process ended.")
}

// createTask returns an example task that sleeps for the specified // createTask 返回一个根据 id
// number of seconds based on the id.  // 休眠指定秒数的示例任务
func createTask() func(int) {
	return func(id int) {
		log.Printf("Processor - Task #%d.", id)
		time.Sleep(time.Duration(id) * time.Second)
	}
}
