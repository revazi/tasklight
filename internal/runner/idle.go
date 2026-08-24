package runner

import (
	"io"
	"time"
)

type idleTimer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(time.Duration) bool
}

type systemIdleTimer struct {
	*time.Timer
}

func (timer systemIdleTimer) C() <-chan time.Time {
	return timer.Timer.C
}

type idleTimerFactory func(time.Duration) idleTimer

type idleMonitor struct {
	timeout  time.Duration
	onIdle   func()
	activity chan struct{}
	stop     chan struct{}
	done     chan struct{}
}

func newIdleMonitor(timeout time.Duration, onIdle func()) *idleMonitor {
	return &idleMonitor{
		timeout:  timeout,
		onIdle:   onIdle,
		activity: make(chan struct{}, 1),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (monitor *idleMonitor) start(factory idleTimerFactory) {
	go monitor.run(factory(monitor.timeout))
}

func (monitor *idleMonitor) run(timer idleTimer) {
	defer close(monitor.done)
	armed := true

	for {
		select {
		case <-monitor.activity:
			resetIdleTimer(timer, monitor.timeout)
			armed = true
		case <-timer.C():
			if armed {
				armed = false
				monitor.onIdle()
			}
		case <-monitor.stop:
			timer.Stop()
			return
		}
	}
}

func (monitor *idleMonitor) touch() {
	select {
	case monitor.activity <- struct{}{}:
	default:
	}
}

func (monitor *idleMonitor) close() {
	close(monitor.stop)
	<-monitor.done
}

func resetIdleTimer(timer idleTimer, timeout time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C():
		default:
		}
	}
	timer.Reset(timeout)
}

type activityWriter struct {
	writer io.Writer
	touch  func()
}

func (writer activityWriter) Write(p []byte) (int, error) {
	n, err := writer.writer.Write(p)
	if n > 0 {
		writer.touch()
	}
	return n, err
}

func newSystemIdleTimer(timeout time.Duration) idleTimer {
	return systemIdleTimer{Timer: time.NewTimer(timeout)}
}
