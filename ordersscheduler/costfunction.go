package scheduler

import (
	cf "../config"
	dt "../datatypes"
	ed "../elevatordriver"
)

// Simulates the time that this elevator would use to execute the order
func estimateOrderExecTime(elevator dt.ElevatorState, orderMatrix dt.OrderMatrixType, newOrder dt.OrderType) int {

	simElevatorState := elevator

	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)
	boolOrderMatrix = ed.SetOrder(boolOrderMatrix, newOrder, ed.activeOrder)

	duration := 0

	switch simElevatorState.State {
	case dt.IdleState:
		simElevatorState.MovingDirection = ed.ChooseDirection(simElevatorState.MovingDirection, simElevatorState.Floor, boolOrderMatrix)
		//An idle, non moving elevator is always the best choice
		if simElevatorState.MovingDirection == dt.MovingNeutral {
			return duration
		}
	case dt.MovingState:
		duration += cf.TravelTime / 2
		simElevatorState.Floor += int(simElevatorState.MovingDirection)

	case dt.DoorOpenState:
		//An elevator with the door open at the correct floor is also a good choice
		if simElevatorState.Floor == newOrder.Floor {
			return duration
		}
		duration -= cf.DoorOpenTime / 2
	}

	tries := 0

	for {
		if ed.ElevatorShouldStop(simElevatorState.MovingDirection, simElevatorState.Floor, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(simElevatorState.Floor, boolOrderMatrix)

			if simElevatorState.Floor == newOrder.Floor {
				return duration
			}

			duration += cf.DoorOpenTime

			simElevatorState.MovingDirection = ed.ChooseDirection(simElevatorState.MovingDirection, simElevatorState.Floor, boolOrderMatrix)
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
			if order == dt.AcceptedOrder {
				boolMatrix[btnIndex][floor] = ed.activeOrder
			}
		}
	}

	return boolMatrix
}
