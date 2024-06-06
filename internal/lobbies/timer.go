package lobbies

import "time"

type cancellableTimer struct {
	timer      *time.Timer
	cancelChan chan struct{}
}

func NewCancellableTimer(d time.Duration) *cancellableTimer {
	return &cancellableTimer{
		timer:      time.NewTimer(d),
		cancelChan: make(chan struct{}),
	}
}

func (ct *cancellableTimer) Cancel() {
	// Close the cancelChan to signal cancellation
	close(ct.cancelChan)

	// Stop the timer if it is still running
	if !ct.timer.Stop() {
		// If the timer had already expired and fired,
		// we need to drain the timer channel to prevent a potential deadlock
		<-ct.timer.C
	}
}
