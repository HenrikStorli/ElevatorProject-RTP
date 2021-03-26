package scheduler

import (
	dt "../datatypes"
	ed "../elevatordriver"
)

const (
	TRAVEL_TIME    int = 4
	DOOR_OPEN_TIME     = 3
)

func TimeToIdle(elevator dt.ElevatorState, ordermatrix dt.OrderMatrixType) int {
	duration := 0
	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)

	switch elevator.State {
	case dt.Idle:
		newDirection := ed.ChooseDirection(elevator, boolOrderMatrix)
		if newDirection == dt.MovingStopped {
			return duration
		}
	case dt.Moving:
		duration += TRAVEL_TIME / 2
		elevator.Floor += int(elevator.MovingDirection)
	case dt.DoorOpen:
		duration -= DOOR_OPEN_TIME / 2
	}

	for {
		if ed.ElevatorShouldStop(elevator) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(elevator, boolOrderMatrix, nil) // nil means that the orders shouldnt really be cleared. I don't think that i is really necessary
			duration += DOOR_OPEN_TIME
			elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
			if elevator.MovingDirection == dt.MovingStopped {
				return duration
			}
		}
		elevator.Floor += int(elevator.MovingDirection)
		duration += TRAVEL_TIME
	}
	return duration
}

func convertOrderTypeToBool(orderMatrix dt.OrderMatrixType) ed.orderMatrixBool {
		var boolMatrix orderMatrixBool
		for floor := 0; floor < dt.FloorCount; floor ++ {
				for btnType := 0; btnType < dt.ButtonCount; btnType++ {
						if orderMatrix[floor][btnType] == dt.Accepted {
								boolMatrix[floor][btnType] = ed.ACTIVE
						}
				}
		}
		return boolMatrix
}
