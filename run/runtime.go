package run

import (
	"github.com/sirupsen/logrus"
	"ngw/log"
	"reflect"
	"runtime"
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
	r.args[r.cap] = args
	r.runMessages[r.cap] = rMessage
	r.bakMessages[r.cap] = bMessage
	r.cap += 1
	r.l.Debugf("registering function %s, function id is %d",
		runtime.FuncForPC(reflect.ValueOf(handle).Pointer()).Name(),
		r.cap)
}

func GetFuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// 开始运行
func (r *Runner) Run() {
	var err error
	for i := 0; i < r.cap; i++ {
		r.l.Debugf("running method index: %d", i)
		r.l.Info(r.runMessages[i])
		// 执行函数
		r.l.Debugf("runner index %d", i)
		r.l.Traceln("flows map:", r.flows)
		r.l.Traceln("back map:", r.rollback)
		r.l.Traceln("fMsg map:", r.runMessages)
		r.l.Traceln("bakMsg map:", r.bakMessages)
		r.l.Traceln("args map:", r.args)
		r.l.Debugf("runner give parameter args %s", r.args[i])
		r.l.Debugf("prepare to running function: %s", GetFuncName(r.flows[i]))
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
	r.runMessages = make(map[int]string)
	r.bakMessages = make(map[int]string)
	r.flows = make(map[int]Handle)
	r.rollback = make(map[int]Handle)
	r.args = make(map[int][]string)
	r.rbd = rbd
	r.l = log.GetLogger()
	return *r
}
