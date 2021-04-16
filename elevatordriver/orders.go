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

func ClearOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) OrderMatrixBool {

	for btnIndex, _ := range orderMatrix {
		orderMatrix[btnIndex][elevator.Floor] = INACTIVE
	}

	return orderMatrix
}

func anyOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	for btnIndex, _ := range orderMatrix {
		if orderMatrix[btnIndex][elevator.Floor] {
			return true
		}
	}

	return false
}

func anyOrdersAbove(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	for floor := elevator.Floor + 1; floor < cf.FloorCount; floor++ {
		for btnIndex := 0; btnIndex < cf.ButtonCount; btnIndex++ {
			if orderMatrix[btnIndex][floor] {
				return true
			}
		}
	}

	return false
}

func anyOrdersBelow(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	for floor := elevator.Floor - 1; floor > -1; floor-- {
		for btnIndex := 0; btnIndex < cf.ButtonCount; btnIndex++ {
			if orderMatrix[btnIndex][floor] {
				return true
			}
		}
	}

	return false
}

func anyCabOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	if orderMatrix[dt.BtnCab][elevator.Floor] {
		return true
	}

	return false
}

func anyOrdersInTravelingDirectionAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {

	switch elevator.MovingDirection {
	case dt.MovingDown:
		if orderMatrix[dt.BtnHallDown][elevator.Floor] {
			return true
		}

	case dt.MovingUp:
		if orderMatrix[dt.BtnHallUp][elevator.Floor] {
			return true
		}
	}

	return false
}
