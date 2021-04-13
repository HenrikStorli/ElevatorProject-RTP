package scheduler

import (
	dt "../datatypes"
	ed "../elevatordriver"
)

const (
	TRAVEL_TIME    int = 4
	DOOR_OPEN_TIME     = 3
	MAX_TRIES          = 5000
)


func convertOrderTypeToBool(orderMatrix dt.OrderMatrixType) ed.OrderMatrixBool {
	var boolMatrix ed.OrderMatrixBool
	for floor := 0; floor < dt.FloorCount; floor++ {
		for btnType := 0; btnType < dt.ButtonCount; btnType++ {
			if orderMatrix[btnType][floor] == dt.Accepted {
				boolMatrix[btnType][floor] = ed.ACTIVE
			}
		}
	}
	return boolMatrix
}

func timeToServeRequest(elevator dt.ElevatorState, orderMatrix dt.OrderMatrixType, newOrder dt.OrderType) int {
	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)
	boolOrderMatrix = ed.UpdateOrder(boolOrderMatrix, newOrder, ed.ACTIVE)

	duration := 0

	switch elevator.State {
	case dt.Idle:
		elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
		if elevator.MovingDirection == dt.MovingStopped {
			return duration
		}
		break
	case dt.Moving:
		duration += TRAVEL_TIME / 2
		elevator.Floor += int(elevator.MovingDirection)
		break
	case dt.DoorOpen:
		duration -= DOOR_OPEN_TIME / 2
	}
	tries := 0
	for {
		if ed.ElevatorShouldStop(elevator, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(elevator, boolOrderMatrix) //requests_clearAtCurrentFloor(elevator, ifEqual);
			if elevator.Floor == newOrder.Floor {
				return duration
			}
			duration += DOOR_OPEN_TIME
			elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
		}
		elevator.Floor += int(elevator.MovingDirection)
		duration += TRAVEL_TIME

		if tries > MAX_TRIES {
			return duration
		}
		tries += 1
	}
}
