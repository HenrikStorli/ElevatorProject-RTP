package scheduler

import{

}

//Local variables needed in the moduel


func runscheduler(ch_orderFromHardware <-chan ButtonEvent,
                  ch_updatedOrderMatrix chan<- OrderMatrix,
                  ch_elevatorStateReceiver <-chan ElevatorStates){

  // Possibly initialize some values.

    for{
    select
    case newOrder:= <- ch_orderFromHardware:
      chosenElevator:= compareCostFunctions(elevators,newOrder)
      chosenElevator.OrderMatrix[newOrder.Floor][newOrder.Button] = "new"
    case elevators:= <-ch_elevatorStateReceiver
      //Maybe unwrape the data in "elevators" to "elevator1", "elevtor2" etc.
      disfunctioningElevators:= elevatorsNotFunctioning(elevators)
      for floor:= 1; floor < NumFloors; ++{ // Check order at every floor
        for buttonType:= 0; buttonType < 2; ++{ // Check only hall orders
          for every elevator in disfunctioningElevators {
              if disfunctioningElevator.OrderMatrix[floor][buttonType] == "accpted"{
                newOrder:= {floor, buttonType}
              }else if  disfunctioningElevator.OrderMatrix[floor][buttonType] == "new"{
                newOrder:= {floor, buttonType}
              } // else if
              chosenElevator:= compareCostFunctions(elevators, newOrder)
              chosenElevator.OrderMatrix[newOrder.Floor][newOrder.Button] = "new"
              disfunctioningElevator.OrderMatrix[floor][buttonType] = "unknown"
          } // for
        } // for
      } // for
  } //for
} // function
