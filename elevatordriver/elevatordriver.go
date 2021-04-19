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
)

type OrderMatrixBool [cf.ButtonCount][cf.FloorCount]bool

func RunStateMachine(elevatorID int,
	// To statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	completedOrdersCh chan<- int,
	// From statehandler
	acceptedOrderCh <-chan dt.OrderType,

	// To main
	connectNetworkCh chan<- bool,

	// From iomodule
	floorSwitchCh <-chan int,
	stopBtnCh <-chan bool,
	obstructionSwitchCh <-chan bool,
	// To iomodule
	floorIndicatorCh chan<- int,
	motorDirectionCh chan<- dt.MoveDirectionType,
	doorOpenCh chan<- bool,
	setStopCh chan<- bool,
) {

	var elevator dt.ElevatorState = dt.ElevatorState{
		ElevatorID:      elevatorID,
		MovingDirection: dt.MovingStopped,
		Floor:           0,
		State:           dt.InitState,
		IsFunctioning:   true,
	}

	// Internal variables
	var orderMatrix OrderMatrixBool
	var doorObstructed bool
	var timeStuckLimit time.Duration = time.Duration(cf.TimeoutStuckSec) * time.Second //seconds
	var timeDoorOpen time.Duration = time.Duration(cf.DoorOpenTime) * time.Second      //seconds

	// Previous values register
	var oldState dt.DriverStateType = elevator.State
	var oldDirection dt.MoveDirectionType = elevator.MovingDirection
	var oldFloor int = elevator.Floor

	// Internal channels

	restartDoorTimerCh := make(chan bool)
	doorClosingTimerCh := make(chan bool)

	restartFailTimerCh := make(chan bool)
	stopFailTimerCh := make(chan bool)

	timeOutDetectedCh := make(chan bool)

	// Time-out-module in case of elevator not working
	go runTimeOut(timeStuckLimit, restartFailTimerCh, stopFailTimerCh, timeOutDetectedCh)

	// Time-out module for closing the door
	go runTimeOut(timeDoorOpen, restartDoorTimerCh, make(<-chan bool), doorClosingTimerCh)

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

	elevator.State = dt.IdleState
	driverStateUpdateCh <- elevator

	// Run State machine
	for {
		newOrderMatrix := orderMatrix
		isFunctioning := elevator.IsFunctioning

		newState := elevator.State
		newDirection := elevator.MovingDirection
		newFloor := elevator.Floor

		select {

		// AcceptedOrder order to be executed by the elevator
		case newAcceptedOrder := <-acceptedOrderCh:

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			newOrderMatrix = SetOrder(orderMatrix, newAcceptedOrder, ACTIVE)

			if elevator.State == dt.IdleState || elevator.State == dt.DoorOpenState {
				if elevator.Floor == newAcceptedOrder.Floor {

					newState = dt.DoorOpenState

				} else if elevator.State != dt.DoorOpenState {

					newDirection = ChooseDirection(elevator.MovingDirection, elevator.Floor, newOrderMatrix)
					newState = dt.MovingState

				}
			}

		case floor := <-floorSwitchCh:

			if ElevatorShouldStop(elevator.MovingDirection, floor, orderMatrix) {
				newState = dt.DoorOpenState
			}

			newFloor = floor

		case <-doorClosingTimerCh:
			if doorObstructed {
				restartDoorTimerCh <- true
			} else {

				newDirection = ChooseDirection(elevator.MovingDirection, elevator.Floor, orderMatrix)

				if newDirection == dt.MovingStopped {
					newState = dt.IdleState
				} else {
					newState = dt.MovingState
				}
			}

		case obstructedSwitch := <-obstructionSwitchCh:
			doorObstructed = obstructedSwitch

		case <-timeOutDetectedCh:
			newState = dt.ErrorState

		case <-stopBtnCh:
		}

		if newState != oldState {
			fmt.Printf("STATE: %v  \n", string(newState))

			switch oldState {
			case dt.ErrorState:
				connectNetworkCh <- true
				isFunctioning = true
			case dt.DoorOpenState:
				completedOrdersCh <- elevator.Floor

				//TODO: should order be closed when opening or closing the door?
				newOrderMatrix = ClearOrdersAtCurrentFloor(elevator.Floor, orderMatrix)

				doorOpenCh <- CLOSE_DOOR
			}

			switch newState {
			case dt.IdleState:
				stopFailTimerCh <- true

			case dt.DoorOpenState:
				newDirection = dt.MovingStopped

				restartDoorTimerCh <- true

				doorOpenCh <- OPEN_DOOR
				newOrderMatrix = ClearOrdersAtCurrentFloor(elevator.Floor, newOrderMatrix)
				completedOrdersCh <- elevator.Floor

				restartFailTimerCh <- true

			case dt.MovingState:
				restartFailTimerCh <- true

			case dt.ErrorState:
				isFunctioning = false
				connectNetworkCh <- false
			}

		}

		if newDirection != oldDirection {
			motorDirectionCh <- newDirection
		}

		if newFloor != oldFloor {
			floorIndicatorCh <- newFloor
			restartFailTimerCh <- true

			if oldState == dt.ErrorState {
				connectNetworkCh <- true
				isFunctioning = true
			}
		}

		elevator.State = newState
		elevator.MovingDirection = newDirection
		elevator.Floor = newFloor
		elevator.IsFunctioning = isFunctioning

		// Send state update to statehandler
		driverStateUpdateCh <- elevator

		orderMatrix = newOrderMatrix

		oldState = newState
		oldDirection = newDirection
		oldFloor = newFloor

	}
}
