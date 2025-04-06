package qtyp

import (
	"context"
	"runtime/debug"
)

func Safe(ctx context.Context, key string, f func(ctx context.Context), em *EventMonitor) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			em.PushPanicEvent(ctx, key, stack)
		}
	}()
	f(ctx)
}
