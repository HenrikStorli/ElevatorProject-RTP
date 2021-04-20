package elevatordriver

import (
	cf "../config"
	dt "../datatypes"
)

const (
	activeOrder   = true
	inActiveOrder = false
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

func SetOrder(orderMatrix OrderMatrixBool, order dt.OrderType, status bool) OrderMatrixBool {

	orderMatrix[order.Button][order.Floor] = status
	return orderMatrix
}

func ClearOrdersAtCurrentFloor(currentFloor int, orderMatrix OrderMatrixBool) OrderMatrixBool {

	for btnIndex, _ := range orderMatrix {
		orderMatrix[btnIndex][currentFloor] = inActiveOrder
	}

	return orderMatrix
}

func anyOrdersAtCurrentFloor(currentFloor int, orderMatrix OrderMatrixBool) bool {

	for btnIndex, _ := range orderMatrix {
		if orderMatrix[btnIndex][currentFloor] {
			return true
		}
	}

	return false
}

func anyOrdersAbove(currentFloor int, orderMatrix OrderMatrixBool) bool {

	for floor := currentFloor + 1; floor < cf.FloorCount; floor++ {
		for btnIndex := 0; btnIndex < cf.ButtonCount; btnIndex++ {
			if orderMatrix[btnIndex][floor] {
				return true
			}
		}
	}

	return false
}

func anyOrdersBelow(currentFloor int, orderMatrix OrderMatrixBool) bool {

	for floor := currentFloor - 1; floor > -1; floor-- {
		for btnIndex := 0; btnIndex < cf.ButtonCount; btnIndex++ {
			if orderMatrix[btnIndex][floor] {
				return true
			}
		}
	}

	return false
}

func anyCabOrdersAtCurrentFloor(currentFloor int, orderMatrix OrderMatrixBool) bool {

	if orderMatrix[dt.ButtonCab][currentFloor] {
		return true
	}

	return false
}

func anyOrdersInTravelingDirectionAtCurrentFloor(movingDirection dt.MoveDirectionType, currentFloor int, orderMatrix OrderMatrixBool) bool {

	switch movingDirection {
	case dt.MovingDown:
		if orderMatrix[dt.ButtonHallDown][currentFloor] {
			return true
		}

	case dt.MovingUp:
		if orderMatrix[dt.ButtonHallUp][currentFloor] {
			return true
		}
	}

	return false
}
