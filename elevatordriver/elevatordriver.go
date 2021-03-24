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
		var currentElevator elevatorType

		//Internal channels
		doorTimerCh := make(chan bool)

		// Initialize the elevators position
		select{
		case currentFloor := <- floorIndicatorCh:
				currentState =  ElevatorState_Resting
				if currentFloor != numFloors{
					floorUp = currentFloor + 1
				}
		default:
				SetMotorDirection(MD_Down)
				currentFloor <-floorIndicatorCh
				SetMotorDirection(MD_Stop)
				if currentFloor != numFloors{
					floorUp = currentFloor + 1
				}
				previousDirection = MD_Down
		}
		currentDirection = MD_Stop


		//run state machine here
		for {
				select {
				case newAcceptedOrder:= <- acceptedOrderCh:
						newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, currentElevator)

						if currentElevator.state != newElevator.state {
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

						currentElevator = newElevator

				case newFloor:= <- floorSwitchCh:

						newElevator := updateOnNewFloorArrival(newFloor, currentElevator)
						
						floorIndicatorCh <- newFloor
						
						if newElevator.state == dt.DoorOpen {
								motorDirectionCh <- dt.MovingStopped
								doorOpenCh <- OPEN_DOOR
								go startDoorTimer(doorTimerCh)
								completedOrdersCh <- newFloor
						}

						currentElevator = newElevator

				case <- restartCh:

				case <- doorTimerCh:
						if currentElevator.doorObstructed {
								go startDoorTimer(doorTimerCh)
						
						} else {
								doorOpenCh <- CLOSE_DOOR
								newElevator := updateOnDoorClosing(currentElevator)

								currentElevator = newElevator
						}

				case currentElevator.doorObstructed <- obstructionSwitchCh:

				case <- stopBtnCh:


			}
			// Send updated elevator to statehandler

		}
}



 