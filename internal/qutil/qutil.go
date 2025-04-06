package qutil

import (
	"sync"
	"time"
)

func CloseWait(timeoutMill int64, wgList ...*sync.WaitGroup) {
	//阻塞等
	if timeoutMill < 0 {
		for _, wg := range wgList {
			wg.Wait()
		}
	} else {
		closeTm := time.After(time.Duration(timeoutMill) * time.Millisecond)
		select {
		case <-closeWait(wgList...):
		case <-closeTm:
		}
	}
}

func closeWait(wgList ...*sync.WaitGroup) (ch chan struct{}) {
	ch = make(chan struct{}, 1)
	go func() {
		defer func() {
			ch <- struct{}{}
			close(ch)
		}()
		for _, wg := range wgList {
			wg.Wait()
		}
	}()

	return ch
}
