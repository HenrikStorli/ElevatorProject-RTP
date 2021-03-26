package elevatordriver

import (
	dt "../datatypes"
)

const (
  ACTIVE    = true
  INACTIVE  = false
)

func updateOrder(orderMatrix OrderMatrixBool, order dt.OrderType, status bool) OrderMatrixBool {
        orderMatrix[order.Button][order.Floor] = status
        return orderMatrix
}

func ClearOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) OrderMatrixBool {    
        for btnType := 0; btnType < dt.ButtonCount; btnType++ {
                orderMatrix[btnType][elevator.Floor] = INACTIVE
        }
        return orderMatrix
}

func anyOrders(orderMatrix OrderMatrixBool) bool {
        for floor := 0; floor < dt.FloorCount; floor++ {
                for btnType := 0; btnType < dt.ButtonCount; btnType++ {
                        if orderMatrix[btnType][floor] {
                                return true
                        }
                }
        }
        return false
}

func anyOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {
        for btnType := 0; btnType < dt.ButtonCount; btnType++ {
                if orderMatrix[btnType][elevator.Floor] {
                        return true
                }
        }
        return false
}

func anyOrdersAbove(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {                   
        for floor := elevator.Floor + 1; floor < dt.FloorCount; floor++ {      
                for btnType := 0; btnType < dt.ButtonCount; btnType++ {
                        if orderMatrix[btnType][floor] {
                                return true
                        }
                }
        }
        return false
}

func anyOrdersBelow(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {              
        for floor := elevator.Floor - 1; floor > -1; floor-- {
                for btnType := 0; btnType < dt.ButtonCount; btnType++ {
                        if orderMatrix[btnType][floor] {
                                return true
                        }
                }
        }
        return false
}

func anyCabOrdersAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {
        if orderMatrix[dt.BtnCab][elevator.Floor] {
                return true
        }
        return false
}

func anyOrdersInTravelingDirectionAtCurrentFloor(elevator dt.ElevatorState, orderMatrix OrderMatrixBool) bool {
        switch(elevator.MovingDirection){
        case dt.MovingDown:
                if orderMatrix[dt.BtnHallDown][elevator.Floor] {
                        return true
                }

        case dt.MovingUp:
                if orderMatrix[dt.BtnHallUp][elevator.Floor] {
                        return true
                }
        }
        return false
}