package statehandler

import "time"

//Sends a timeout signal on a timeout channel when the duration has timeout
func runTimeoutTracker(orderFloor int, timeout time.Duration, completedOrderFloorCh <-chan int, timeoutCh chan<- bool) {
	var initialTime time.Time = time.Now()

	for {
		elapsedTime := time.Now().Sub(initialTime)

		if elapsedTime > timeout {
			timeoutCh <- true
			return
		}
		select {
		case completedOrderFloor := <-completedOrderFloorCh:
			if orderFloor == completedOrderFloor {
				timeoutCh <- false
				return
			}
		}

	}
}
