package main

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

func request(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestTwentyFunctionalChecks(t *testing.T) {
	a, h := newAPI(), newAPI().routes()
	_ = a
	// Use one API instance for all checks so approval state can be verified.
	a = newAPI()
	h = a.routes()
	cases := []struct {
		name, method, path, body string
		want                     int
	}{
		{"health", "GET", "/healthz", "", 200}, {"wallet list", "GET", "/api/v1/wallets", "", 200}, {"wallet detail", "GET", "/api/v1/wallets/wallet-demo-001", "", 200},
		{"nodes list", "GET", "/api/v1/nodes", "", 200}, {"node detail", "GET", "/api/v1/nodes/node-1", "", 200}, {"transaction list", "GET", "/api/v1/transactions", "", 200},
		{"transaction detail", "GET", "/api/v1/transactions/tx-1", "", 200}, {"audit logs", "GET", "/api/v1/audit-logs", "", 200}, {"metrics", "GET", "/api/v1/metrics", "", 200},
		{"key epochs", "GET", "/api/v1/key-epochs", "", 200}, {"key epoch detail", "GET", "/api/v1/key-epochs/1", "", 200}, {"missing session", "GET", "/api/v1/sessions/missing", "", 404},
		{"reject empty digest", "POST", "/api/v1/sessions", `{}`, 400}, {"reject wallet post", "POST", "/api/v1/wallets", `{}`, 405}, {"reject node post", "POST", "/api/v1/nodes", `{}`, 405},
	}
	for _, c := range cases {
		if got := request(h, c.method, c.path, c.body).Code; got != c.want {
			t.Errorf("%s: got %d want %d", c.name, got, c.want)
		}
	}
	created := request(h, "POST", "/api/v1/sessions", `{"digest":"abc123"}`)
	if created.Code != 201 {
		t.Fatalf("create session: %d", created.Code)
	}
	var id string
	a.mu.RLock()
	for k := range a.sessions {
		id = k
	}
	a.mu.RUnlock()
	for i := 0; i < 2; i++ {
		if got := request(h, "POST", "/api/v1/sessions/"+id+"?action=approve", "").Code; got != 200 {
			t.Fatalf("approval %d: %d", i, got)
		}
	}
	if got := request(h, "POST", "/api/v1/sessions/"+id+"?action=approve", "").Code; got != 409 {
		t.Errorf("third approval: got %d want 409", got)
	}
	if got := request(h, "GET", "/api/v1/sessions/"+id, "").Code; got != 200 {
		t.Errorf("session status: %d", got)
	}
}

func TestConcurrentSessionCreation(t *testing.T) {
	a, h := newAPI(), newAPI().routes()
	_ = a
	a = newAPI()
	h = a.routes()
	const count = 100
	var wg sync.WaitGroup
	errs := make(chan int, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if got := request(h, "POST", "/api/v1/sessions", `{"digest":"stress"}`).Code; got != 201 {
				errs <- got
			}
		}()
	}
	wg.Wait()
	close(errs)
	for code := range errs {
		t.Errorf("unexpected status %d", code)
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.sessions) != count {
		t.Fatalf("got %d sessions want %d", len(a.sessions), count)
	}
}

func TestConcurrencyProfile(t *testing.T) {
	for _, count := range []int{10, 25, 50, 100} {
		a, h := newAPI(), newAPI().routes()
		_ = a
		a, h = newAPI(), newAPI().routes()
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		latencies := make([]time.Duration, count)
		statuses := make(chan int, count)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				begin := time.Now()
				statuses <- request(h, "POST", "/api/v1/sessions", `{"digest":"profile"}`).Code
				latencies[i] = time.Since(begin)
			}(i)
		}
		suiteStart := time.Now()
		close(start)
		wg.Wait()
		elapsed := time.Since(suiteStart)
		close(statuses)
		failures := 0
		for status := range statuses {
			if status != 201 {
				failures++
			}
		}
		runtime.ReadMemStats(&after)
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		percentile := func(p float64) time.Duration { index := int(float64(len(latencies)-1) * p); return latencies[index] }
		fmt.Printf("PROFILE concurrency=%d p50=%s p95=%s p99=%s throughput=%.1f req/s failures=%d heap_delta=%dB\n", count, percentile(.50), percentile(.95), percentile(.99), float64(count)/elapsed.Seconds(), failures, int64(after.HeapAlloc)-int64(before.HeapAlloc))
		if failures != 0 {
			t.Fatalf("concurrency %d produced %d failures", count, failures)
		}
	}
}
