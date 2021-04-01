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

	acceptedOrderSHCh := make(chan statehandler.OrderWithTime)
	completedFloorEDCh := make(chan int)

	acceptedOrderEDCh := make(chan dt.OrderType)
	completedFloorSHCh := make(chan int)

	restartCh := make(chan int)

	scheduledOrdersCh := make(chan [dt.ElevatorCount]dt.OrderMatrixType)
	newOrderCh := make(chan dt.OrderType)

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
		newOrderCh,
		floorSensorCh,
		stopBtnCh,
		obstructionSwitchCh,
	)

	go statehandler.RunOrderWatchdog(acceptedOrderSHCh, acceptedOrderEDCh, completedFloorEDCh, completedFloorSHCh, restartCh)

	go statehandler.RunStateHandlerModule(
		elevatorID,
		incomingOrderCh, outgoingOrderCh,
		incomingStateCh, outgoingStateCh,
		disconnectCh, connectCh,
		stateUpdateCh, orderUpdateCh,
		scheduledOrdersCh,
		newOrderCh,
		driverStateUpdateCh,
		acceptedOrderSHCh, completedFloorSHCh,
	)

	go elevatordriver.RunStateMachine(
		elevatorID,
		driverStateUpdateCh,
		completedFloorEDCh, acceptedOrderEDCh,
		restartCh,
		floorSensorCh, stopBtnCh, obstructionSwitchCh,
		floorIndicatorCh, motorDirCh, doorOpenCh, stopLampCh,
	)

	go scheduler.RunOrdersScheduler(
		elevatorID,
		newOrderCh,
		stateUpdateCh, orderUpdateCh,
		scheduledOrdersCh,
		buttonLampCh,
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
