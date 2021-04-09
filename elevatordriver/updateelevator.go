package elevatordriver

import (
	dt "../datatypes"
)

func updateOnNewAcceptedOrder(order dt.OrderType, elevator dt.ElevatorState, orderMatrix OrderMatrixBool)  (OrderMatrixBool, dt.ElevatorState) {

		switch(elevator.State){
		case dt.Idle:
				if elevator.Floor == order.Floor {
						nextState := dt.DoorOpen
						elevator.State = nextState
				} else {
						nextState := dt.Moving
						elevator.State = nextState
						orderMatrix = UpdateOrder(orderMatrix, order, ACTIVE)
						elevator.MovingDirection = ChooseDirection(elevator, orderMatrix)
				}
		case dt.Moving:
				orderMatrix = UpdateOrder(orderMatrix, order, ACTIVE)
		case dt.DoorOpen:
				if order.Floor != elevator.Floor {
						orderMatrix = UpdateOrder(orderMatrix, order, ACTIVE)
				}
		case dt.Error:

		default:

		}
		return orderMatrix, elevator
}

func updateOnNewFloorArrival(newFloor int, elevator dt.ElevatorState, orderMatrix OrderMatrixBool)  (OrderMatrixBool, dt.ElevatorState) {

		elevator.Floor = newFloor

		switch (elevator.State) {
		case dt.Moving:
				if ElevatorShouldStop(elevator, orderMatrix) {
						orderMatrix = ClearOrdersAtCurrentFloor(elevator, orderMatrix)
						//elevator.directionPriority = calculatedirectionPriority(elevator)
						elevator.State = dt.DoorOpen
				}
		//case dt:Error:
			// Test for reinitialize-criteria
		default:
		}

		return orderMatrix, elevator
}

func updateOnDoorClosing(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) dt.ElevatorState {
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

func ElevatorShouldStop(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {
		if anyCabOrdersAtCurrentFloor(elevator, orderMatrix ) {
				return true

		} else if anyOrdersInTravelingDirectionAtCurrentFloor(elevator, orderMatrix) {
				return true

		} else if anyOrdersAtCurrentFloor(elevator, orderMatrix ) {
				if (elevator.MovingDirection == dt.MovingUp) && (!anyOrdersAbove(elevator, orderMatrix)){
						return true

				} else if (elevator.MovingDirection == dt.MovingDown) && (!anyOrdersBelow(elevator, orderMatrix)) {
						return true
				}
		} else if elevator.Floor == dt.FloorCount - 1{
				return true
		} else if elevator.Floor == 0 {
				return true
		}
		return false
}

func ChooseDirection(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) dt.MoveDirectionType {
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
