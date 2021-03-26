package elevatordriver

import (
	dt "../datatypes"
)

const (
	OPEN_DOOR 	= true
	CLOSE_DOOR 	= false
)

type orderMatrixBool [dt.ButtonCount][dt.FloorCount] bool


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

		var elevator ElevatorState
		var orderMatrix	orderMatrixBool
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
						orderMatrix, newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, elevator, orderMatrix)

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

				case newFloor:= <- floorSwitchCh:
						orderMatrix, newElevator := updateOnNewFloorArrival(newFloor, elevator, orderMatrix)
						
						floorIndicatorCh <- newFloor
						
						if newElevator.State == dt.DoorOpen {
								motorDirectionCh <- dt.MovingStopped
								doorOpenCh <- OPEN_DOOR
								go startDoorTimer(doorTimerCh)
								completedOrdersCh <- newFloor
						}

						elevator = newElevator

				case <- doorTimerCh:
						if elevator.doorObstructed {
								go startDoorTimer(doorTimerCh)
						
						} else {
								doorOpenCh <- CLOSE_DOOR
								newElevator := updateOnDoorClosing(elevator)

								elevator = newElevator
						}

				case <- restartCh:

				case elevator.doorObstructed = <- obstructionSwitchCh:

				case <- stopBtnCh:
			}
			// Send updated elevator to statehandler
			//driverStateUpdateCh <- elevator // This type does not match the type of the channel
		}
}



 