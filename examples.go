package tinvestclient

import (
	"fmt"
	"time"
)

// Examples
func Examples(ivToken string)  {

	// Create client
	loClient := Client{}

	// Initialization
	loClient.Init(ivToken)

	// Get accounts
	fmt.Println("My accounts:")

	ltAccounts, loError := loClient.GetAccounts()

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsAccount := range ltAccounts {
		fmt.Printf("%+v\n", lsAccount)
	}

	// Get positions
	fmt.Println("My positions:")

	ltPositions, loError := loClient.GetPositions()

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsPosition := range ltPositions {
		fmt.Printf("%+v\n", lsPosition)
	}

	// Get operations
	fmt.Println("My operations:")

	ltOperations, loError := loClient.GetOperations(FigiAAPL, time.Now().AddDate(0, 0, -365), time.Now() )

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsOperation := range ltOperations {
		fmt.Printf("%+v\n", lsOperation)
	}

	// Get stock shares(similar for GetCurrecnies, GetBonds and GetETFs)
	fmt.Println("Shares:")

	ltShares, loError := loClient.GetShares()

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsShare := range ltShares {
		fmt.Printf("%+v\n", lsShare)
	}

	// Get candles
	fmt.Println("Candles:")

	ltCandles, loError := loClient.GetCandles(FigiAAPL, IntervalDay, time.Now().AddDate(0, 0, -10), time.Now())

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsCandle := range ltCandles {
		fmt.Printf("%+v\n", lsCandle)
	}

	// Create limit order
	fmt.Println("Create limit order:")

	lvOrderID, loError := loClient.CreateLimitOrder(FigiAAPL, OperationBuy, 1, 100 )

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	fmt.Printf("Order ID %v was created\n", lvOrderID)

	// Get orders
	fmt.Println("Orders:")

	ltOrders, loError := loClient.GetOrders()

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	for _, lsOrder := range ltOrders{
		fmt.Printf("%+v\n", lsOrder)
	}

	// Cancel order
	fmt.Println("Cancel order:")

	loError = loClient.CancelOrder(lvOrderID)

	if loError != nil {
		fmt.Printf("Error: %+v\n", loError)
		return
	}

	fmt.Printf("Order ID %v was canceled\n", lvOrderID)

}