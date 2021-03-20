package elevatordriver

func updateOrder(orderMatrix orderMatrixBool, order dt.OrderType, status bool){
		orderMatrix[order.Button][order.Floor] = status
}

func anyOrders() bool {
  for floor: = 0; floor < _numFloors; ++{
    for orderType := 0; orderType < 3; ++ {
      if orderButtonMatrix[orderType][floor] == 1
        return true
    }
  }
  return false
}

func anyOrdersAbove(currentfloor int) bool {
  if currentFloor == _numFloors - 1{
    return false
  }
  for floor:= currentFloor; floor < _numFloors; ++{
    for orderType:= 0; orderType < 3; ++{
      if orderButtonMatrix[orderType][floor] == 1
        return true
    }
  }
  return false
}

func anyOrdersBelow(currentfloor int) bool {
  if currentFloor == 1{
    return false
  }
  for floor:= currentFloor - 2; floor > -1; --{
    for orderType:= 0; orderType < 3; ++{
      if orderButtonMatrix[orderType][floor] == 1
        return true
    }
  }
  return false
}

func anyOrdersAtFloor(floor int) bool {
  for orderType:= 0; orderType < 3; ++{
    if orderButtonMatrix[orderType][floor - 1] == 1{
      return true
    }
  }
  return false
}

func anyOrdersOfType(orderType ButtonType) bool{
  for floor:= 0; floor < _numFloors; ++{
    if orderButtonMatrix[orderType][floor] == 1
    return true
  }
  return false
}

func clearOrdersOnFloor(floor int){
  for orderType:= 0; orderType < 3; ++{
    updateOrder(ButtonEvent{floor, orderType}, 0)
  }
}
