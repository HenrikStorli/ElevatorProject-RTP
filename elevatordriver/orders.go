package elevatordriver

const (
  ACTIVE    = true
  INACTIVE  = false
)

func updateOrder(elevator elevatorType, order dt.OrderType, status bool) elevatorType { // Endre returtype til orderMatrix
    elevator.orderMatrix[order.Button][order.Floor - 1] = status

    return elevator
}

func clearOrdersAtCurrentFloor(elevator elevatorType) elevatorType {
    for btnType := 0; btnType < 3; ++ {
        elevator.orderMatrix[btnType][elevator.currentFloor - 1] = false
    }
    return elevator
}

func anyOrdersAtCurrentFloor(elevator elevatorType) bool {
    for btnType := 0; btnType < 3; ++ {
        if elevator.orderMatrix[btnType][elevator.currentFloor] {
            return true
        }
    }
}


func anyOrdersAbove(elevator elevatorType) bool {
    for floor := elevator.currentFloor; floor < numFloors; ++{
        for orderType := 0; orderType < 3; ++{
            if elevator.orderMatrix[orderType][floor] {
                return true
            }
        }
    }
    return false
}

func anyOrdersBelow(elevator elevatorType) bool {
    for floor:= elevator.currentFloor - 1; floor > 0; --{
        for btnType:= 0; btnType < 3; ++{
            if elevator.orderMatrix[btnType][floor] {
                return true
            }
        }
    }
    return false
}



func cabOrdersAtCurrentFloor(elevator elevatorType) bool {

    if elevator.orderMatrix[dt.BtnCab][floor - 1] {
        return true
    }
    return false
}

func ordersInTravelingDirectionAtCurrentFloor(elevator elevatorType) bool {

    switch(elevator.direction){
    case dt.MovingDown:
        if elevator.orderMatrix[BtnHallDown][elevator.currentFloor - 1] {
            return true
        }

    case dt.MovingUp:
        if elevator.orderMatrix[BtnHallUp][elevator.currentFloor - 1] {
          return true
        }
    }
}