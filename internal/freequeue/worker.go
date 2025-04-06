package freequeue

import (
	"errors"
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"github.com/zhanmengao/JobQueue/qtyp"
	"sync"
)

type tWorker struct {
	ch  chan *qutil.TJob
	wg  sync.WaitGroup
	em  *qtyp.EventMonitor
	opt *qtyp.Option
}

func NewWorker(opt *qtyp.Option, em *qtyp.EventMonitor) *tWorker {
	return &tWorker{
		ch:  make(chan *qutil.TJob, opt.Size),
		em:  em,
		opt: opt,
	}
}

func (worker *tWorker) Push(job *qutil.TJob) (isBusy bool, err error) {
	if worker.opt.NonBlock {
		worker.wg.Add(1)
		select {
		case worker.ch <- job:
			isBusy = false
		default:
			isBusy = true
			worker.wg.Done()
			err = errors.New("busy")
		}
	} else {
		worker.wg.Add(1)
		worker.ch <- job
		if len(worker.ch) >= int(0.8*float64(cap(worker.ch))) {
			isBusy = true
		}
	}
	return
}

func (worker *tWorker) run(closeCh <-chan struct{}) {
	for {
		select {
		case w := <-worker.ch:
			qtyp.Safe(w.Ctx, w.Key, w.Func, worker.em)
			worker.wg.Done()
		case <-closeCh:
			worker.clear()
			return
		}
	}
}

func (worker *tWorker) Wait() {
	worker.wg.Wait()
}

func (worker *tWorker) clear() {
	defer close(worker.ch)
	for {
		select {
		case w := <-worker.ch:
			qtyp.Safe(w.Ctx, w.Key, w.Func, worker.em)
			worker.wg.Done()
		}
	}
}
