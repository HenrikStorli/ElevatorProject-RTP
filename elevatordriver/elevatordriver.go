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
}

func RunStateMachine(elevatorID int,
		//Interface towards statehandler
		driverStateUpdateCh chan<- dt.ElevatorState,
		acceptedOrderCh <-chan dt.OrderType,
		completedOrderCh chan<- dt.OrderType,,
		restartCh <-chan int, 
		//Interface towards elevio
		floorSwitchCh <-chan int,
		floorIndicatorCh chan<- int,
		motorDirectionCh chan<- dt.MoveDirectionType,
		doorOpenCh chan<- bool
		// setStopCh chan <- bool,
		// stopBtnCh <-chan bool,
		// obstructionSwitchCh <-chan bool
		) 
{
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

		// Local data
		var currentElevator elevatorType

		//run state machine here
		for {
				select {
				case newAcceptedOrder:= <- acceptedOrderCh:
						newElevator := updateOnNewAcceptedOrder(newAcceptedOrder, currentElevator)
						currentElevator = newElevator

				case newFloor:= <- floorSwitchCh:
						shouldStop, newElevator := updateOnNewFloorArrival(newFloor, currentElevator)

						newDirection := calculateNewDirection(newFloor, currentElevator)
						currentElevator.direction = newDirection
						motorDirectionCh <- newDirection

						if newDirection == dt.MovingStopped {
							newOrderMatrix := clearOrdersOnFloor(newFloor)


						}
					


						floorIndicatorCh <- newFloor
				case <- restartCh:


			}
		}
}


func updateOnNewAcceptedOrder(order dt.OrderType, oldElevator elevatorType) elevatorType {
		newElevator := oldElevator	

		if oldElevator.currentState != dt.Error {
				newElevator.orderMatrix[order.Button][order.Floor] = true
		}
		return newElevator
}

func updateOnNewFloorArrival(newFloor int, oldElevator elevatorType) (dt.MoveDirectionType, elevatorType) {

		newElevator := oldElevator
		var newDirection dt.MoveDirectionType

		newElevator.currentFloor = newFloor

		switch (oldElevator.state) {
		case dt.Moving:
				if ElevatorShouldStop(newElevator) {
						newElevator.direction = dt.MovingStopped
						newElevator.state = dt.DoorOpen
				}

		case dt:Error:
			// Test for reinitialize-criteria
		}
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
		
}
 

func listenForAcceptedOrders(acceptedOrderCh <-chan dt.OrderType) {
	for {
		select {
		case <-acceptedOrderCh:

		}
	}
}
