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
	stopMotorFailTimerCh := make(chan bool)
	timeOutDetectedCh := make(chan bool)

	// Time-out-module in case of motor not working
	go runTimeOut(timeLimit, startMotorFailTimerCh, stopMotorFailTimerCh, timeOutDetectedCh)

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

		// Accepted order to be executed by the elevator
		case newAcceptedOrder := <-acceptedOrderCh:

			fmt.Printf("Accepting Order %v\n", newAcceptedOrder)

			newOrderMatrix := SetOrder(orderMatrix, newAcceptedOrder, ACTIVE)
			updatedElevator := elevator

			if elevator.State == dt.Idle {
				if elevator.Floor == newAcceptedOrder.Floor {
					updatedElevator.State = dt.DoorOpen
					go startDoorTimer(doorTimerCh)
					doorOpenCh <- OPEN_DOOR
					completedOrdersCh <- updatedElevator.Floor

				} else {
					updatedElevator.State = dt.Moving
					updatedElevator.MovingDirection = ChooseDirection(updatedElevator, newOrderMatrix)
					motorDirectionCh <- updatedElevator.MovingDirection
					startMotorFailTimerCh <- true
				}
			} else if elevator.State == dt.DoorOpen {
				newOrderMatrix = ClearOrdersAtCurrentFloor(updatedElevator, newOrderMatrix)
				completedOrdersCh <- elevator.Floor
			}

			elevator = updatedElevator
			orderMatrix = newOrderMatrix

		case newFloor := <-floorSwitchCh:

			newOrderMatrix := orderMatrix
			updatedElevator := elevator

			updatedElevator.Floor = newFloor
			floorIndicatorCh <- newFloor

			if elevator.State == dt.Moving {
				if ElevatorShouldStop(updatedElevator, orderMatrix) {
					motorDirectionCh <- dt.MovingStopped

					updatedElevator.State = dt.DoorOpen
					doorOpenCh <- OPEN_DOOR
					go startDoorTimer(doorTimerCh)

					newOrderMatrix = ClearOrdersAtCurrentFloor(updatedElevator, orderMatrix)
					completedOrdersCh <- newFloor

					stopMotorFailTimerCh <- true
				} else {
					startMotorFailTimerCh <- true
				}
			} else {
				fmt.Println("Was not in moving state")
			}

			elevator = updatedElevator
			orderMatrix = newOrderMatrix

		case <-doorTimerCh:

			if doorObstructed {
				go startDoorTimer(doorTimerCh)
			} else {
				doorOpenCh <- CLOSE_DOOR

				updatedElevator := updateOnDoorClosing(elevator, orderMatrix)

				motorDirectionCh <- updatedElevator.MovingDirection

				if updatedElevator.MovingDirection != dt.MovingStopped {
					startMotorFailTimerCh <- true
				}

				elevator = updatedElevator
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
