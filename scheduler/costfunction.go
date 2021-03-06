package scheduler

import{

}

func TimeToIdle(elevator ElevatorStates) int {
  var duartion int = 0

  switch(elevator.state){
  case Idle:
    newDirection = chooseDirection(elevator)
    if newDirection == MovementStop {
      return duration
      }
  case Moving:
    duration += TimeToTravel/2
    elevator.floor += elevator.Direction
  case OpenDoorState.
    duartion -= DoorOpenTime/2
  }


  for {
    if ElevatorShouldStop(elevator) {
      elevator = clearOrdersAtCurrentFloor(elevator, NULL) // NULL means that the orders shouldnt really be cleared.
      duartion += DoorOpenTime
      elevator.Direction = chooseDirection(elevator)
      if elevator.Direction == MovementStop {
        return duration
      }
    }
    e.floor += elevator.Direction
    duration += TimeToTravel
  }
}
