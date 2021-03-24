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
						elevator.direction = chooseDirectionFromIdle(elevator, order)
				}
		case dt.Moving:
				elevator.orderMatrix = updateOrder(elevator, order, ACTIVE)
		case dt.DoorOpen:
				if order.Floor != elevator.currentFloor {
						elevator.orderMatrix = updateOrder(elevator, order, ACTIVE)
				}
		case dt.Error:
		}
}

func updateOnNewFloorArrival(newFloor int, elevator elevatorType)  elevatorType {

		elevator.currentFloor = newFloor

		switch (elevator.state) {
		case dt.Moving:
				if ElevatorShouldStop(elevator) {
						elevator.orderMatrix = clearOrdersAtCurrentFloor(elevator)
						elevator.priorityDirection = elevator.direction
						elevator.direction = dt.MovingStopped
						elevator.state = dt.DoorOpen
				}
		case dt:Error:
			// Test for reinitialize-criteria
		}

		return elevator
}

func updateOnDoorClosing(elevator elevatorType) elevatorType {
		switch(elevator.state){
		case dt.DoorOpen:

				elevator.direction = chooseDirectionFromDoorOpen(elevator)
				
				if elevator.direction == dt.MovingStopped {
						elevator.state = dt.Idle
				} else {
						elevator.state = dt.Moving
				}
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

		if !anyOrders(elevator){
				return dt.MovingStopped
		} 
		else if elevator.currentElevator == dt.FloorCount - 1 && anyOrdersBelow(elevator){
				return dt.MovingDown

		} else if elevator.currentFloor == 0 && anyOrdersAbove(elevator) {
			return dt.MovingUp
			
		} else if {
			
		} else if {
			
		} else if {
			
		} else if {
			
		}


		if elevator.currentFloor == dt.FloorCount && anyOrdersBelow(elevator) {
				return dt.MovingDown

		} else if elevator.currentFloor == 0 && anyOrdersAbove(elevator) {
				return dt.MovingUp
		
		} else if 
}

Software_state elevator_movement_from_idle(int current_floor, HardwareMovement previous_direction){

    if(queue_check_orders_waiting() == NO_ORDERS){
      return Software_state_waiting;
    }
    else if((current_floor == (HARDWARE_NUMBER_OF_FLOORS - 1)) && queue_check_order_below(current_floor)){
      return Software_state_moving_down;
    } 
    else if((current_floor == 0) && queue_check_order_above(current_floor)){
      return Software_state_moving_up;
    } 
    else if((priority == PRIORITY_DOWN) && queue_check_order_below(current_floor)){
      return Software_state_moving_down;
    }
    else if((priority == PRIORITY_UP) && queue_check_order_above(current_floor)){
      return Software_state_moving_up;
    } 
    else if(((previous_direction == HARDWARE_MOVEMENT_UP) && queue_check_order_above(current_floor))){
      return Software_state_moving_up;
    }
    else if((previous_direction == HARDWARE_MOVEMENT_DOWN) && queue_check_order_below(current_floor)){
      return Software_state_moving_down;
    }
    else if ((previous_direction == HARDWARE_MOVEMENT_UP) && queue_check_order_below(current_floor)){
      return Software_state_moving_down;
    }
    else if ((previous_direction == HARDWARE_MOVEMENT_DOWN) && queue_check_order_above(current_floor)){
      return Software_state_moving_up;
    }
    return Software_state_waiting;
}

