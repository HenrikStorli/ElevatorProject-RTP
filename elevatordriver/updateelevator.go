package elevatordriver

func updateOnNewAcceptedOrder(order dt.OrderType, elevator elevatorType) (dt.MachineStateType, elevatorType) {
		
		nextState := dt.None

		switch(elevator.state){
		case dt.Idle:
			if elevator.currentFloor == order.Floor {
					nextState := dt.DoorOpen
					elevator.state = nextState
			
			} else {
					nextState := dt.moving
					elevator.state = nextState

					elevator = updateOrder(elevator, order, ACTIVE)

					elevator.direction = chooseDirectionFromIdle(elevator, order)
			}

		case dt.Moving:
			elevator = updateOrder(elevator, order, ACTIVE)

		case dt.DoorOpen:
				if order.Floor != elevator.currentFloor {
						elevator = updateOrder(elevator, order, ACTIVE)
				}

		case dt.Error:


		}

}

func updateOnNewFloorArrival(newFloor int, elevator elevatorType) (bool, elevatorType) {

		shouldStop := false

		elevator.currentFloor = newFloor

		switch (elevator.state) {
		case dt.Moving:
				if ElevatorShouldStop(elevator) {
						elevator := clearOrdersAtFloor(elevator)
						elevator.direction = dt.MovingStopped
						elevator.state = dt.DoorOpen
						shouldStop = true
				}
				return shouldStop, elevator

		case dt:Error:
			// Test for reinitialize-criteria
}

return shouldStop, newElevator
}

func updateOnDoorClosing(elevator elevatorType) elevatorType {

		switch(elevator.state){
		case dt.DoorOpen:
				elevator.direction = chooseDirection()
				
				if newElevator.direction != dt.Moving {

				}

		}
}


func ElevatorShouldStop(elevator elevatorType) bool {
		if cabOrdersAtCurrentFloor(elevator) {
				return true

		} else if ordersInTravelingDirectionAtCurrentFloor(elevator) {
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

func chooseDirectionFromIdle(elevator elevatorType, order dt.OrderType) dt.MoveDirectionType {
		if order.Floor < elevator.currentFloor {
				return dt.MovingDown

		} else if order.Floor > elevator.currentFloor {
				return dt.MovingUp

		} else {
				return dt.MovingStopped
		}
}


func chooseDirectionFromDoorOpen(elevator elevatorType) dt.MoveDirectionType {
		if elevator.currentFloor == dt.FloorCount && anyOrdersBelow(elevator) {
				return dt.MovingDown

		} else if elevator.currentFloor == 0 && anyOrdersAbove(elevator) {
				return dt.MovingUp
		
		} else if 
}