package queue

import (
	"context"
	"github.com/zhanmengao/JobQueue/internal/antsqueue"
	"github.com/zhanmengao/JobQueue/internal/dynqueue"
	"github.com/zhanmengao/JobQueue/internal/freequeue"
	"github.com/zhanmengao/JobQueue/qtyp"
)

type IQueue interface {
	PushJob(ctx context.Context, key string, f func(ctx context.Context)) (err error)
	Close()
}

func NewQueue(ctx context.Context, option *qtyp.Option) (q IQueue) {
	switch option.QType {
	case qtyp.QueueDyn:
		q = dynqueue.NewDynQueue(ctx, option, qtyp.NewEventMonitor(option.RcvEvent))
	case qtyp.QueueQuick:
		q = freequeue.NewFreeQueue(ctx, option, qtyp.NewEventMonitor(option.RcvEvent))
	case qtyp.QueueAnts:
		q = antsqueue.NewAntsJobQueue(ctx, option, qtyp.NewEventMonitor(option.RcvEvent))
	}
	return
}
