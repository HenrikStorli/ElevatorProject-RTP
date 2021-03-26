package elevatordriver

import "time"


func startDoorTimer(doorTimerCh chan<- bool){
		time.Sleep(3000 * time.Millisecond)
		doorTimerCh <- true
}