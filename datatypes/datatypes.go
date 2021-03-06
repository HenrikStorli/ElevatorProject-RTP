package datatypes

type MoveDirectionType int

const (
	MovingDown    MoveDirectionType = -1
	MovingStopped                   = 0
	MovingUp                        = 1
)

type MachineStateType int

type OrderStateType int

const (
	Unknown OrderStateType = iota
	New
	Accepted
	Completed
)

const (
	FloorCount    int = 4
	ElevatorCount     = 3
	ButtonCount       = 3
)

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
