package antsqueue

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

func TestAntsQueue(t *testing.T) {
	var count int32 = 0
	ctx := context.Background()
	q := NewAntsJobQueue(ctx, qtyp.DefaultOption, qtyp.DefaultMonitor)
	addWg := sync.WaitGroup{}
	addWg.Add(10000 * 1000)
	begin := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				if err := q.PushJob(ctx, strconv.Itoa(j%1000), func(ctx context.Context) {
					atomic.AddInt32(&count, 1)
				}); err != nil {
					log.Println(err)
				}
				addWg.Done()
			}
		}()
	}
	addWg.Wait()
	q.Close()
	if err := q.PushJob(ctx, "1", func(context.Context) {}); err == nil {
		t.Fatal("queue is closed")
	}
	use := time.Now().Sub(begin)
	fmt.Println(count, use.String())
	if count != int32(10000*1000) {
		t.Fail()
	}
}
