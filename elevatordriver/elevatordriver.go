package elevatordriver

import (
	dt "../datatypes"
)

	
type elevator struct {
	
}

func RunStateMachine(elevatorID int,
	//Interface towards statehandler
	driverStateUpdateCh chan<- dt.ElevatorState,
	acceptedOrderCh <-chan dt.OrderType,
	completedOrderCh chan<- dt.OrderType,,
	restartCh <-chan int, 
	//Interface towards elevio
	floorSwitchCh <-chan int
	// stopBtnCh <-chan bool,
	// obstructionSwitchCh <-chan bool
	) 
{
	// Local data


	//var currentState dt.MachineStateType
	go listenForAcceptedOrders(acceptedOrderCh)

	//run state machine here
	for {
		select {
		case newAcceptedOrder<-acceptedOrderCh:

		case newFloor <- floorSwitchCh:

		case <- restartCh:


		}
	}
}

func listenForAcceptedOrders(acceptedOrderCh <-chan dt.OrderType) {
	for {
		select {
		case <-acceptedOrderCh:

		}
	}
}
