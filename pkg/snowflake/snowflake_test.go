package snowflake

import (
	"sync"
	"testing"
)

// 1️⃣ 基础测试：能不能生成 ID
func TestGenUserID(t *testing.T) {
	id := GenUserID()
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	t.Logf("generated user id: %d", id)
}

// 2️⃣ 唯一性测试：单线程生成
func TestGenUserID_Unique(t *testing.T) {
	const n = 10000
	ids := make(map[int64]struct{}, n)

	for i := 0; i < n; i++ {
		id := GenUserID()
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate id found: %d", id)
		}
		ids[id] = struct{}{}
	}
}

// 3️⃣ 并发测试：多 goroutine 生成（最重要）
func TestGenUserID_Concurrent(t *testing.T) {
	const (
		goroutines = 20
		perRoutine = 5000
		total      = goroutines * perRoutine
	)

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		ids = make(map[int64]struct{}, total)
	)

	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perRoutine; i++ {
				id := GenUserID()

				mu.Lock()
				if _, exists := ids[id]; exists {
					t.Fatalf("duplicate id found in concurrent test: %d", id)
				}
				ids[id] = struct{}{}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
}

// 4️⃣ 顺序性测试（非严格递增，但应该整体递增）
func TestGenUserID_Order(t *testing.T) {
	prev := GenUserID()

	for i := 0; i < 1000; i++ {
		curr := GenUserID()
		if curr <= prev {
			t.Fatalf("ids not increasing: prev=%d curr=%d", prev, curr)
		}
		prev = curr
	}
}
