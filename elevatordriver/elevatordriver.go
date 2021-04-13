package elevatordriver

import (
	"fmt"
	"time"
	dt "../datatypes"
)

const (
	OPEN_DOOR  = true
	CLOSE_DOOR = false

	TIMER_ON = true
	TIMER_OFF = false
)

type OrderMatrixBool [dt.ButtonCount][dt.FloorCount]bool

func RunStateMachine(elevatorID int,
	//To statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	completedOrdersCh chan<- int,
	//From statehandler
	acceptedOrderCh <-chan dt.OrderType,
	restartCh chan<- bool,
	//From elevio
	floorSwitchCh <-chan int,
	stopBtnCh <-chan bool,
	obstructionSwitchCh <-chan bool,
	//To elevio
	floorIndicatorCh chan<- int,
	motorDirectionCh chan<- dt.MoveDirectionType,
	doorOpenCh chan<- bool,
	setStopCh chan<- bool) {
	// Local data

	var elevator dt.ElevatorState = dt.ElevatorState{
		ElevatorID:      elevatorID,
		MovingDirection: dt.MovingStopped,
		Floor:           0,
		State:           dt.Idle,
		IsFunctioning:   true,
	}
	var oldState dt.MachineStateType
	var orderMatrix OrderMatrixBool
	var doorObstructed bool
	var timeLimit time.Duration = time.Duration(7)*time.Second //seconds


	//Internal channels
	doorTimerCh  		:= make(chan bool)
	startTimerCh 		:= make(chan bool)
	stopTimerCh 		:= make(chan bool)
	timeOutDetectedCh 	:= make(chan bool)

	// Time-out-module in case of motor not working
	go runTimeOut(timeLimit, startTimerCh, stopTimerCh, timeOutDetectedCh)

	// Close door at start
	doorOpenCh <- CLOSE_DOOR

	// Initialize the elevators position
	select {
	case newFloor := <-floorSwitchCh:
		floorIndicatorCh <- newFloor
		elevator.Floor = newFloor
	default:
		motorDirectionCh <- dt.MovingDown
		newFloor := <-floorSwitchCh
		floorIndicatorCh <- newFloor
		elevator.Floor = newFloor
		motorDirectionCh <- dt.MovingStopped
	}

	driverStateUpdateCh <- elevator



	// Run State machine
	for {
		select {
		case newAcceptedOrder := <-acceptedOrderCh:
			newOrderMatrix, newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, elevator, orderMatrix)

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			if elevator.State != newElevator.State {
				if newElevator.State == dt.Moving {
					motorDirectionCh <- newElevator.MovingDirection
					// Start timeout-timer
					startTimerCh <- TIMER_ON

				} else if newElevator.State == dt.DoorOpen {
					go startDoorTimer(doorTimerCh)
					doorOpenCh <- OPEN_DOOR
					completedOrdersCh <- newElevator.Floor
				}

			} else {
				if newElevator.State == dt.DoorOpen {
					completedOrdersCh <- newElevator.Floor
				}
			}

			elevator = newElevator
			orderMatrix = newOrderMatrix

		case newFloor := <-floorSwitchCh:
			newOrderMatrix, newElevator := updateOnNewFloorArrival(newFloor, elevator, orderMatrix)

			floorIndicatorCh <- newFloor
			// Stop timout-timer
			stopTimerCh <- TIMER_OFF

			if newElevator.State == dt.DoorOpen {
				motorDirectionCh <- dt.MovingStopped
				doorOpenCh <- OPEN_DOOR
				go startDoorTimer(doorTimerCh)
				completedOrdersCh <- newFloor
			} else {
				// Start timeout-timer
				startTimerCh <- TIMER_ON
			}

			elevator = newElevator
			orderMatrix = newOrderMatrix

		case <-doorTimerCh:
			if doorObstructed {
				go startDoorTimer(doorTimerCh)

			} else {
				doorOpenCh <- CLOSE_DOOR
				newElevator := updateOnDoorClosing(elevator, orderMatrix)
				motorDirectionCh <- newElevator.MovingDirection

				if newElevator.MovingDirection != dt.MovingStopped{
					startTimerCh <- TIMER_ON
					fmt.Println("Timer is turned on after door closes")
				}
				elevator = newElevator
			}


		case doorObstructed = <-obstructionSwitchCh:

		case <-timeOutDetectedCh:
			restartCh <- true

		case <-stopBtnCh:
		}
		// Send updated elevator to statehandler

		driverStateUpdateCh <- elevator // This type does not match the type of the channel

		if elevator.State != oldState {
			fmt.Printf("STATE: %v  \n", string(elevator.State))
		}

		oldState = elevator.State
	}
}
