package scheduler

import (
	cf "../config"
	dt "../datatypes"
	ed "../elevatordriver"
)

// time to execute order
func estimateOrderExecTime(elevator dt.ElevatorState, orderMatrix dt.OrderMatrixType, newOrder dt.OrderType) int {

	simElevatorState := elevator

	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)
	boolOrderMatrix = ed.SetOrder(boolOrderMatrix, newOrder, ed.ACTIVE)

	duration := 0

	switch simElevatorState.State {
	case dt.Idle:
		simElevatorState.MovingDirection = ed.ChooseDirection(simElevatorState, boolOrderMatrix)
		if simElevatorState.MovingDirection == dt.MovingStopped {
			return duration
		}
	case dt.Moving:
		duration += cf.TravelTime / 2
		simElevatorState.Floor += int(simElevatorState.MovingDirection)

	case dt.DoorOpen:
		duration -= cf.DoorOpenTime / 2
	}

	tries := 0

	for {
		if ed.ElevatorShouldStop(simElevatorState, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(simElevatorState, boolOrderMatrix)

			if simElevatorState.Floor == newOrder.Floor {
				return duration
			}

			duration += cf.DoorOpenTime

			simElevatorState.MovingDirection = ed.ChooseDirection(simElevatorState, boolOrderMatrix)
		}

		simElevatorState.Floor += int(simElevatorState.MovingDirection)

		duration += cf.TravelTime

		if tries > cf.MaxTries {
			return duration
		}
		tries += 1
	}
}

func convertOrderTypeToBool(orderMatrix dt.OrderMatrixType) ed.OrderMatrixBool {

	var boolMatrix ed.OrderMatrixBool

	for btnIndex, row := range orderMatrix {
		for floor, order := range row {
			if order == dt.Accepted {
				boolMatrix[btnIndex][floor] = ed.ACTIVE
			}
		}
	}

	return boolMatrix
}
