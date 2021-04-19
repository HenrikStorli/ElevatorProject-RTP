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

func RunElevatorDriverModule(elevatorID int,
	// To statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	completedOrdersCh chan<- int,
	// From statehandler
	acceptedOrderCh <-chan dt.OrderType,

	// To main
	connectNetworkCh chan<- bool,

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
		State:           dt.Init,
		IsFunctioning:   true,
	}

	var oldState dt.MachineStateType
	var orderMatrix OrderMatrixBool
	var doorObstructed bool
	var timeStuckLimit time.Duration = time.Duration(cf.TimeoutStuckSec) * time.Second //seconds
	var timeDoorOpen time.Duration = time.Duration(cf.DoorOpenTime) * time.Second      //seconds


	// Internal channels
	doorClosingTimerCh 	:= make(chan bool)
	startDoorTimerCh 		:= make(chan bool)

	startFailTimerCh 		:= make(chan bool)
	stopFailTimerCh 		:= make(chan bool)

	timeOutDetectedCh 	:= make(chan bool)

	// Time-out-module in case of motor not working
	go runTimeOut(timeStuckLimit, startFailTimerCh, stopFailTimerCh, timeOutDetectedCh)

	go runTimeOut(timeDoorOpen, startDoorTimerCh, make(<-chan bool), doorClosingTimerCh)

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

	elevator.State = dt.Idle
	driverStateUpdateCh <- elevator

	// Run State machine
	for {
		newOrderMatrix := orderMatrix
		updatedElevator := elevator

		select {

		// Accepted order to be executed by the elevator
		case newAcceptedOrder := <-acceptedOrderCh:

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			newOrderMatrix = SetOrder(orderMatrix, newAcceptedOrder, ACTIVE)

			if elevator.State == dt.Idle || elevator.State == dt.DoorOpen {
				if updatedElevator.Floor == newAcceptedOrder.Floor {
					updatedElevator.State = dt.DoorOpen
					startDoorTimerCh <- true
					doorOpenCh <- OPEN_DOOR
					newOrderMatrix = ClearOrdersAtCurrentFloor(updatedElevator, newOrderMatrix)
					completedOrdersCh <- updatedElevator.Floor

				} else if elevator.State == dt.Idle {
					updatedElevator.State = dt.Moving
					updatedElevator.MovingDirection = ChooseDirection(updatedElevator, newOrderMatrix)
					motorDirectionCh <- updatedElevator.MovingDirection
				}
				startFailTimerCh <- true
			}

		case newFloor := <-floorSwitchCh:
			floorIndicatorCh <- newFloor
			startFailTimerCh <- true
			if elevator.State == dt.Error {
					connectNetworkCh <- true
					updatedElevator.IsFunctioning = true
			}
			updatedElevator.Floor = newFloor
			if ElevatorShouldStop(updatedElevator, orderMatrix) {
				motorDirectionCh <- dt.MovingStopped
				completedOrdersCh <- newFloor
				startDoorTimerCh <- true
				doorOpenCh <- OPEN_DOOR
				updatedElevator.State = dt.DoorOpen
				newOrderMatrix = ClearOrdersAtCurrentFloor(updatedElevator, orderMatrix)
			}

		case <- doorClosingTimerCh:
			if doorObstructed {
				startDoorTimerCh <- true
			} else {
				completedOrdersCh <- updatedElevator.Floor
				newOrderMatrix = ClearOrdersAtCurrentFloor(elevator, orderMatrix)
				doorOpenCh <- CLOSE_DOOR
				if elevator.State == dt.Error {
						connectNetworkCh <- true
						updatedElevator.IsFunctioning = true
				}
				updatedElevator.MovingDirection = ChooseDirection(elevator, orderMatrix)
				motorDirectionCh <- updatedElevator.MovingDirection
				if updatedElevator.MovingDirection == dt.MovingStopped {
					updatedElevator.State = dt.Idle
					stopFailTimerCh <- true
				} else {
					updatedElevator.State = dt.Moving
					startFailTimerCh <- true
				}
			}


		case obstructedSwitch := <-obstructionSwitchCh:
			doorObstructed = obstructedSwitch

		case <-timeOutDetectedCh:
			updatedElevator.State = dt.Error
			updatedElevator.IsFunctioning = false
			connectNetworkCh <- false

		case <-stopBtnCh:
		}

		elevator = updatedElevator
		orderMatrix = newOrderMatrix

		driverStateUpdateCh <- elevator

		if elevator.State != oldState {
			fmt.Printf("STATE: %v  \n", string(elevator.State))
		}
		oldState = elevator.State
	}
}
