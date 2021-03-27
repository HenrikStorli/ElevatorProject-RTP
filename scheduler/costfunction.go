package scheduler

import (
	dt "../datatypes"
	ed "../elevatordriver"
)

const (
	TRAVEL_TIME    int = 4
	DOOR_OPEN_TIME     = 3
	MAX_TRIES          = 100
)

func TimeToIdle(elevator dt.ElevatorState, orderMatrix dt.OrderMatrixType) int {
	duration := 0
	boolOrderMatrix := convertOrderTypeToBool(orderMatrix)
	//fmt.Println("Inside TimeToIdle")
	switch elevator.State {
	case dt.Idle:
		newDirection := ed.ChooseDirection(elevator, boolOrderMatrix)
		if newDirection == dt.MovingStopped {
			//fmt.Println("TimeToIdle first retun value")
			return duration
		}
	case dt.Moving:
		duration += TRAVEL_TIME / 2
		elevator.Floor += int(elevator.MovingDirection)
	case dt.DoorOpen:
		duration -= DOOR_OPEN_TIME / 2
	default:
	}
	//fmt.Println("Before For loop in costfunc")
	for {
		//fmt.Println("For loop in costfunc")
		if ed.ElevatorShouldStop(elevator, boolOrderMatrix) {
			boolOrderMatrix = ed.ClearOrdersAtCurrentFloor(elevator, boolOrderMatrix) // nil means that the orders shouldnt really be cleared. I don't think that i is really necessary
			duration += DOOR_OPEN_TIME
			elevator.MovingDirection = ed.ChooseDirection(elevator, boolOrderMatrix)
			if elevator.MovingDirection == dt.MovingStopped {
				return duration
			}
		}
		elevator.Floor += int(elevator.MovingDirection)
		////fmt.Printf("Elevator floor is: %v ", elevator.Floor)
		duration += TRAVEL_TIME
		if duration/TRAVEL_TIME > MAX_TRIES {
			return duration
		}
	}
}

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
