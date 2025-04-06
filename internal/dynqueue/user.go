package dynqueue

import (
	"container/list"
	"context"
	"errors"
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"github.com/zhanmengao/JobQueue/qtyp"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var (
	queueCnt       int32
	jobIncomingCnt uint64
	jobConsumeCnt  uint64
)

// userQueue 每个用户一个队列
type userQueue struct {
	sync.RWMutex               // 用于数据安全
	key          string        // 队列的key
	jobList      list.List     // 队列中的回调
	lastTime     time.Time     // 上次的最后一次入队列时间
	incoming     chan struct{} // 队列有消息
	closeing     chan struct{} // 队列退出
	wg           *sync.WaitGroup
	em           *qtyp.EventMonitor
	sz           int
}

func newUserQueue(ctx context.Context, key string, wg *sync.WaitGroup, em *qtyp.EventMonitor, sz int) *userQueue {
	q := &userQueue{
		key:      key,
		lastTime: time.Now(),
		incoming: make(chan struct{}, 1),
		closeing: make(chan struct{}),
		wg:       wg,
		em:       em,
		sz:       sz,
	}
	q.jobList.Init()

	go q.run(ctx)

	atomic.AddInt32(&queueCnt, 1)

	return q
}

// push 往队列加入任务
func (uq *userQueue) push(ctx context.Context, job *qutil.TJob) (err error) {
	uq.Lock()
	defer uq.Unlock()
	//检测job上限，如果超过5000，就不让push了
	if uq.jobList.Len() >= uq.sz {
		err = errors.New("busy")
		log.Printf("user queue busy .len = %d", uq.jobList.Len())
		return
	}
	uq.jobList.PushBack(job)
	uq.setLastTime(time.Now())
	// 尝试通知worker消费
	select {
	case uq.incoming <- struct{}{}:
	default: //do nothing
	}
	return
}

// poll 找出所有已经入队列的任务
func (uq *userQueue) poll() []*qutil.TJob {
	uq.Lock()
	defer uq.Unlock()
	sz := uq.jobList.Len()
	tmp := make([]*qutil.TJob, 0, sz)
	for uq.jobList.Len() != 0 {
		elem := uq.jobList.Front()
		uq.jobList.Remove(elem)
		tmp = append(tmp, elem.Value.(*qutil.TJob))
	}
	return tmp
}

// run 用户队列的loop函数
func (uq *userQueue) run(ctx context.Context) {
	defer func() {
		close(uq.incoming)
		uq.closeing = nil
		uq.incoming = nil
	}()
	for {
		select {
		case <-uq.incoming:
			jobs := uq.poll()
			for _, job := range jobs {
				atomic.AddUint64(&jobConsumeCnt, 1)
				job.Func(job.Ctx)
				uq.wg.Done()
			}
		case <-uq.closeing:
			log.Printf("worker recieve an exit signal, key:%s", uq.key)
			atomic.AddInt32(&queueCnt, -1)
			return
		}
	}
}

func (uq *userQueue) doAllJob() {
	jobs := uq.poll()
	for _, f := range jobs {
		qtyp.Safe(f.Ctx, uq.key, f.Func, uq.em)
		uq.wg.Done()
	}
}

func (uq *userQueue) getLastTime() time.Time {
	uq.RLock()
	defer uq.RUnlock()
	return uq.lastTime
}

func (uq *userQueue) setLastTime(tm time.Time) {
	uq.lastTime = tm
}
