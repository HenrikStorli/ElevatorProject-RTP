package elevatordriver

func updateOnNewAcceptedOrder(order dt.OrderType, elevator elevatorType)  elevatorType {

		switch(elevator.state){
		case dt.Idle:
				if elevator.currentFloor == order.Floor {
						nextState := dt.DoorOpen
						elevator.state = nextState
				} else {
						nextState := dt.Moving
						elevator.state = nextState
						elevator.orderMatrix = updateOrder(elevator, order, ACTIVE)
						elevator.direction = chooseDirection(elevator)
				}
		case dt.Moving:
				elevator.orderMatrix = updateOrder(elevator, order, ACTIVE)
		case dt.DoorOpen:
				if order.Floor != elevator.currentFloor {
						elevator.orderMatrix = updateOrder(elevator, order, ACTIVE)
				}
		case dt.Error:

		default:
		
		}
}

func updateOnNewFloorArrival(newFloor int, elevator elevatorType)  elevatorType {

		elevator.currentFloor = newFloor

		switch (elevator.state) {
		case dt.Moving:
				if ElevatorShouldStop(elevator) {
						elevator.orderMatrix = clearOrdersAtCurrentFloor(elevator)
						//elevator.directionPriority = calculatedirectionPriority(elevator)
						elevator.state = dt.DoorOpen
				}
		case dt:Error:
			// Test for reinitialize-criteria
		default:

		}

		return elevator
}

func updateOnDoorClosing(elevator elevatorType) elevatorType {
		switch(elevator.state){
		case dt.DoorOpen:
				elevator.direction = chooseDirection(elevator)
				
				if elevator.direction == dt.MovingStopped {
						elevator.state = dt.Idle
				} else {
						elevator.state = dt.Moving
				}
		default:
		}	
}

func ElevatorShouldStop(elevator elevatorType) bool {
		if anyCabOrdersAtCurrentFloor(elevator) {
				return true

		} else if anyOrdersInTravelingDirectionAtCurrentFloor(elevator) {
				return true 

		} else if anyOrdersAtCurrentFloor(elevator) {
				if elevator.Direction == dt.MovingUp || !anyOrdersAbove(elevator){
						return true

				} else if elevator.Direction == dt.MovingDown || !anyOrdersBelow(elevator) {
						return true
				}
		}
		return false 
}

func ChooseDirection(elevator elevatorType) dt.MoveDirectionType {
		switch(elevator.direction){
		case dt:MovingUp:
				if anyOrdersAbove(elevator) {
						return dt.MovingUp
				} else if anyOrdersBelow(elevator) {
						return dt.MovingDown
				} else {
						return dt.MovingStopped
				}
		case dt:MovingDown:
				if anyOrdersBelow(elevator) {
						return dt.MovingDown
				} else if anyOrdersAbove(elevator) {
						return dt.MovingUp
				} else {
						return dt.MovingStopped
				}
		case dt.MovingStopped:
				if anyOrdersBelow(elevator) {
						return dt.MovingDown
				} else if anyOrdersAbove(elevator) {
						return dt.MovingUp
				} else {
					return dt.MovingStopped
				}
		default:
				return dt:MovingStopped
		}	
}