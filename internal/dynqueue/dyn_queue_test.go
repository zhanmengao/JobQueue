package dynqueue

import (
	"context"
	"fmt"
	"forevernine.com/midplat/base_libs/queue/qtyp"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	queue *DynQueue
	id    int64
	datas [7]int
)

func TestMain(m *testing.M) {
	queue = NewDynQueue(context.Background(), qtyp.DefaultOption, qtyp.DefaultMonitor)
	if queue == nil {
		panic("fail to new queue")
	}
	m.Run()
}

func TestPush(t *testing.T) {
	tm1 := time.Now()
	ctx := context.Background()
	queue.PushJob(ctx, "1", func(ctx context.Context) {
		diff := time.Now().Sub(tm1)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[1] += 2
	})

	tm2 := time.Now()
	queue.PushJob(ctx, "3", func(ctx context.Context) {
		diff := time.Now().Sub(tm2)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[2] += 5
	})

	tm3 := time.Now()
	queue.PushJob(ctx, "1", func(ctx context.Context) {
		diff := time.Now().Sub(tm3)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[3] += 4
	})

	tm4 := time.Now()
	queue.PushJob(ctx, "3", func(ctx context.Context) {
		diff := time.Now().Sub(tm4)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[4] += 7
	})

	tm5 := time.Now()
	queue.PushJob(ctx, "5", func(ctx context.Context) {
		diff := time.Now().Sub(tm5)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[5] += 3
	})

	tm6 := time.Now()
	queue.PushJob(ctx, "6", func(ctx context.Context) {
		diff := time.Now().Sub(tm6)
		log.Printf("seq:%d, time is %s\n", atomic.AddInt64(&id, 1), diff.String())
		datas[6] += 1
	})
	time.Sleep(1 * time.Second)
	t.Log(datas)
}

func TestDynQueue(t *testing.T) {
	ctx := context.Background()
	var count int32 = 0
	addWg := sync.WaitGroup{}
	addWg.Add(10000 * 1000)
	begin := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				err := queue.PushJob(ctx, Int2Str(int64(j%1000)), func(ctx context.Context) {
					atomic.AddInt32(&count, 1)
				})
				if err != nil {
					t.Error(err.Error())
				}
				addWg.Done()
			}
		}()
	}
	addWg.Wait()
	queue.Close()
	use := time.Now().Sub(begin)
	fmt.Println(count, use.String())
	if count != int32(10000*1000) {
		t.Fail()
	}
	err := queue.PushJob(ctx, "1", func(ctx context.Context) {

	})
	if err != nil {
		t.Log(err)
	}
}

func Int2Str(uid int64) string {
	s := strconv.FormatInt(uid, 10)
	return s
}
