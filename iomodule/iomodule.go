package iomodule

import (
	"strconv"

	cf "../config"
	dt "../datatypes"
	"./elevio"
)

type ButtonLampType struct {
	Order  dt.OrderType
	TurnOn bool
}

func RunIOModule(
	port int,
	//Input
	motorDirCh <-chan dt.MoveDirectionType,
	floorIndicatorCh <-chan int,
	doorOpenCh <-chan bool,
	stopLampCh <-chan bool,
	buttonLampCh <-chan ButtonLampType,

	//Output
	buttonEventCh chan<- dt.OrderType,
	floorSensorCh chan<- int,
	stopBtnCh chan<- bool,
	obstructionSwitchCh chan<- bool,

) {
	portString := strconv.Itoa(port)
	elevio.Init("localhost:"+portString, cf.FloorCount)

	go elevio.PollButtons(buttonEventCh)
	go elevio.PollFloorSensor(floorSensorCh)
	go elevio.PollStopButton(stopBtnCh)
	go elevio.PollObstructionSwitch(obstructionSwitchCh)

	for {
		select {
		case motorDir := <-motorDirCh:
			elevio.SetMotorDirection(motorDir)
		case buttonLamp := <-buttonLampCh:
			elevio.SetButtonLamp(buttonLamp.Order.Button, buttonLamp.Order.Floor, buttonLamp.TurnOn)
		case floorIndicator := <-floorIndicatorCh:
			elevio.SetFloorIndicator(floorIndicator)
		case doorOpen := <-doorOpenCh:
			elevio.SetDoorOpenLamp(doorOpen)
		case stopLamp := <-stopLampCh:
			elevio.SetStopLamp(stopLamp)
		}
	}
}
