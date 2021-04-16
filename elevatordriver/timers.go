package elevatordriver

import "time"

func runTimeOut(timeLimit time.Duration, startMotorFailTimerCh <-chan bool, stopTimerCh <-chan bool, timeOutDetectedCh chan<- bool) {
	var initialTime time.Time
	var timerOn bool = false

	for {
		select {
		case <-startMotorFailTimerCh:
			timerOn = true

			initialTime = time.Now()

		case <-stopTimerCh:
			timerOn = false

		default:
			if timerOn {
				elapsedTime := time.Now().Sub(initialTime)

				if elapsedTime > timeLimit {
					timeOutDetectedCh <- true
					timerOn = false
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func startDoorTimer(doorTimerCh chan<- bool) {
	time.Sleep(3000 * time.Millisecond)
	doorTimerCh <- true
}
