package dynqueue

import (
	"context"
	"errors"
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"github.com/zhanmengao/JobQueue/qtyp"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// DynQueue 基于用户个数的动态队列
type DynQueue struct {
	m      map[string]*userQueue
	lock   sync.RWMutex
	wg     sync.WaitGroup
	isOpen bool
	em     *qtyp.EventMonitor
	opt    *qtyp.Option
	ctx    context.Context
}

// NewDynQueue 创建queue，参数size没有实际用途
func NewDynQueue(ctx context.Context, opt *qtyp.Option, em *qtyp.EventMonitor) *DynQueue {
	q := &DynQueue{
		m:      make(map[string]*userQueue, opt.Size),
		isOpen: false,
		em:     em,
		opt:    opt,
		ctx:    ctx,
	}
	q.run()
	return q
}

func (q *DynQueue) run() {
	q.isOpen = true
	go q.cron()
}

// cron 定期检查超过一定时间没有请求的队列，并且清理掉
func (q *DynQueue) cron() {
	for {
		q.clearExpiredQueue()
		time.Sleep(30 * time.Second)
	}
}

// clearExpiredQueue 清理过期队列
func (q *DynQueue) clearExpiredQueue() {
	q.lock.Lock()
	defer q.lock.Unlock()
	now := time.Now()
	for key, uq := range q.m {
		if now.Sub(uq.getLastTime()) >= 3*time.Minute {
			log.Printf("main cron send closing signal to key:%s, interval:%s", key, now.Sub(uq.lastTime).String())
			delete(q.m, key)
			close(uq.closeing)
		}
	}
}

// PushJob 往队列里面加入任务
func (q *DynQueue) PushJob(ctx context.Context, key string, job func(ctx context.Context)) (err error) {
	if !q.isOpen {
		err = errors.New("queue stop")
		q.em.PushEvent(ctx, key, qtyp.EventClosed)
		return
	}
	q.wg.Add(1)
	atomic.AddUint64(&jobIncomingCnt, 1)
	q.lock.RLock()
	uq, ok := q.m[key]
	q.lock.RUnlock()
	// fast path
	if ok {
		uq.push(ctx, &qutil.TJob{
			Func: job,
			Ctx:  ctx,
		})
		return
	}

	// slow path
	q.lock.Lock()
	uq, ok = q.m[key]
	if !ok {
		uq = newUserQueue(ctx, key, &q.wg, q.em, q.opt.Size)
		q.m[key] = uq
	}
	q.lock.Unlock()
	uq.push(ctx, &qutil.TJob{
		Func: job,
		Ctx:  ctx,
	})
	return
}

func (q *DynQueue) Close() {
	q.lock.Lock()
	q.isOpen = false
	for key, uq := range q.m {
		delete(q.m, key)
		close(uq.closeing)
	}
	q.lock.Unlock()
	qutil.CloseWait(q.opt.CloseWait, &q.wg)
}

// Size 返回当前的用户队列个数
func (q *DynQueue) Size() (qlen, qcount int) {
	return len(q.m), int(queueCnt)
}

func (q *DynQueue) JobSize() (inCnt, consumeCnt uint64) {
	inCnt = atomic.LoadUint64(&jobIncomingCnt)
	consumeCnt = atomic.LoadUint64(&jobConsumeCnt)
	return
}

// GetKeyList 返回队列中的所有key
func (q *DynQueue) GetKeyList() []string {
	q.lock.RLock()
	defer q.lock.RUnlock()
	ret := make([]string, 0, len(q.m))
	for key := range q.m {
		ret = append(ret, key)
	}
	return ret
}
