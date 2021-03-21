package elevatordriver

import (
	dt "../datatypes"
)

type orderMatrixBool [dt.ButtonCount][dt.FloorCount] bool
	
type elevatorType struct {
	direction			dt.MoveDirectionType
	previousDirection	dt.MoveDirectionType

	currentFloor		int
	floorUp				int

	currentState        dt.MachineStateType
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
						
						

				case newFloor:= <- floorSwitchCh:
						shouldStop, newElevator := updateOnNewFloorArrival(newFloor, currentElevator)
						currentElevator = newElevator

						motorDirectionCh <- currentElevator.direction
						floorIndicatorCh <- newFloor
						
						if shouldStop {
								doorOpenCh <- true
								go startDoorTimer(doorTimerCh)
						}

				case <- restartCh:

				case <- doorTimerCh:

				case currentElevator.doorObstructed <- obstructionSwitchCh:

				case <- stopBtnCh:


			}

			// Send updated elevator to statehandler

		}
}


func updateOnNewAcceptedOrder(order dt.OrderType, oldElevator elevatorType) elevatorType {
		newElevator := oldElevator	

		if oldElevator.currentState != dt.Error {
				newElevator.orderMatrix[order.Button][order.Floor - 1] = true
		}
		return newElevator
}

func updateOnNewFloorArrival(newFloor int, oldElevator elevatorType) (bool, elevatorType) {

	shouldStop := false
	newElevator := oldElevator

	newElevator.currentFloor = newFloor

	switch (oldElevator.state) {
	case dt.Moving:
			if ElevatorShouldStop(newElevator) {
					newElevator := clearOrdersAtFloor(newElevator)
					newElevator.direction = dt.MovingStopped
					newElevator.state = dt.DoorOpen
					shouldStop = true
			}
			return shouldStop, newElevator

	case dt:Error:
		// Test for reinitialize-criteria
	}
	
	return shouldStop, newElevator
}


func ElevatorShouldStop(elevator elevatorType) bool {
		if cabOrdersAtCurrentFloor(elevator) {
				return true

		} else if ordersInTravelingDirectionAtCurrentFloor(elevator) {
				return true 

		} else if anyOrdersAtCurrentFloor(elevator) {
				if elevator.Direction == dt.MovingUp || !anyOrdersAbove(elevator){
						return true

				} else if elevator.Direction == dt.MovingDown || !anyOrdersBelow(elevator) {
						return true
				}
		}
		return false 
		
}
 

func listenForAcceptedOrders(acceptedOrderCh <-chan dt.OrderType) {
	for {
		select {
		case <-acceptedOrderCh:

		}
	}
}
