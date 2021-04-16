package elevatordriver

import (
	"fmt"
	"time"

	cf "../config"
	dt "../datatypes"
)

const (
	OPEN_DOOR  = true
	CLOSE_DOOR = false

	TIMER_ON  = true
	TIMER_OFF = false
)

type OrderMatrixBool [cf.ButtonCount][cf.FloorCount]bool

func RunStateMachine(elevatorID int,
	// To statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	completedOrdersCh chan<- int,
	// From statehandler
	acceptedOrderCh <-chan dt.OrderType,

	// To main
	restartCh chan<- bool,

	// From elevio
	floorSwitchCh <-chan int,
	stopBtnCh <-chan bool,
	obstructionSwitchCh <-chan bool,
	// To elevio
	floorIndicatorCh chan<- int,
	motorDirectionCh chan<- dt.MoveDirectionType,
	doorOpenCh chan<- bool,
	setStopCh chan<- bool,
) {

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
	var timeLimit time.Duration = time.Duration(cf.TimeoutStuckSec) * time.Second //seconds

	// Internal channels
	doorTimerCh := make(chan bool)
	startMotorFailTimerCh := make(chan bool)
	stopTimerCh := make(chan bool)
	timeOutDetectedCh := make(chan bool)

	// Time-out-module in case of motor not working
	go runTimeOut(timeLimit, startMotorFailTimerCh, stopTimerCh, timeOutDetectedCh)

	// Close door at start
	doorOpenCh <- CLOSE_DOOR

	// Initialize the elevator position
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

			newOrderMatrix := SetOrder(orderMatrix, newAcceptedOrder, ACTIVE)
			newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, elevator, newOrderMatrix)

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			if elevator.State != newElevator.State {
				switch newElevator.State {
				case dt.Moving:
					motorDirectionCh <- newElevator.MovingDirection
					startMotorFailTimerCh <- TIMER_ON
				case dt.DoorOpen:
					go startDoorTimer(doorTimerCh)
					doorOpenCh <- OPEN_DOOR
					completedOrdersCh <- newElevator.Floor
				}

			} else {
				if elevator.State == dt.DoorOpen {
					newOrderMatrix = ClearOrdersAtCurrentFloor(newElevator, newOrderMatrix)
					completedOrdersCh <- newElevator.Floor
				}
			}

			elevator = newElevator
			orderMatrix = newOrderMatrix

		case newFloor := <-floorSwitchCh:

			newOrderMatrix, newElevator := updateOnFloorArrival(newFloor, elevator, orderMatrix)

			floorIndicatorCh <- newFloor

			stopTimerCh <- TIMER_OFF

			if newElevator.State == dt.DoorOpen {
				motorDirectionCh <- dt.MovingStopped

				doorOpenCh <- OPEN_DOOR

				go startDoorTimer(doorTimerCh)

				completedOrdersCh <- newFloor

			} else {
				startMotorFailTimerCh <- TIMER_ON
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

				if newElevator.MovingDirection != dt.MovingStopped {
					startMotorFailTimerCh <- TIMER_ON
				}

				elevator = newElevator
			}

		case obstructedSwitch := <-obstructionSwitchCh:
			doorObstructed = obstructedSwitch

		case <-timeOutDetectedCh:
			restartCh <- true

		case <-stopBtnCh:
		}

		driverStateUpdateCh <- elevator

		if elevator.State != oldState {
			fmt.Printf("STATE: %v  \n", string(elevator.State))
		}
		oldState = elevator.State
	}
}
