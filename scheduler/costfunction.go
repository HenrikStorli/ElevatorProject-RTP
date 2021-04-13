package scheduler

import (
	cf "../config"
	dt "../datatypes"
	ed "../elevatordriver"
)

func TimeToIdle(elevator dt.ElevatorState, orderMatrix dt.OrderMatrixType) int {
	duration := 0
	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)
	switch elevator.State {
	case dt.Idle:
		newDirection := ed.ChooseDirection(elevator, boolOrderMatrix)
		if newDirection == dt.MovingStopped {
			return duration
		}
	case dt.Moving:
		duration += cf.TravelTime / 2
		elevator.Floor += int(elevator.MovingDirection)
	case dt.DoorOpen:
		duration -= cf.DoorOpenTime / 2
	default:
	}
	tries := 0
	for {
		if ed.ElevatorShouldStop(elevator, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(elevator, boolOrderMatrix) // nil means that the orders shouldnt really be cleared. I don't think that i is really necessary
			duration += cf.DoorOpenTime
			elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
			if elevator.MovingDirection == dt.MovingStopped {
				return duration
			}
		}
		elevator.Floor += int(elevator.MovingDirection)
		duration += cf.TravelTime
		if tries > cf.MaxTries {
			return duration
		}
		tries += 1
	}
}

func convertOrderTypeToBool(orderMatrix dt.OrderMatrixType) ed.OrderMatrixBool {
	var boolMatrix ed.OrderMatrixBool
	for floor := 0; floor < cf.FloorCount; floor++ {
		for btnType := 0; btnType < cf.ButtonCount; btnType++ {
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
		duration += cf.TravelTime / 2
		elevator.Floor += int(elevator.MovingDirection)
		break
	case dt.DoorOpen:
		duration -= cf.DoorOpenTime / 2
	}
	tries := 0
	for {
		if ed.ElevatorShouldStop(elevator, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(elevator, boolOrderMatrix) //requests_clearAtCurrentFloor(elevator, ifEqual);
			if elevator.Floor == newOrder.Floor {
				return duration
			}
			duration += cf.DoorOpenTime
			elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
		}
		elevator.Floor += int(elevator.MovingDirection)
		duration += cf.TravelTime

		if tries > cf.MaxTries {
			return duration
		}
		tries += 1
	}
}
