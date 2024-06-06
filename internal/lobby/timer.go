package lobby

import "time"

type CancellableTimer struct {
	timer      *time.Timer
	cancelChan chan struct{}
}

func NewCancellableTimer(d time.Duration) *CancellableTimer {
	return &CancellableTimer{
		timer:      time.NewTimer(d),
		cancelChan: make(chan struct{}),
	}
}

func (ct *CancellableTimer) Cancel() {
	// Close the cancelChan to signal cancellation
	close(ct.cancelChan)

	// Stop the timer if it is still running
	if !ct.timer.Stop() {
		// If the timer had already expired and fired,
		// we need to drain the timer channel to prevent a potential deadlock
		<-ct.timer.C
	}
}
