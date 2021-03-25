package scheduler

import (
	dt "../datatypes"
	ed "../elevatordriver"
)

const(
	TRAVEL_TIME		int = 4
	DOOR_OPEN_TIME		= 3
)

func TimeToIdle(elevator dt.ElevatorState, ordermatrix dt.OrderMatrixType) int {
		var duration int = 0

		switch elevator.State {
		case dt.Idle:
				newDirection := ChooseDirection(elevator)
				if newDirection == dt.MovingStopped {
						return duration
				}
		case dt.Moving:
				duration += TRAVEL_TIME / 2
				elevator.Floor += elevator.Direction
		case dt.DoorOpen:
				duration -= DOOR_OPEN_TIME / 2
		}

		for {
			if ElevatorShouldStop(elevator) {
					elevator = ClearOrdersAtCurrentFloor(elevator, nil) // nil means that the orders shouldnt really be cleared. I don't think that i is really necessary 
					duration += DOOR_OPEN_TIME
					elevator.Direction = ChooseDirection(elevator)
					if elevator.Direction == dt.MovingStopped {
							return duration
					}
			}
			elevator.Floor += elevator.Direction
			duration += TRAVEL_TIME
		}
		return duration
}

