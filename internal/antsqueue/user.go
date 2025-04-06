package antsqueue

import (
	"github.com/zhanmengao/JobQueue/internal/qutil"
	"time"
)

type UserRequestList struct {
	request    chan *qutil.TJob
	key        string
	qSize      int
	lastActive int64
}

func NewUserRequestList(key string, sz int) *UserRequestList {
	return &UserRequestList{
		key:   key,
		qSize: sz,
	}
}

func (u *UserRequestList) Init() {
	u.request = make(chan *qutil.TJob, u.qSize)
	u.lastActive = time.Now().Unix()
}

func (u *UserRequestList) Clear() {
	close(u.request)
}
