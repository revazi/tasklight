package runner

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestActivityWriterForwardsOutputAndTouchesMonitor(t *testing.T) {
	var output bytes.Buffer
	touches := 0
	writer := activityWriter{
		writer: &output,
		touch:  func() { touches++ },
	}

	n, err := writer.Write([]byte("output"))
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}
	if n != len("output") || output.String() != "output" {
		t.Fatalf("Write() = %d, output %q; want forwarded output", n, output.String())
	}
	if touches != 1 {
		t.Fatalf("touches = %d, want 1", touches)
	}
}

func TestIdleMonitorNotifiesOnceAndRearmsAfterActivity(t *testing.T) {
	timer := newFakeIdleTimer()
	notifications := make(chan struct{}, 2)
	monitor := newIdleMonitor(5*time.Minute, func() {
		notifications <- struct{}{}
	})
	monitor.start(func(timeout time.Duration) idleTimer {
		if timeout != 5*time.Minute {
			t.Fatalf("timer timeout = %s, want 5m", timeout)
		}
		return timer
	})
	t.Cleanup(monitor.close)

	timer.fire()
	waitForSignal(t, notifications, "first idle notification")
	timer.fire()

	select {
	case <-notifications:
		t.Fatal("received repeated notification without resumed output")
	case <-time.After(20 * time.Millisecond):
	}

	monitor.touch()
	select {
	case timeout := <-timer.resets:
		if timeout != 5*time.Minute {
			t.Fatalf("reset timeout = %s, want 5m", timeout)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for idle timer reset")
	}

	timer.fire()
	waitForSignal(t, notifications, "rearmed idle notification")
}

func waitForSignal(t *testing.T, ch <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

type fakeIdleTimer struct {
	ch     chan time.Time
	resets chan time.Duration
	mu     sync.Mutex
	active bool
}

func newFakeIdleTimer() *fakeIdleTimer {
	return &fakeIdleTimer{
		ch:     make(chan time.Time, 1),
		resets: make(chan time.Duration, 1),
		active: true,
	}
}

func (timer *fakeIdleTimer) C() <-chan time.Time {
	return timer.ch
}

func (timer *fakeIdleTimer) Stop() bool {
	timer.mu.Lock()
	defer timer.mu.Unlock()
	wasActive := timer.active
	timer.active = false
	return wasActive
}

func (timer *fakeIdleTimer) Reset(timeout time.Duration) bool {
	timer.mu.Lock()
	wasActive := timer.active
	timer.active = true
	timer.mu.Unlock()
	timer.resets <- timeout
	return wasActive
}

func (timer *fakeIdleTimer) fire() {
	timer.mu.Lock()
	if !timer.active {
		timer.mu.Unlock()
		return
	}
	timer.active = false
	timer.mu.Unlock()
	timer.ch <- time.Now()
}
