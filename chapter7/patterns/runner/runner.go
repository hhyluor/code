package runner

// Example is provided with help by Gabriel Aszalos. // Gabriel Aszalos 协助完成了这个示例
// Package runner manages the running and lifetime of a process. // runner 包管理处理任务的运行和生命周期
import (
	"errors"
	"os"
	"os/signal"
	"time"
)

// Runner runs a set of tasks within a given timeout and can be //  Runner 在给定的超时时间内执行一组任务，
// shut down on an operating system interrupt. // 并且在操作系统发送中断信号时结束这些任务.
type Runner struct {
	// interrupt channel reports a signal from the // interrupt  通道报告从操作系统
	// operating system. // 发送的信号.
	interrupt chan os.Signal

	// complete channel reports that processing is done. // complete 通道报告处理任务已经完成.
	complete chan error

	// timeout reports that time has run out. // timeout 报告处理任务已经超时.
	timeout <-chan time.Time

	// tasks holds a set of functions that are executed // tasks 持有一组以索引顺序依次执行的
	// synchronously in index order.
	tasks []func(int)
}

// ErrTimeout is returned when a value is received on the timeout channel. // ErrTimeout 会在任务执行超时时返回
var ErrTimeout = errors.New("received timeout")

// ErrInterrupt is returned when an event from the OS is received.  // ErrInterrupt 会在接收到操作系统的事件时返回
var ErrInterrupt = errors.New("received interrupt")

// New returns a new ready-to-use Runner. // New 返回一个新的准备使用的 Runner
func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

// Add attaches tasks to the Runner. A task is a function that  // Add 将一个任务附加到 Runner 上。这个任务是一个
// takes an int ID. // 接收一个 int 类型的 ID 作为参数的函数
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

// Start runs all tasks and monitors channel events. // Start 执行所有任务，并监视通道事件
func (r *Runner) Start() error {
	// We want to receive all interrupt based signals. // 我们希望接收所有中断信号
	signal.Notify(r.interrupt, os.Interrupt)

	// Run the different tasks on a different goroutine. // 用不同的 goroutine 执行不同的任务
	go func() {
		r.complete <- r.run()
	}()

	select {
	// Signaled when processing is done. // 用不同的 goroutine 执行不同的任务
	case err := <-r.complete:
		return err

	// Signaled when we run out of time. // 当任务处理程序运行超时时发出的信号
	case <-r.timeout:
		return ErrTimeout
	}
}

// run executes each registered task. // 当任务处理程序运行超时时发出的信号
func (r *Runner) run() error {
	for id, task := range r.tasks {
		// Check for an interrupt signal from the OS. // 检测操作系统的中断信号
		if r.gotInterrupt() {
			return ErrInterrupt
		}

		// Execute the registered task. // 执行已注册的任务
		task(id)
	}

	return nil
}

// gotInterrupt verifies if the interrupt signal has been issued. // gotInterrupt 验证是否接收到了中断信号
func (r *Runner) gotInterrupt() bool {
	select {
	// Signaled when an interrupt event is sent. // 当中断事件被触发时发出的信号
	case <-r.interrupt:
		// Stop receiving any further signals. // 停止接收后续的任何信号
		signal.Stop(r.interrupt)
		return true

	// Continue running as normal. // 继续正常运行
	default:
		return false
	}
}
