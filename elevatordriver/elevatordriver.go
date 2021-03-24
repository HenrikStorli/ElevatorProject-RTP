package elevatordriver

import (
	dt "../datatypes"
)

const (
	OPEN_DOOR 	= true
	CLOSE_DOOR 	= false
)

type directionPriorityType

const (
	PRIORITY_DOWN	directionPriorityType = -1
	PRIORITY_NONE                         = 0
	PRIORITY_UP                           = 1
)

type orderMatrixBool [dt.ButtonCount][dt.FloorCount] bool
	
type elevatorType struct {
	direction			dt.MoveDirectionType
	//directionPriority	directionPriorityType
	state        		dt.MachineStateType
	orderMatrix			orderMatrixBool
	currentFloor		int
	doorObstructed		bool
}

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
		setStopCh chan<- bool,		
		) 
{
		// Local data
		var elevator elevatorType

		//Internal channels
		doorTimerCh := make(chan bool)

		// Initialize the elevators position
		select{
		case newFloor := <- floorIndicatorCh:
		default:
				motorDirectionCh <- dt.MovingDown
				newFloor := <- floorIndicatorCh
				motorDirectionCh <- dt.MovingStopped
		}
		elevator.direction = dt.MovingStopped
		elevator.currentFloor = newFloor
		elevator.state = dt.Idle
		
		// Run state machine
		for {
				select {
				case newAcceptedOrder:= <- acceptedOrderCh:
						newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, elevator)

						if elevator.state != newElevator.state {
								if newElevator.state == dt.Moving  {
										motorDirectionCh <- newElevator.direction

								} else if newElevator.state == dt.DoorOpen {
										go startDoorTimer(doorTimerCh)
										doorOpenCh <- OPEN_DOOR
										completedOrdersCh <- newElevator.currentFloor
								}

						} else {
								if newElevator.state == dt.DoorOpen {
										completedOrdersCh <- newElevator.currentFloor
								}
						}

						elevator = newElevator

				case newFloor:= <- floorSwitchCh:
						newElevator := updateOnNewFloorArrival(newFloor, elevator)
						
						floorIndicatorCh <- newFloor
						
						if newElevator.state == dt.DoorOpen {
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

				case elevator.doorObstructed <- obstructionSwitchCh:

				case <- stopBtnCh:
			}
			// Send updated elevator to statehandler
			driverStateUpdateCh <- elevator // This type does not match the type of the channel
		}
}



 