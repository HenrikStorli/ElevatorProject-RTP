package elevatordriver

import (
	cf "../config"
	dt "../datatypes"
)

const (
	ACTIVE   = true
	INACTIVE = false
)

func SetOrder(orderMatrix OrderMatrixBool, order dt.OrderType, status bool) OrderMatrixBool {

	orderMatrix[order.Button][order.Floor] = status
	return orderMatrix
}

func ClearOrdersAtCurrentFloor(currentFloor int, orderMatrix OrderMatrixBool) OrderMatrixBool {

	for btnIndex, _ := range orderMatrix {
		orderMatrix[btnIndex][currentFloor] = INACTIVE
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
