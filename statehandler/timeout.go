package statehandler

import "time"

//Sends a timeout signal on a timeout channel when the duration has timeout
func runTimeoutTracker(timeout time.Duration, startCh <-chan bool, stopCh <-chan bool, timeoutCh chan<- bool) {
	var initialTime time.Time
	var runTimer bool = false

	for {
		elapsedTime := time.Now().Sub(initialTime)

		if runTimer && elapsedTime > timeout {
			timeoutCh <- true
		}
		select {
		case <-stopCh:
			runTimer = false
		case <-startCh:
			initialTime = time.Now()
			runTimer = true
		}

	}
}
