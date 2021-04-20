package elevatordriver

import (
	cf "../config"
	dt "../datatypes"
)

func ElevatorShouldStop(movingDirection dt.MoveDirectionType, currentFloor int, orderMatrix OrderMatrixBool) bool {

	if anyCabOrdersAtCurrentFloor(currentFloor, orderMatrix) {
		return true

	} else if anyOrdersInTravelingDirectionAtCurrentFloor(movingDirection, currentFloor, orderMatrix) {
		return true

	} else if anyOrdersAtCurrentFloor(currentFloor, orderMatrix) {
		if (movingDirection == dt.MovingUp) && (!anyOrdersAbove(currentFloor, orderMatrix)) {
			return true

		} else if (movingDirection == dt.MovingDown) && (!anyOrdersBelow(currentFloor, orderMatrix)) {
			return true
		}

	} else if currentFloor == cf.FloorCount-1 {
		return true

	} else if currentFloor == 0 {
		return true
	}

	return false
}

func ChooseDirection(movingDirection dt.MoveDirectionType, currentFloor int, orderMatrix OrderMatrixBool) dt.MoveDirectionType {

	switch movingDirection {
	case dt.MovingUp:
		if anyOrdersAbove(currentFloor, orderMatrix) {
			return dt.MovingUp
		} else if anyOrdersBelow(currentFloor, orderMatrix) {
			return dt.MovingDown
		} else {
			return dt.MovingNeutral
		}

	case dt.MovingDown:
		if anyOrdersBelow(currentFloor, orderMatrix) {
			return dt.MovingDown
		} else if anyOrdersAbove(currentFloor, orderMatrix) {
			return dt.MovingUp
		} else {
			return dt.MovingNeutral
		}

	case dt.MovingNeutral:
		if anyOrdersBelow(currentFloor, orderMatrix) {
			return dt.MovingDown
		} else if anyOrdersAbove(currentFloor, orderMatrix) {
			return dt.MovingUp
		} else {
			return dt.MovingNeutral
		}

	default:
		return dt.MovingNeutral
	}
}
