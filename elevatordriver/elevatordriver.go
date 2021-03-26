package elevatordriver

import (
	dt "../datatypes"
	"fmt"
)

const (
	OPEN_DOOR 	= true
	CLOSE_DOOR 	= false
)

type OrderMatrixBool [dt.ButtonCount][dt.FloorCount] bool


func RunStateMachine(elevatorID int,
		//To statehandler
		driverStateUpdateCh chan<- dt.ElevatorState,
		completedOrdersCh chan<- int,
		//From statehandler
		acceptedOrderCh <-chan dt.OrderType,
		restartCh <-chan int,
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

		var elevator dt.ElevatorState
		var orderMatrix	OrderMatrixBool
		var	doorObstructed bool

		//Internal channels
		doorTimerCh := make(chan bool)

		// Initialize the elevators position
		select{
		case newFloor := <- floorSwitchCh:
				elevator.Floor = newFloor
		default:
				motorDirectionCh <- dt.MovingDown
				newFloor := <- floorSwitchCh
				elevator.Floor = newFloor
				motorDirectionCh <- dt.MovingStopped
		}
		elevator.MovingDirection = dt.MovingStopped
		elevator.State = dt.Idle

		// Run State machine
		for {
				select {
				case newAcceptedOrder:= <- acceptedOrderCh:
						newOrderMatrix, newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, elevator, orderMatrix)
						fmt.Printf("order matrix %v \n", newOrderMatrix)
						if elevator.State != newElevator.State {
								if newElevator.State == dt.Moving  {
										motorDirectionCh <- newElevator.MovingDirection

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

				case newFloor:= <- floorSwitchCh:
						newOrderMatrix, newElevator := updateOnNewFloorArrival(newFloor, elevator, orderMatrix)

						floorIndicatorCh <- newFloor

						if newElevator.State == dt.DoorOpen {
								motorDirectionCh <- dt.MovingStopped
								doorOpenCh <- OPEN_DOOR
								go startDoorTimer(doorTimerCh)
								completedOrdersCh <- newFloor
						}

						elevator = newElevator
						orderMatrix = newOrderMatrix

				case <- doorTimerCh:
						if doorObstructed {
								go startDoorTimer(doorTimerCh)

						} else {
								doorOpenCh <- CLOSE_DOOR
								newElevator := updateOnDoorClosing(elevator, orderMatrix)

								elevator = newElevator
						}

				case <- restartCh:

				case doorObstructed = <- obstructionSwitchCh:

				case <- stopBtnCh:
			}
			// Send updated elevator to statehandler

			driverStateUpdateCh <- elevator // This type does not match the type of the channel
		}
}
