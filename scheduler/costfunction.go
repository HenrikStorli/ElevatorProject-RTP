package scheduler

func TimeToIdle(elevator ElevatorStates) int {
	var duration int = 0

	switch elevator.state {
	case Idle:
		newDirection = chooseDirection(elevator)
		if newDirection == MovementStop {
			return duration
		}
	case Moving:
		duration += timeToTravel() / 2
		elevator.floor += elevator.Direction
	case OpenDoorState:
		duration -= DoorOpenTime / 2
	}

	for {
		if ElevatorShouldStop(elevator) {
			elevator = clearOrdersAtCurrentFloor(elevator, NULL) // NULL means that the orders shouldnt really be cleared.
			duration += DoorOpenTime
			elevator.Direction = chooseDirection(elevator)
			if elevator.Direction == MovementStop {
				return duration
			}
		}
		e.floor += elevator.Direction
		duration += timeToTravel()
	}
}

func timeToTravel() int {
	return 0
}
