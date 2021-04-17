package elevatordriver

import (
	cf "../config"
	dt "../datatypes"
)

func ElevatorShouldStop(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	if anyCabOrdersAtCurrentFloor(elevator, orderMatrix) {
		return true

	} else if anyOrdersInTravelingDirectionAtCurrentFloor(elevator, orderMatrix) {
		return true

	} else if anyOrdersAtCurrentFloor(elevator, orderMatrix) {
		if (elevator.MovingDirection == dt.MovingUp) && (!anyOrdersAbove(elevator, orderMatrix)) {
			return true

		} else if (elevator.MovingDirection == dt.MovingDown) && (!anyOrdersBelow(elevator, orderMatrix)) {
			return true
		}

	} else if elevator.Floor == cf.FloorCount-1 {
		return true

	} else if elevator.Floor == 0 {
		return true
	}

	return false
}

func ChooseDirection(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) dt.MoveDirectionType {

	switch elevator.MovingDirection {
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
