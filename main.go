package main

import (
	"flag"
	"time"

	dt "./datatypes"
	"./elevatordriver"
	"./iomodule"
	"./netmodule"
	"./scheduler"
	"./statehandler"
)

func main() {

	elevatorID, port := parseFlag()

	ports := netmodule.NetworkPorts{
		PeerTxPort:  16363,
		PeerRxPort:  16363,
		BcastRxPort: 26363,
		BcastTxPort: 26363,
	}

	stateUpdateCh := make(chan [dt.ElevatorCount]dt.ElevatorState)
	orderUpdateCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	driverStateUpdateCh := make(chan dt.ElevatorState)
	acceptedOrderCh := make(chan dt.OrderType)
	completedOrderFloorCh := make(chan int)

	restartCh := make(chan int)

	newOrdersCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	redirectedOrderCh := make(chan dt.OrderType)

	outgoingStateCh := make(chan dt.ElevatorState)
	incomingStateCh := make(chan dt.ElevatorState)

	outgoingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	incomingOrderCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)

	disconnectCh := make(chan int)
	connectCh := make(chan int)

	motorDirCh := make(chan dt.MoveDirectionType)
	floorIndicatorCh := make(chan int)
	doorOpenCh := make(chan bool)
	stopLampCh := make(chan bool)
	buttonLampCh := make(chan iomodule.ButtonLampType)

	buttonEventCh := make(chan dt.OrderType)
	floorSensorCh := make(chan int)
	stopBtnCh := make(chan bool)
	obstructionSwitchCh := make(chan bool)

	go netmodule.RunNetworkModule(
		elevatorID,
		ports,
		outgoingStateCh, incomingStateCh,
		outgoingOrderCh, incomingOrderCh,
		disconnectCh, connectCh,
	)

	go iomodule.RunIOModule(
		port,
		motorDirCh,
		floorIndicatorCh,
		doorOpenCh,
		stopLampCh,
		buttonLampCh,
		buttonEventCh,
		floorSensorCh,
		stopBtnCh,
		obstructionSwitchCh,
	)

	go statehandler.RunStateHandlerModule(
		elevatorID,
		incomingOrderCh, outgoingOrderCh,
		incomingStateCh, outgoingStateCh,
		disconnectCh, connectCh,
		stateUpdateCh, orderUpdateCh,
		newOrdersCh,
		redirectedOrderCh,
		driverStateUpdateCh,
		acceptedOrderCh, completedOrderFloorCh,
		buttonLampCh,
	)

	go elevatordriver.RunStateMachine(
		elevatorID,
		driverStateUpdateCh,
		completedOrderFloorCh, acceptedOrderCh,
		restartCh,
		floorSensorCh, stopBtnCh, obstructionSwitchCh,
		floorIndicatorCh, motorDirCh, doorOpenCh, stopLampCh,
	)

	go scheduler.RunOrdersScheduler(
		buttonEventCh, redirectedOrderCh,
		stateUpdateCh, orderUpdateCh,
		newOrdersCh,
	)

	for {

		time.Sleep(10 * time.Millisecond)
	}

}

func parseFlag() (int, int) {
	var elevatorID int
	var port int
	flag.IntVar(&elevatorID, "id", 1, "Id of the elevator")
	flag.IntVar(&port, "port", 15657, "IP port to harware server")
	flag.Parse()
	return elevatorID, port
}
