package elevatordriver_test

import (
	"fmt"
	"testing"
	"time"

	dt "../datatypes"
	"../elevatordriver"
)

func TestElevatorDriverModule(*testing.T) {
	//To statehandler
	driverStateUpdateCh := make(chan dt.ElevatorState)
	completedOrdersCh := make(chan int)
	//From statehandler
	acceptedOrderCh := make(chan dt.OrderType)
	restartCh := make(chan bool)
	//From elevio
	floorSwitchCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)
	//To elevio
	floorIndicatorCh := make(chan int)
	motorDirectionCh := make(chan dt.MoveDirectionType)
	doorOpenCh := make(chan bool)
	setStopCh := make(chan bool)

	elevatorID := 0

	go elevatordriver.RunStateMachine(elevatorID, driverStateUpdateCh, completedOrdersCh,
		acceptedOrderCh, restartCh, floorSwitchCh, stopBtnCh, obstructionSwitchCh,
		floorIndicatorCh, motorDirectionCh, doorOpenCh, setStopCh)

	go func() {
		time.Sleep(10 * time.Millisecond)
		acceptedOrderCh <- dt.OrderType{Button: dt.BtnHallUp, Floor: 1}
		time.Sleep(10 * time.Millisecond)
	}()

	for {
		select {
		case driverStateUpdate := <-driverStateUpdateCh:
			fmt.Println(driverStateUpdate)
		case completedOrders := <-completedOrdersCh:
			fmt.Println(completedOrders)
		case motorDirection := <-motorDirectionCh:
			fmt.Println(motorDirection)
		}
	}

}
