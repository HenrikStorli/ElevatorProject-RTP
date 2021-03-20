package iomodule

import (
	dt "../datatypes"
	"./elevio"
)

func RunIOModule(
	//Input
	motorDirCh <-chan dt.MoveDirectionType,
	floorIndicatorCh <-chan int,
	doorOpenCh <-chan bool,
	stopLampCh <-chan bool,

	//Output
	buttonEventCh chan<- dt.OrderType,
	floorSensorCh chan<- int,
	stopBtnCh chan<- bool,
	obstructionSwitchCh chan<- bool,

) {
	elevio.Init("localhost:15657", dt.FloorCount)

	go elevio.PollButtons(buttonEventCh)
	go elevio.PollFloorSensor(floorSensorCh)
	go elevio.PollStopButton(stopBtnCh)
	go elevio.PollObstructionSwitch(obstructionSwitchCh)

	for {
		select {
		case motorDir := <-motorDirCh:
			elevio.SetMotorDirection(motorDir)
		case floorIndicator := <-floorIndicatorCh:
			elevio.SetFloorIndicator(floorIndicator)
		case doorOpen := <-doorOpenCh:
			elevio.SetDoorOpenLamp(doorOpen)
		case stopLamp := <-stopLampCh:
			elevio.SetStopLamp(stopLamp)
		}
	}
}
