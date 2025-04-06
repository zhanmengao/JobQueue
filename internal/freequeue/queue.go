package freequeue

import (
	"context"
	"errors"
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"github.com/zhanmengao/JobQueue/qtyp"
	"strconv"
	"sync"
)

type FreeQueue struct {
	worker  []*tWorker
	isOpen  bool
	closeCh chan struct{}
	closeWg sync.WaitGroup
	em      *qtyp.EventMonitor
	opt     *qtyp.Option
	ctx     context.Context
}

func NewFreeQueue(ctx context.Context, opt *qtyp.Option, em *qtyp.EventMonitor) *FreeQueue {
	fq := &FreeQueue{
		worker:  make([]*tWorker, opt.Size),
		closeCh: make(chan struct{}, opt.Size),
		isOpen:  true,
		em:      em,
		opt:     opt,
		ctx:     ctx,
	}
	for i, _ := range fq.worker {
		fq.worker[i] = NewWorker(fq.opt, em)
	}
	fq.run()
	return fq
}

func (fq *FreeQueue) run() {
	if fq.em != nil {
		fq.em.Run()
	}
	for i, _ := range fq.worker {
		go fq.worker[i].run(fq.closeCh)
	}
}

func (fq *FreeQueue) PushJob(ctx context.Context, key string, f func(ctx context.Context)) (err error) {
	if !fq.isOpen {
		fq.em.PushEvent(ctx, key, qtyp.EventClosed)
		err = errors.New("queue stop")
		return
	}
	var busy bool
	if busy, err = fq.worker[fq.getHashKey(key)].Push(&qutil.TJob{
		Func: f,
		Key:  key,
		Ctx:  ctx,
	}); busy {
		fq.em.PushEvent(ctx, key, qtyp.EventBusy)
	}
	return
}

func (fq *FreeQueue) Close() {
	if fq.em != nil {
		defer fq.em.Close()
	}
	fq.isOpen = false
	for _, _ = range fq.worker {
		fq.closeCh <- struct{}{}
	}
	wgList := make([]*sync.WaitGroup, 0, len(fq.worker))
	for _, w := range fq.worker {
		wgList = append(wgList, &w.wg)
	}
	qutil.CloseWait(fq.opt.CloseWait, wgList...)

}

func (fq *FreeQueue) getHashKey(uid string) int32 {
	uid64, _ := strconv.ParseInt(uid, 10, 64)
	return int32(uid64 % int64(fq.opt.Size))
}
