package timewheel

import (
	"Hyper/pkg/strutil"
	"fmt"
	"testing"
	"time"
)

func TestNewSimpleTimeWheel(t *testing.T) {
	obj := NewSimpleTimeWheel[int](
		1*time.Second,
		10,
		func(
			wheel *SimpleTimeWheel[int],
			key string,
			value int,
		) {
			fmt.Println("trigger:", key, value)
		},
	)

	go obj.Start()

	for round := 0; round < 30; round++ {
		for i := 0; i < 100000; i++ {
			m := strutil.NewMsgId()
			obj.Add(m, i, 1*time.Second)
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Hour)
}
