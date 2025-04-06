package antsqueue

import (
	"context"
	"errors"
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"github.com/zhanmengao/JobQueue/qtyp"
	"runtime/debug"
	"sync"
	"time"
)

type AntsJobQueue struct {
	queue *ants.PoolWithFunc
	job   chan *UserRequestList //放user的channel
	close chan struct{}
	wg    sync.WaitGroup

	isOpen bool
	em     *qtyp.EventMonitor
	//user map
	userMap sync.Map
	opt     *qtyp.Option
	tm      *time.Timer
	ctx     context.Context
}

func NewAntsJobQueue(ctx context.Context, opt *qtyp.Option, m *qtyp.EventMonitor) *AntsJobQueue {
	ret := &AntsJobQueue{
		job:    make(chan *UserRequestList, opt.Size),
		close:  make(chan struct{}, 0),
		isOpen: true,
		em:     m,
		opt:    opt,
		ctx:    ctx,
	}
	opts := make([]ants.Option, 0)
	opts = append(opts, ants.WithPanicHandler(ret.panicHandler))
	if opt.NonBlock {
		opts = append(opts, ants.WithNonblocking(true))
	}
	ret.queue, _ = ants.NewPoolWithFunc(opt.Size, ret.do, opts...)
	ret.run(ctx)
	return ret
}

func (q *AntsJobQueue) do(iUser interface{}) {
	//实际执行任务的地方
	u := iUser.(*UserRequestList)
	for {
		select {
		case f := <-u.request:
			f.Func(f.Ctx)
		default:
			return
		}
	}
}

func (q *AntsJobQueue) run(ctx context.Context) {
	if q.em != nil {
		q.em.Run()
	}
	q.isOpen = true
	go func() {
		defer q.queue.Release()
		q.tm = time.NewTimer(time.Duration(5) * time.Second)
		for {
			select {
			case u := <-q.job:
				//注意这里取的是user，也就是说只要阻塞的用户不多，是不会卡到其他用户的
				if err := q.queue.Invoke(u); err != nil {
					q.em.PushEvent(ctx, u.key, qtyp.EventBusy)
				}
			case <-q.close:
				//取出所有消息，处理完return
				q.clear()
				q.wg.Wait()
			case <-q.tm.C:
				q.tm = time.NewTimer(time.Duration(5) * time.Second)
			}
		}
	}()
}

func (q *AntsJobQueue) PushJob(ctx context.Context, key string, f func(ctx2 context.Context)) (err error) {
	if !q.isOpen {
		q.em.PushEvent(ctx, key, qtyp.EventClosed)
		err = errors.New("queue stop")
		return
	}
	u := NewUserRequestList(key, 20)
	if iu, exist := q.userMap.LoadOrStore(key, u); exist {
		u = iu.(*UserRequestList)
	} else {
		u.Init()
	}
	err = q.pushJob(u, &qutil.TJob{
		Func: f,
		Ctx:  ctx,
	})
	return
}

func (q *AntsJobQueue) pushJob(u *UserRequestList, f *qutil.TJob) (err error) {
	//先放进该用户的请求列表
	if q.opt.NonBlock {
		select {
		case u.request <- f:
		default:
			err = errors.New("busy")
			q.em.PushEvent(f.Ctx, u.key, qtyp.EventBusy)
			return
		}
	} else {
		u.request <- f
	}

	//把该用户放进队列的请求列表
	if q.opt.NonBlock {
		select {
		case q.job <- u:
		default:
			err = errors.New("busy ")
			q.em.PushEvent(f.Ctx, u.key, qtyp.EventBusy)
			return
		}
	} else {
		q.job <- u
	}

	return
}

func (q *AntsJobQueue) clearUser() {
	now := time.Now().Unix()
	q.userMap.Range(func(iUID interface{}, iUser interface{}) bool {
		user := iUser.(*UserRequestList)
		if now-user.lastActive > 5 {
			q.userMap.Delete(iUID)
		}
		return true
	})
}

func (q *AntsJobQueue) Close() {
	if q.em != nil {
		defer q.em.Close()
	}
	q.isOpen = false
	if q.close != nil {
		q.close <- struct{}{}
	}
	qutil.CloseWait(q.opt.CloseWait, &q.wg)
}

func (q *AntsJobQueue) clear() {
	for {
		select {
		case u := <-q.job:
			//如果是非阻塞模式，这里会返回err
			if err := q.queue.Invoke(u); err != nil {
				q.em.PushEvent(q.ctx, u.key, qtyp.EventBusy)
			}
		default:
			close(q.job)
			return
		}
	}
}

func (q *AntsJobQueue) panicHandler(iu interface{}) {
	q.em.PushPanicEvent(q.ctx, "", string(debug.Stack()))
}
