package scheduler

import (
	dt "../datatypes"
)

func TimeToIdle(elevator dt.ElevatorState, ordermatrix dt.OrderMatrixType) int {
	var duration int = 0

	switch elevator.State {
	case dt.Idle:
		newDirection := chooseDirection(elevator)
		if newDirection == dt.MovingStopped {
			return duration
		}
	case dt.Moving:
		duration += timeToTravel() / 2
		elevator.Floor += int(elevator.MovingDirection)
	case dt.DoorOpen:
		duration -= DoorOpenTime / 2
	}

	for {
		if ElevatorShouldStop(elevator) {
			elevator = clearOrdersAtCurrentFloor(elevator, nil) // nil means that the orders shouldnt really be cleared.
			duration += DoorOpenTime
			elevator.MovingDirection = chooseDirection(elevator)
			if elevator.MovingDirection == dt.MovingStopped {
				return duration
			}
		}
		elevator.Floor += int(elevator.MovingDirection)
		duration += timeToTravel()
	}
	return duration
}

func timeToTravel() int {
	return 0
}
