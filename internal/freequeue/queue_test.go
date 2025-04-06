package freequeue

import (
	"context"
	"fmt"
	"forevernine.com/midplat/base_libs/queue/qtyp"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	Total = 10000000
	Mod   = 1000000
)

func TestNewOrderJobQueue(t *testing.T) {
	var count int64
	q := NewFreeQueue(context.Background(), qtyp.DefaultOption, qtyp.DefaultMonitor)
	fmt.Printf("begin = %s \n", time.Now().String())
	for i := 0; i < Total; i++ {
		id := rand.Int63() + int64(i)
		q.PushJob(context.Background(), Int2Str(id), func(ctx context.Context) {
			newVal := atomic.AddInt64(&count, 1)
			if newVal%Mod == 0 {
				fmt.Printf("%d ok end = %s \n", count, time.Now().String())
			}
		})
	}
}
func TestFrrQueue(t *testing.T) {
	var count int32 = 0
	q := NewFreeQueue(context.Background(), qtyp.DefaultOption, qtyp.DefaultMonitor)
	addWg := sync.WaitGroup{}
	addWg.Add(10000 * 1000)
	begin := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				err := q.PushJob(context.Background(), Int2Str(int64(j%1000)), func(ctx context.Context) {
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
	q.Close()
	use := time.Now().Sub(begin)
	fmt.Println(count, use.String())
	if count != int32(10000*1000) {
		t.Fail()
	}
	err := q.PushJob(context.Background(), "1", func(ctx context.Context) {
	})
	if err != nil {
		t.Log(err)
	}
}

func Int2Str(uid int64) string {
	s := strconv.FormatInt(uid, 10)
	return s
}
