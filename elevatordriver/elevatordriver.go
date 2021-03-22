package elevatordriver

import (
	dt "../datatypes"
)

const (
	OPEN_DOOR 	= true
	CLOSE_DOOR 	= false
)

type orderMatrixBool [dt.ButtonCount][dt.FloorCount] bool
	
type elevatorType struct {
	direction			dt.MoveDirectionType
	previousDirection	dt.MoveDirectionType

	currentFloor		int
	floorUp				int

	state        		dt.MachineStateType
	previousState		dt.MachineStateType

	orderMatrix			orderMatrixBool

	doorObstructed		bool
}

func RunStateMachine(elevatorID int,
		//To statehandler
		driverStateUpdateCh chan<- dt.ElevatorState,
		completedOrderCh chan<- dt.OrderType,
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
		dootTimerCh := make(chan bool)

		// Initialize the elevators position
		select{
		case currentFloor := <- floorSwitchCh:
				currentState =  ElevatorState_Resting
				if currentFloor != numFloors{
					floorUp = currentFloor + 1
				}
		default:
				SetMotorDirection(MD_Down)
				currentFloor <-ch_floorArrival
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
						currentElevator = newElevator

						if nextState == dt.Moving {
								motorDirectionCh <- currentElevator.direction

						} else if nextState == dt.DoorOpen {
								go startDoorTimer(doorTimerCh)
								doorOpenCh <- OPEN_DOOR
						}

				case newFloor:= <- floorSwitchCh:
						shouldStop, newElevator := updateOnNewFloorArrival(newFloor, currentElevator)
						currentElevator = newElevator

						motorDirectionCh <- currentElevator.direction
						floorIndicatorCh <- newFloor
						
						if shouldStop {
								doorOpenCh <- OPEN_DOOR
								go startDoorTimer(doorTimerCh)
								//Send floorNumber to statehandler
						}

				case <- restartCh:

				case <- doorTimerCh:
						if currentElevator.doorObstructed {
								go startDoorTimer(doorTimerCh)
						
						} else {
								doorOpenCh <- CLOSE_DOOR
								newElevator := updateOnDoorClosing(currentElevator)
						}

				case currentElevator.doorObstructed <- obstructionSwitchCh:

				case <- stopBtnCh:


			}

			// Send updated elevator to statehandler

		}
}



 