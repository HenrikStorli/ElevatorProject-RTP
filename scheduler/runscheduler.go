package scheduler

import(
  "fmt"
  dt "../datatypes"
)

//Local variables needed in the moduel


func runOrdersScheduler(newOrderIOCh <-chan OrderType,
                        newOrderSHCh <-chan OrderType,
                        ElevatorStatesCh <-chan [dt.ElevatorCount]dt.ElevatorState,
                        orderMatricesCh <-chan [dt.ElevatorCount]dt.OrderMatrixType,
                        updateOrderMatricesCh chan<- [dt.ElevatorCount]dt.OrderMatrixType

){
        var elevatorStatesCopy [dt.ElevatorCount]dt.ElevatorState
        var orderMatricesCopy[dt.ElevatorCount]dt.OrderMatrixType

        elevatorStatesCopy <- elevatorStatesCh
        orderMatricesCopy <- orderMatricesCh

        for{
                select{
                case newOrder := <- newOrderIOCh:
						updatedOrderMatrices:= placeOrder(newOrder, elevatorStatesCopy, orderMatricesCopy)
						updateOrderMatricesCh <- updatedOrderMatrices

				case newOrder := <- newOrderSHCh:
						updatedOrderMatrices:= placeOrder(newOrder, elevatorStatesCopy, orderMatricesCopy)
						updateOrderMatricesCh <- updatedOrderMatrices

                case elevatorStatesCopy <- elevatorStatesCh:

                case orderMatricesCopy <- orderMatricesCh:
                }
        }


}


func placeOrder(newOrder dt.OrderType,
				elevatorStates [dt.ElevatorCount]dt.ElevatorState,
				orderMatrices[dt.ElevatorCount]dt.OrderMatrixType
				) [dt.ElevatorCount]dt.OrderMatrixType
{
		var fastestElevatorIndex int = 0
		var fastestExecutionTime int = 1000

		for elevatorIndex, state := range elevatorStates{
				if state.IsFunctioning {
						executionTime: = TimeToIdle(state, orderMatrices[elevatorIndex] )

						if executionTime < fastestExecutionTime {
								fastestExecutionTime = executionTime
								fastestElevatorIndex = elevatorIndex
						}
				}
		}

		var updatedOrderMatrices [dt.ElevatorCount]dt.OrderMatrixType = orderMatrices
		updatedOrderMatrices[fastestElevatorIndex][newOrder.ButtonType][newOrder.floor] = dt.New

		return updatedOrderMatrices
}
