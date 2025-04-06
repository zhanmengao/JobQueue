package qtyp

import "context"

type Event int32

const (
	EventSuccess Event = iota
	EventBusy          //队列繁忙，可以考虑禁用该UID用户提交请求。使用非阻塞队列时Push失败
	EventPanic         //执行中出现了Panic
	EventClosed        //调用了关闭，不允许继续Push
)

type QueueEvent struct {
	Key   string
	Ev    Event
	Stack string
	Ctx   context.Context
}
