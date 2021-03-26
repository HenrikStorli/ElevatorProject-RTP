package elevatordriver

import (
	dt "../datatypes"
)

func updateOnNewAcceptedOrder(order dt.OrderType, elevator ElevatorState, orderMatrix orderMatrixBool)  (orderMatrixBool, ElevatorState) {

		switch(elevator.State){
		case dt.Idle:
				if elevator.Floor == order.Floor {
						nextState := dt.DoorOpen
						elevator.State = nextState
				} else {
						nextState := dt.Moving
						elevator.State = nextState
						orderMatrix = updateOrder(orderMatrix, order, ACTIVE)
						elevator.MovingDirection = ChooseDirection(elevator, orderMatrix)
				}
		case dt.Moving:
				orderMatrix = updateOrder(elevator, order, ACTIVE)
		case dt.DoorOpen:
				if order.Floor != elevator.Floor {
						orderMatrix = updateOrder(elevator, order, ACTIVE)
				}
		case dt.Error:

		default:
		
		}
		return elevator, orderMatrix
}

func updateOnNewFloorArrival(newFloor int, elevator ElevatorState, orderMatrix orderMatrixBool)  (orderMatrixBool, ElevatorState) {

		elevator.Floor = newFloor

		switch (elevator.State) {
		case dt.Moving:
				if ElevatorShouldStop(elevator) {
						orderMatrix = clearOrdersAtCurrentFloor(elevator, orderMatrix)
						//elevator.directionPriority = calculatedirectionPriority(elevator)
						elevator.State = dt.DoorOpen
				}
		//case dt:Error:
			// Test for reinitialize-criteria
		default:
		}

		return elevator, orderMatrix
}

func updateOnDoorClosing(elevator ElevatorState, orderMatrix orderMatrixBool) ElevatorState {
		switch(elevator.State){
		case dt.DoorOpen:
				elevator.MovingDirection = ChooseDirection(elevator, orderMatrix)
				
				if elevator.MovingDirection == dt.MovingStopped {
						elevator.State = dt.Idle
				} else {
						elevator.State = dt.Moving
				}
		default:
		}	
		return elevator
}

func ElevatorShouldStop(elevator ElevatorState, orderMatrix orderMatrixBool) bool {
		if anyCabOrdersAtCurrentFloor(elevator, orderMatrix ) {
				return true

		} else if anyOrdersInTravelingDirectionAtCurrentFloor(elevator, orderMatrix) {
				return true 

		} else if anyOrdersAtCurrentFloor(elevator, orderMatrix ) {
				if elevator.MovingDirection == dt.MovingUp || !anyOrdersAbove(elevator, orderMatrix){
						return true

				} else if elevator.MovingDirection == dt.MovingDown || !anyOrdersBelow(elevator, orderMatrix) {
						return true
				}
		}
		return false 
}

func ChooseDirection(elevator ElevatorState, orderMatrix orderMatrixBool) dt.MoveDirectionType {
		switch(elevator.MovingDirection){
		case dt.MovingUp:
				if anyOrdersAbove(elevator, orderMatrix) {
						return dt.MovingUp
				} else if anyOrdersBelow(elevator, orderMatrix) {
						return dt.MovingDown
				} else {
						return dt.MovingStopped
				}
		case dt.MovingDown:
				if anyOrdersBelow(elevator, orderMatrix) {
						return dt.MovingDown
				} else if anyOrdersAbove(elevator, orderMatrix) {
						return dt.MovingUp
				} else {
						return dt.MovingStopped
				}
		case dt.MovingStopped:
				if anyOrdersBelow(elevator, orderMatrix) {
						return dt.MovingDown
				} else if anyOrdersAbove(elevator, orderMatrix) {
						return dt.MovingUp
				} else {
					return dt.MovingStopped
				}
		default:
				return dt.MovingStopped
		}	
}