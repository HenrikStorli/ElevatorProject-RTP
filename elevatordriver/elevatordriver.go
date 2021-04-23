package elevatordriver

import (
	"fmt"
	"time"

	cf "../config"
	dt "../datatypes"
)

const (
	openDoor  = true
	closeDoor = false
)

type OrderMatrixBool [cf.ButtonCount][cf.FloorCount]bool

func RunElevatorDriverModule(elevatorID int,
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
		MovingDirection: dt.MovingNeutral,
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
	doorOpenCh <- closeDoor

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
		motorDirectionCh <- dt.MovingNeutral
	}

	elevator.State = dt.IdleState
	driverStateUpdateCh <- elevator

	// Run State machine
	for {
		newOrderMatrix := orderMatrix
		isFunctioning := elevator.IsFunctioning

		newState := dt.InvalidState
		newDirection := dt.MovingInvalid
		newFloor := elevator.Floor

		select {

		// AcceptedOrder order to be executed by the elevator
		case newAcceptedOrder := <-acceptedOrderCh:

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			newOrderMatrix = SetOrder(orderMatrix, newAcceptedOrder, ActiveOrder)

			if elevator.Floor == newAcceptedOrder.Floor {
				if elevator.State == dt.DoorOpenState || elevator.State == dt.IdleState {
					newState = dt.DoorOpenState
				}
			} else if elevator.State == dt.IdleState {
				newDirection = ChooseDirection(elevator.MovingDirection, elevator.Floor, newOrderMatrix)
				newState = dt.MovingState
			}

		case floor := <-floorSwitchCh:

			if ElevatorShouldStop(elevator.MovingDirection, floor, orderMatrix) {
				newState = dt.DoorOpenState
			} else if oldState == dt.ErrorState {
				newState = dt.MovingState
			}

			floorIndicatorCh <- floor
			restartFailTimerCh <- true
			newFloor = floor

		case <-doorClosingTimerCh:
			if doorObstructed {
				restartDoorTimerCh <- true
			} else {

				newDirection = ChooseDirection(elevator.MovingDirection, elevator.Floor, orderMatrix)

				if newDirection == dt.MovingNeutral {
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
			newDirection = dt.MovingNeutral
		}

		if newState != dt.InvalidState {
			fmt.Printf("STATE: %v  \n", string(newState))

			switch oldState {
			case dt.ErrorState:
				connectNetworkCh <- true
				isFunctioning = true

				if newState == dt.IdleState || newState == dt.MovingState {
					doorOpenCh <- closeDoor
				}

			case dt.DoorOpenState:
				if newState != dt.ErrorState {
					doorOpenCh <- closeDoor
				}
			}

			switch newState {
			case dt.IdleState:
				stopFailTimerCh <- true
				newDirection = dt.MovingNeutral

			case dt.DoorOpenState:
				motorDirectionCh <- dt.MovingNeutral

				restartDoorTimerCh <- true

				doorOpenCh <- openDoor
				newOrderMatrix = ClearOrdersAtCurrentFloor(newFloor, newOrderMatrix)
				completedOrdersCh <- newFloor

				restartFailTimerCh <- true

			case dt.MovingState:
				restartFailTimerCh <- true

			case dt.ErrorState:
				isFunctioning = false
				connectNetworkCh <- false

				newOrderMatrix = clearAllHallOrders(newOrderMatrix)
			}

			elevator.State = newState
			oldState = newState
		}

		if newDirection != dt.MovingInvalid {
			motorDirectionCh <- newDirection

			elevator.MovingDirection = newDirection
		}

		elevator.Floor = newFloor
		elevator.IsFunctioning = isFunctioning

		// Send state update to statehandler
		driverStateUpdateCh <- elevator

		orderMatrix = newOrderMatrix

	}
}
