package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

type latencySample struct {
	latency time.Duration
	failed  bool
}

// fireConcurrent creates concurrency unique signing sessions at the same time
// and returns the latency profile, throughput, failure count and heap growth.
// It is intentionally self-contained so it can be replayed with a single test
// command and does not depend on shared test state.
func fireConcurrent(h http.Handler, concurrency int) (samples []latencySample, heapDelta uint64, elapsed time.Duration) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	start := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			body := fmt.Sprintf(`{"wallet_id":"wallet-demo-001","digest":"concurrent-%d"}`, index)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/sign-sessions", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			reqStart := time.Now()
			h.ServeHTTP(recorder, req)
			latency := time.Since(reqStart)
			mu.Lock()
			samples = append(samples, latencySample{latency: latency, failed: recorder.Code != http.StatusOK})
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	elapsed = time.Since(start)
	runtime.ReadMemStats(&after)
	heapDelta = after.HeapAlloc - before.HeapAlloc
	return samples, heapDelta, elapsed
}

func percentile(samples []latencySample, p float64) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := make([]time.Duration, 0, len(samples))
	for _, sample := range samples {
		sorted = append(sorted, sample.latency)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(float64(len(sorted)) * p)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func countFailures(samples []latencySample) int {
	failures := 0
	for _, sample := range samples {
		if sample.failed {
			failures++
		}
	}
	return failures
}

func TestConcurrentSessionCreation(t *testing.T) {
	h := NewRouter("../../../../../web/admin-console")
	samples, _, _ := fireConcurrent(h, 10)
	if failures := countFailures(samples); failures != 0 {
		t.Fatalf("10-concurrent session creation had %d failures", failures)
	}
}

func TestConcurrencyProfile(t *testing.T) {
	h := NewRouter("../../../../../web/admin-console")
	for _, concurrency := range []int{10, 25, 50, 100} {
		concurrency := concurrency
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			samples, heapDelta, elapsed := fireConcurrent(h, concurrency)
			failures := countFailures(samples)
			throughput := float64(len(samples)) / elapsed.Seconds()
			p50 := percentile(samples, 0.50)
			p95 := percentile(samples, 0.95)
			p99 := percentile(samples, 0.99)
			t.Logf("PROFILE concurrency=%d p50=%s p95=%s p99=%s throughput=%.1f req/s failures=%d heap_delta=%dB",
				concurrency, p50, p95, p99, throughput, failures, heapDelta)
			if failures != 0 {
				t.Fatalf("concurrency=%d had %d failures", concurrency, failures)
			}
		})
	}
}
