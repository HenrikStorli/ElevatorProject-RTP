package elevatordriver

const (
  ACTIVE    = true
  INACTIVE  = false
)

func updateOrder(elevator elevatorType, order dt.OrderType, status bool) orderMatrixBool {
    elevator.orderMatrix[order.Button][order.Floor] = status

    return elevator.orderMatrix
}

func clearOrdersAtCurrentFloor(elevator elevatorType) orderMatrixBool {    
    for btnType := 0; btnType < dt.ButtonCount; ++ {
        elevator.orderMatrix[btnType][elevator.currentFloor] = false
    }

    return elevator.orderMatrix
}

func anyOrders(elevator) bool {
        for floor := 0; floor < dt.FloorCount; ++ {
                for btnType := 0; btnType < dt.ButtonCount; ++ {
                        if elevator.orderMatrix[btnType][floor] {
                                return true
                        }
                }
        }

        return false
}

func anyOrdersAtCurrentFloor(elevator elevatorType) bool {
    for btnType := 0; btnType < dt.ButtonCount; ++ {
        if elevator.orderMatrix[btnType][elevator.currentFloor] {
            return true
        }
    }
}


func anyOrdersAbove(elevator elevatorType) bool {                   // Se på denne
    for floor := elevator.currentFloor; floor < numFloors; ++{      // Endre fra orderType til btnType
        for btnType := 0; btnType < dt.ButtonCount; ++{
            if elevator.orderMatrix[btnType][floor] {
                return true
            }
        }
    }
    return false
}

func anyOrdersBelow(elevator elevatorType) bool {               // Se på denne
    for floor:= elevator.currentFloor - 1; floor > 0; --{
        for btnType:= 0; btnType < dt.ButtonCount; ++{
            if elevator.orderMatrix[btnType][floor] {
                return true
            }
        }
    }
    return false
}



func anyCabOrdersAtCurrentFloor(elevator elevatorType) bool {
    if elevator.orderMatrix[dt.BtnCab][floor] {
        return true
    }
    return false
}

func anyOrdersInTravelingDirectionAtCurrentFloor(elevator elevatorType) bool {
    switch(elevator.direction){
    case dt.MovingDown:
        if elevator.orderMatrix[BtnHallDown][elevator.currentFloor] {
            return true
        }

    case dt.MovingUp:
        if elevator.orderMatrix[BtnHallUp][elevator.currentFloor] {
          return true
        }
    }
}