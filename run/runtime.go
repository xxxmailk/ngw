package run

import (
	"github.com/sirupsen/logrus"
	"ngw/log"
	"reflect"
)

type Handle func(pool, rbd string, args ...string) error
type Runner struct {
	pool        string           // 要操作的pool名称
	rbd         string           // 同上
	flows       map[int]Handle   // 流程步骤函数
	rollback    map[int]Handle   // 对应流程的回滚函数
	cap         int              // 总共注册了多少步骤
	offset      int              // 当前执行到哪个步骤
	l           *logrus.Entry    // 日志handle
	close       chan int         // close channel
	runMessages map[int]string   //执行时显示消息
	bakMessages map[int]string   //回滚时显示消息
	args        map[int][]string //存储参数
}

// 注册要运行的函数
func (r *Runner) Register(
	handle, rollback func(pool, rbd string, args ...string) error,
	rMessage string,
	bMessage string,
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
	r.l.Debugf("registering function %s, function id is %d", reflect.TypeOf(handle).Name(), r.cap)
}

// 开始运行
func (r *Runner) Run() {
	var err error
	for i := 0; i < r.cap; i++ {
		r.l.Info(r.runMessages)
		// 执行函数
		err = r.flows[i](r.pool, r.rbd, r.args[i]...)
		// 如果任何一个步骤执行失败，则回滚
		if err != nil {
			r.l.Error(err)
			r.rollBack()
			return
		}
		r.offset += 1
	}
}

// 失败回滚
func (r *Runner) rollBack() {
	var err error
	for i := r.offset; i > 0; i-- {
		r.l.Info(r.bakMessages[i])
		// 执行回滚条目的前一条,当前条目没有执行成功，不回滚
		err = r.rollback[i-1](r.pool, r.rbd, r.args[i]...)
		if err != nil {
			r.l.Error(err)
			return
		}
	}
}

// 统一使用该函数进行申明
func NewRunner(pool, rbd string) Runner {
	// 带缓冲channel,  有顺序输出的同时避免阻塞
	r := new(Runner)
	// close channel
	r.close = make(chan int)
	r.pool = pool
	r.rbd = rbd
	r.l = log.GetLogger()
	return *r
}
