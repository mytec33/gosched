package runner

import (
	"fmt"
	"sync"
	"testing"
)

// Covers the lock-acquisition race found during review by LLM.
func TestSafeMapTryAcquire(t *testing.T) {
	running := newSafeMap()

	existing, acquired := running.tryAcquire("workflow", "run-1")
	if !acquired {
		t.Fatalf("expected first acquire to succeed, existing run ID %q", existing)
	}
	if existing != "" {
		t.Fatalf("expected no existing run ID, got %q", existing)
	}

	existing, acquired = running.tryAcquire("workflow", "run-2")
	if acquired {
		t.Fatal("expected second acquire for same workflow to fail")
	}
	if existing != "run-1" {
		t.Fatalf("expected existing run ID %q, got %q", "run-1", existing)
	}

	running.delete("workflow")

	existing, acquired = running.tryAcquire("workflow", "run-3")
	if !acquired {
		t.Fatalf("expected acquire after delete to succeed, existing run ID %q", existing)
	}
	if existing != "" {
		t.Fatalf("expected no existing run ID after delete, got %q", existing)
	}
}

// Covers the lock-acquisition race found during review by LLM.
func TestSafeMapTryAcquireConcurrent(t *testing.T) {
	running := newSafeMap()

	const attempts = 64
	start := make(chan struct{})
	results := make(chan bool, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, acquired := running.tryAcquire("workflow", fmt.Sprintf("run-%d", i))
			results <- acquired
		}(i)
	}

	close(start)
	wg.Wait()
	close(results)

	acquiredCount := 0
	for acquired := range results {
		if acquired {
			acquiredCount++
		}
	}

	if acquiredCount != 1 {
		t.Fatalf("expected exactly one successful acquire, got %d", acquiredCount)
	}
}

func TestSafeMapConcurrentDifferentKeys(t *testing.T) {
	running := newSafeMap()

	// Workflow A and Workflow B should be able to run at the same time
	_, acquiredA := running.tryAcquire("workflow-A", "run-1")
	_, acquiredB := running.tryAcquire("workflow-B", "run-2")

	if !acquiredA || !acquiredB {
		t.Errorf("expected both different workflows to be acquired, A: %v, B: %v", acquiredA, acquiredB)
	}
}
