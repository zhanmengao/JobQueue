package qtyp

import "context"

type QueueEventCall func(e *QueueEvent)
type EventMonitor struct {
	eventChannel chan *QueueEvent
	closeChannel chan struct{}
	cb           QueueEventCall
	isOpen       bool
}

var DefaultMonitor = NewEventMonitor(defaultEventHandle)

func NewEventMonitor(cb QueueEventCall) *EventMonitor {
	return &EventMonitor{
		eventChannel: make(chan *QueueEvent, 20),
		closeChannel: make(chan struct{}, 1),
		cb:           cb,
	}
}

func (m *EventMonitor) Run() {
	m.isOpen = true
	go func() {
		defer m.close()
		for {
			select {
			case e := <-m.eventChannel:
				m.cb(e)
			case <-m.closeChannel:
				return
			}
		}
	}()
}

func (m *EventMonitor) PushEvent(ctx context.Context, key string, qe Event) {
	if m == nil || !m.isOpen {
		return
	}
	m.eventChannel <- &QueueEvent{
		Ev:  qe,
		Key: key,
		Ctx: ctx,
	}
}

func (m *EventMonitor) PushPanicEvent(ctx context.Context, key, stack string) {
	if m == nil || !m.isOpen {
		return
	}
	m.eventChannel <- &QueueEvent{
		Key:   key,
		Ev:    EventPanic,
		Stack: stack,
		Ctx:   ctx,
	}
}

func (m *EventMonitor) Close() {
	m.isOpen = false
	m.closeChannel <- struct{}{}
}

func (m *EventMonitor) close() {
	defer close(m.eventChannel)
	defer close(m.closeChannel)
}
