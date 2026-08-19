package timewheel

import (
	"Hyper/pkg/strutil"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSimpleTimeWheel(t *testing.T) {
	var triggered int32
	obj := NewSimpleTimeWheel[int](
		1*time.Second,
		10,
		func(
			wheel *SimpleTimeWheel[int],
			key string,
			value int,
		) {
			atomic.AddInt32(&triggered, 1)
		},
	)

	go obj.Start()

	for round := 0; round < 3; round++ {
		for i := 0; i < 1000; i++ {
			m := strutil.NewMsgId()
			obj.Add(m, i, 1*time.Second)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 等待到期任务触发后停止；Stop 之后 Add 应当直接返回而不是阻塞
	time.Sleep(3 * time.Second)
	obj.Stop()
	obj.Add("after-stop", 0, time.Second)

	if atomic.LoadInt32(&triggered) == 0 {
		t.Fatal("no task triggered")
	}
}
