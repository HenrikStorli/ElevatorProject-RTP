package datatypes

//This belongs in a config file
const (
	FloorCount    int = 4
	ElevatorCount     = 3
	ButtonCount       = 3
)

type MoveDirectionType int

const (
	MovingDown    MoveDirectionType = -1
	MovingStopped                   = 0
	MovingUp                        = 1
)

type ButtonType int

const (
	BtnHallUp   ButtonType = 0
	BtnHallDown            = 1
	BtnCab                 = 2
)

type MachineStateType int

type OrderStateType int

const (
	Unknown OrderStateType = iota
	None
	New
	Acknowledged
	Accepted
	Completed
)

type OrderType struct {
	Button ButtonType
	Floor  int
}

//OrderMatrixType is the type for the order matrix
type OrderMatrixType [ButtonCount][FloorCount]OrderStateType

//ElevatorState ...
type ElevatorState struct {
	ElevatorID      int
	MovingDirection MoveDirectionType
	Floor           int
	State           MachineStateType
	IsFunctioning   bool
}

const (
	Init MachineStateType = iota
	Idle
	Moving
	DoorOpen
	Error
)
