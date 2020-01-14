package run

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"ngw/log"
	"time"
)

type Message struct {
	Type int    // 0 -> info  1 -> successfully 2 -> failed
	Msg  string // print messages
}

type Handle func(pool, rbd string, speaker chan Message, args ...string) error
type Runner struct {
	pool        string           // 要操作的pool名称
	rbd         string           // 同上
	flows       map[int]Handle   // 流程步骤函数
	rollback    map[int]Handle   // 对应流程的回滚函数
	cap         int              // 总共注册了多少步骤
	offset      int              // 当前执行到哪个步骤
	l           *logrus.Entry    // 日志handle
	speaker     chan Message     // 叙述者——不断从buffer中输出消息
	runMessages map[int]Message  //执行时显示消息
	bakMessages map[int]Message  //回滚时显示消息
	args        map[int][]string //存储参数
}

// 注册要运行的函数
func (r *Runner) Register(
	handle, rollback func(pool, rbd string, speaker chan Message, args ...string) error,
	rMessage Message,
	bMessage Message,
	args ...string) {
	if handle == nil || rollback == nil {
		logrus.Errorf("cannot be register an nil function")
		return
	}
	r.flows[r.cap] = handle
	r.rollback[r.cap] = rollback
	r.cap += 1
	r.args[r.cap] = args
	r.runMessages[r.cap] = rMessage
	r.bakMessages[r.cap] = bMessage
	r.l.Debugf("registering function %v, function id is %d", handle, r.cap)
}

// 开始运行
func (r *Runner) Run() {
	var err error
	for i := 0; i < r.cap; i++ {
		r.speaker <- r.runMessages[i]
		// 执行函数
		err = r.flows[i](r.pool, r.rbd, r.speaker, r.args[i]...)
		// 如果任何一个步骤执行失败，则回滚
		if err != nil {
			r.speaker <- Message{
				Type: 2,
				Msg:  err.Error(),
			}
			r.rollBack()
			return
		}
		r.offset += 1
	}
}

// 进度条
// stop channel: 有消息时中断进度条
// c channel: 收到退出程序消息时中断进度条
func (r Runner) progress(stop chan int, c chan int) {
	var output []byte
	max := 30
	offset := 1
	tick := time.NewTicker(time.Duration(time.Second * 1))
	defer tick.Stop()
	for {
		select {
		case <-stop:
			return
		case <-c:
			return
		case <-tick.C:
		}

		if offset == max {
			offset = 1
		} else {
			offset += 1
		}
		// ■
		// =
		for i := 0; i <= offset; i++ {
			output = append(output, '=')
		}
		output = append(output, '>')
		fmt.Print("\r")
		fmt.Printf("[%-30s]", output)
	}

}

// 叙述者——异步将程序输出格式化后打印到屏幕上
func (r Runner) Speaker(c chan int) {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()
	// stop channel for progress bar
	stop := make(chan int, 1)
	defer close(stop)
	for {
		select {
		// 不做default通道，让进城一直阻塞在这里，直到有消息过来
		case msg := <-r.speaker:
			stop <- 1
			fmt.Printf("\r")
			fmt.Printf("[NGW]: %s\n", msg)
			go r.progress(stop, c)
		// 收到close信号，退出进程
		case <-c:
			return
		}
	}
}

// 失败回滚
func (r *Runner) rollBack() {
	var err error
	for i := r.offset; i > 0; i-- {
		r.speaker <- r.bakMessages[i]
		// 执行回滚条目的前一条,当前条目没有执行成功，不回滚
		err = r.rollback[i-1](r.pool, r.rbd, r.speaker, r.args[i]...)
		if err != nil {
			r.speaker <- Message{
				Type: 2,
				Msg:  err.Error(),
			}
			return
		}
	}
}

// 统一使用该函数进行申明
func NewRunner(pool, rbd string) Runner {
	// 带缓冲channel,  有顺序输出的同时避免阻塞
	r := new(Runner)
	r.pool = pool
	r.rbd = rbd
	r.speaker = make(chan Message, 10)
	r.l = log.GetLogger()
	return *r
}
