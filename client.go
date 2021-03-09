package tinvestclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const (
	IntervalMin1    = "1min"
	IntervalMin2    = "2min"
	IntervalMin3    = "3min"
	IntervalMin5    = "5min"
	IntervalMin10   = "10m  in"
	IntervalMin15   = "15min"
	IntervalMin30   = "30min"
	IntervalHour    = "hour"
	IntervalDay     = "day"
	IntervalWeek    = "week"
	IntervalMonth   = "month"
	OperationBuy    = "Buy"
	OperationSell   = "Sell"
	CurrencyRUB     = "RUB"
	CurrencyUSD     = "USD"
	TickerTCS       = "TCS"
	TickerTCSG      = "TCSG"
	FigiTCS         = "BBG005DXJS36"
	FigiTCSG        = "BBG00QPYJ5H0"
	statusError     = "Error"
	orderTypeLimit  = "limit"
	orderTypeMarket = "market"
)

type Client struct {
	mvUrl   string
	mvToken string
}

type Account struct {
	ID   string
	Type string
}

type Instrument struct {
	Type              string
	Ticker            string
	FIGI              string
	ISIN              string
	Text              string
	Currency          string
	Lot               int
	MinPriceIncrement float64
}

type Candle struct {
	Time       time.Time
	High       float64
	Open       float64
	Close      float64
	Low        float64
	Volume     float64
	ShadowHigh float64
	ShadowLow  float64
	Body       float64
	Type       string
}

func (self *Candle) eval() {

	if self.Open < self.Close {
		self.Type = "Green"
		self.ShadowHigh = self.High - self.Close
		self.Body = self.Close - self.Open
		self.ShadowLow = self.Open - self.Low
	} else {
		self.Type = "Red"
		self.ShadowHigh = self.High - self.Open
		self.Body = self.Open - self.Close
		self.ShadowLow = self.Close - self.Low
	}

}

type Position struct {
	FIGI     string
	Ticker   string
	Type     string
	Text     string
	Quantity float64
	Lots     int
	Currency string
	Price    float64
	Profit   float64
}

type Operation struct {
	ID         string
	Type       string
	FIGI       string
	Quantity   float64
	Price      float64
	Value      float64
	Commission float64
	Currency   string
	Date       time.Time
}

type Order struct {
	ID            string
	FIGI          string
	Type          string
	Operation     string
	Price         float64
	Status        string
	RequestedLots int
	ExecutedLots  int
}

func (self *Client) Init(token string) {

	self.mvUrl = "https://api-invest.tinkoff.ru/openapi/"
	self.mvToken = token

}

func (self *Client) GetAccounts() (rtAccounts []Account, roError error) {

	lvUrl := self.mvUrl + "user/accounts"

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message  string `json:"message"`
			Code     string `json:"code"`
			Accounts []struct {
				BrokerAccountType string `json:"brokerAccountType"`
				BrokerAccountID   string `json:"brokerAccountId"`
			} `json:"accounts"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseAccount := range lsResponse.Payload.Accounts {

		lsAccount := Account{}

		lsAccount.ID = lsResponseAccount.BrokerAccountID
		lsAccount.Type = lsResponseAccount.BrokerAccountType

		rtAccounts = append(rtAccounts, lsAccount)

	}

	return

}

func (self *Client) GetCurrencies() (rtCurrencies []Instrument, roError error) {

	rtCurrencies, roError = self.getInstruments("currencies")

	return

}

func (self *Client) GetShares() (rtShares []Instrument, roError error) {

	rtShares, roError = self.getInstruments("stocks")

	return

}

func (self *Client) GetBonds() (rtBonds []Instrument, roError error) {

	rtBonds, roError = self.getInstruments("bonds")

	return

}

func (self *Client) GetETFs() (rtETFs []Instrument, roError error) {

	rtETFs, roError = self.getInstruments("etfs")

	return

}

func (self *Client) getInstruments(ivType string) (rtInstruments []Instrument, roError error) {

	lvUrl := self.mvUrl + "market/" + ivType

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message     string `json:"message"`
			Code        string `json:"code"`
			Total       int    `json:"total"`
			Instruments []struct {
				Figi              string  `json:"figi"`
				Ticker            string  `json:"ticker"`
				Isin              string  `json:"isin"`
				MinPriceIncrement float64 `json:"minPriceIncrement"`
				Lot               int     `json:"lot"`
				MinQuantity       int     `json:"minQuantity"`
				Currency          string  `json:"currency"`
				Name              string  `json:"name"`
				Type              string  `json:"type"`
			} `json:"instruments"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	//log.Println(string(lvBodyBytes))

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseInstrument := range lsResponse.Payload.Instruments {

		lsInstrument := Instrument{}

		lsInstrument.Type = lsResponseInstrument.Type
		lsInstrument.FIGI = lsResponseInstrument.Figi
		lsInstrument.Ticker = lsResponseInstrument.Ticker
		lsInstrument.Text = lsResponseInstrument.Name
		lsInstrument.Currency = lsResponseInstrument.Currency
		lsInstrument.Lot = lsResponseInstrument.Lot
		lsInstrument.MinPriceIncrement = lsResponseInstrument.MinPriceIncrement

		rtInstruments = append(rtInstruments, lsInstrument)

	}

	return

}

func (self *Client) GetInstrumentByTicker(ivTicker string) (rsInstrument Instrument, roError error) {

	if ivTicker == "" {
		return
	}

	loParams := url.Values{}

	loParams.Add("ticker", ivTicker)

	lvUrl := self.mvUrl + "/market/search/by-ticker?" + loParams.Encode()

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message     string `json:"message"`
			Code        string `json:"code"`
			Total       int    `json:"total"`
			Instruments []struct {
				Figi              string  `json:"figi"`
				Ticker            string  `json:"ticker"`
				Isin              string  `json:"isin"`
				MinPriceIncrement float64 `json:"minPriceIncrement"`
				Lot               int     `json:"lot"`
				MinQuantity       int     `json:"minQuantity"`
				Currency          string  `json:"currency"`
				Name              string  `json:"name"`
				Type              string  `json:"type"`
			} `json:"instruments"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	//log.Println(string(lvBodyBytes))

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseInstrument := range lsResponse.Payload.Instruments {

		rsInstrument.Type = lsResponseInstrument.Type
		rsInstrument.FIGI = lsResponseInstrument.Figi
		rsInstrument.Ticker = lsResponseInstrument.Ticker
		rsInstrument.Text = lsResponseInstrument.Name
		rsInstrument.Currency = lsResponseInstrument.Currency
		rsInstrument.Lot = lsResponseInstrument.Lot
		rsInstrument.MinPriceIncrement = lsResponseInstrument.MinPriceIncrement

		break

	}

	return

}

func (self *Client) GetInstrumentByFIGI(ivFIGI string) (rsInstrument Instrument, roError error) {

	loParams := url.Values{}

	loParams.Add("figi", ivFIGI)

	lvUrl := self.mvUrl + "/market/search/by-figi?" + loParams.Encode()

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message           string  `json:"message"`
			Code              string  `json:"code"`
			Figi              string  `json:"figi"`
			Ticker            string  `json:"ticker"`
			Isin              string  `json:"isin"`
			MinPriceIncrement float64 `json:"minPriceIncrement"`
			Lot               int     `json:"lot"`
			Currency          string  `json:"currency"`
			Name              string  `json:"name"`
			Type              string  `json:"type"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	//log.Println(string(lvBodyBytes))

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	rsInstrument.Type = lsResponse.Payload.Type
	rsInstrument.FIGI = lsResponse.Payload.Figi
	rsInstrument.Ticker = lsResponse.Payload.Ticker
	rsInstrument.Text = lsResponse.Payload.Name
	rsInstrument.Currency = lsResponse.Payload.Currency
	rsInstrument.Lot = lsResponse.Payload.Lot
	rsInstrument.MinPriceIncrement = lsResponse.Payload.MinPriceIncrement

	return

}

func (self *Client) GetCandles(ivTicker string, ivInterval string, ivFrom time.Time, ivTo time.Time) (rtCandles []Candle, roError error) {

	loInstrument, roError := self.GetInstrumentByTicker(ivTicker)

	if roError != nil {
		return
	}

	loParams := url.Values{}

	loParams.Add("figi", loInstrument.FIGI)
	loParams.Add("interval", ivInterval)
	loParams.Add("from", ivFrom.Format(time.RFC3339))
	loParams.Add("to", ivTo.Format(time.RFC3339))

	lvUrl := self.mvUrl + "market/candles?" + loParams.Encode()

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message  string `json:"message"`
			Code     string `json:"code"`
			Figi     string `json:"figi"`
			Interval string `json:"interval"`
			Candles  []struct {
				Figi     string    `json:"figi"`
				Interval string    `json:"interval"`
				O        float64   `json:"o"`
				C        float64   `json:"c"`
				H        float64   `json:"h"`
				L        float64   `json:"l"`
				V        float64   `json:"v"`
				Time     time.Time `json:"time"`
			} `json:"candles"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}


	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseCandle := range lsResponse.Payload.Candles {

		lsCandle := Candle{}

		lsCandle.Time = lsResponseCandle.Time
		lsCandle.High = lsResponseCandle.H
		lsCandle.Open = lsResponseCandle.O
		lsCandle.Close = lsResponseCandle.C
		lsCandle.Low = lsResponseCandle.L
		lsCandle.Volume = lsResponseCandle.V

		lsCandle.eval()

		rtCandles = append(rtCandles, lsCandle)

	}

	return

}

func (self *Client) GetPositions() (rtPositions []Position, roError error) {

	lvUrl := self.mvUrl + "portfolio"

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message   string `json:"message"`
			Code      string `json:"code"`
			Positions []struct {
				Figi           string  `json:"figi"`
				Ticker         string  `json:"ticker"`
				Isin           string  `json:"isin"`
				InstrumentType string  `json:"instrumentType"`
				Balance        float64 `json:"balance"`
				Blocked        int     `json:"blocked"`
				ExpectedYield  struct {
					Currency string  `json:"currency"`
					Value    float64 `json:"value"`
				} `json:"expectedYield"`
				Lots                 int `json:"lots"`
				AveragePositionPrice struct {
					Currency string  `json:"currency"`
					Value    float64 `json:"value"`
				} `json:"averagePositionPrice"`
				AveragePositionPriceNoNkd struct {
					Currency string  `json:"currency"`
					Value    float64 `json:"value"`
				} `json:"averagePositionPriceNoNkd"`
				Name string `json:"name"`
			} `json:"positions"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	//log.Println(string(lvBodyBytes))

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponsePosition := range lsResponse.Payload.Positions {

		lsPosition := Position{}

		lsPosition.FIGI = lsResponsePosition.Figi
		lsPosition.Ticker = lsResponsePosition.Ticker
		lsPosition.Type = lsResponsePosition.InstrumentType
		lsPosition.Text = lsResponsePosition.Name
		lsPosition.Quantity = lsResponsePosition.Balance
		lsPosition.Lots = lsResponsePosition.Lots
		lsPosition.Currency = lsResponsePosition.AveragePositionPrice.Currency
		lsPosition.Price = lsResponsePosition.AveragePositionPrice.Value
		lsPosition.Profit = lsResponsePosition.ExpectedYield.Value

		rtPositions = append(rtPositions, lsPosition)

	}

	return

}

func (self *Client) GetOperations(ivTicker string, ivFrom time.Time, ivTo time.Time) (rtOperations []Operation, roError error) {

	loInstrument := Instrument{}

	loParams := url.Values{}

	loParams.Add("from", ivFrom.Format(time.RFC3339))
	loParams.Add("to", ivTo.Format(time.RFC3339))

	if ivTicker != "" {

		loInstrument, roError = self.GetInstrumentByTicker(ivTicker)

		if roError != nil {
			return
		}

		if loInstrument.FIGI == FigiTCSG {
			loInstrument.FIGI = FigiTCS
		}

		loParams.Add("figi", loInstrument.FIGI)

	}

	lvUrl := self.mvUrl + "operations?" + loParams.Encode()

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message    string `json:"message"`
			Code       string `json:"code"`
			Operations []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Trades []struct {
					TradeID  string    `json:"tradeId"`
					Date     time.Time `json:"date"`
					Price    float64   `json:"price"`
					Quantity float64   `json:"quantity"`
				} `json:"trades"`
				Commission struct {
					Currency string  `json:"currency"`
					Value    float64 `json:"value"`
				} `json:"commission"`
				Currency         string    `json:"currency"`
				Payment          float64   `json:"payment"`
				Price            float64   `json:"price"`
				Quantity         float64   `json:"quantity"`
				QuantityExecuted float64   `json:"quantityExecuted"`
				Figi             string    `json:"figi"`
				InstrumentType   string    `json:"instrumentType"`
				IsMarginCall     bool      `json:"isMarginCall"`
				Date             time.Time `json:"date"`
				OperationType    string    `json:"operationType"`
			} `json:"operations"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseOperation := range lsResponse.Payload.Operations {

		if lsResponseOperation.Figi == FigiTCS &&
			lsResponseOperation.Currency == CurrencyRUB {
			lsResponseOperation.Figi = FigiTCSG
		}

		if ivTicker != "" {

			if ivTicker == TickerTCS &&
				lsResponseOperation.Figi != FigiTCS {
				continue
			}

			if ivTicker == TickerTCSG &&
				lsResponseOperation.Figi != FigiTCSG {
				continue
			}

		}

		if lsResponseOperation.OperationType != "Buy" &&
			lsResponseOperation.OperationType != "BuyCard" &&
			lsResponseOperation.OperationType != "Sell" &&
			lsResponseOperation.OperationType != "Dividend" &&
			lsResponseOperation.OperationType != "TaxDividend" &&
			lsResponseOperation.OperationType != "Coupon" {
			continue
		}

		if lsResponseOperation.Status != "Done" {
			continue
		}

		if lsResponseOperation.OperationType == "BuyCard" {
			lsResponseOperation.OperationType = "Buy"
		}

		lsOperation := Operation{}

		lsOperation.ID = lsResponseOperation.ID
		lsOperation.Type = lsResponseOperation.OperationType
		lsOperation.FIGI = lsResponseOperation.Figi
		lsOperation.Currency = lsResponseOperation.Currency
		lsOperation.Date = lsResponseOperation.Date
		lsOperation.Quantity = lsResponseOperation.QuantityExecuted
		lsOperation.Price = math.Abs(lsResponseOperation.Price)
		lsOperation.Value = math.Abs(lsResponseOperation.Payment)
		lsOperation.Commission = math.Abs(lsResponseOperation.Commission.Value)

		rtOperations = append(rtOperations, lsOperation)

	}

	sort.Slice(rtOperations, func(i, j int) bool {
		return rtOperations[i].Date.Before(rtOperations[j].Date)
	})

	return

}

func (self *Client) GetOrders() (rtOrders []Order, roError error) {

	lvUrl := self.mvUrl + "orders"

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponseGeneric struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
	}

	type ltsResponseError struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"payload"`
	}

	type ltsResponseResult struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    []struct {
			OrderID       string  `json:"orderId"`
			Figi          string  `json:"figi"`
			Operation     string  `json:"operation"`
			Status        string  `json:"status"`
			RequestedLots int     `json:"requestedLots"`
			ExecutedLots  int     `json:"executedLots"`
			Type          string  `json:"type"`
			Price         float64 `json:"price"`
		} `json:"payload"`
	}

	lsResponseGeneric := ltsResponseGeneric{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponseGeneric)

	if roError != nil {
		return
	}

	if lsResponseGeneric.Status == statusError {

		lsResponseError := ltsResponseError{}

		roError = json.Unmarshal(lvBodyBytes, &lsResponseError)

		if roError != nil {
			return
		}

		roError = errors.New(lsResponseError.Payload.Message)

		return

	}

	lsResponseResult := ltsResponseResult{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponseResult)

	if roError != nil {
		return
	}

	for _, lsResponseOrder := range lsResponseResult.Payload {

		lsOrder := Order{}

		lsOrder.ID = lsResponseOrder.OrderID
		lsOrder.FIGI = lsResponseOrder.Figi
		lsOrder.Type = lsResponseOrder.Type
		lsOrder.Operation = lsResponseOrder.Operation
		lsOrder.Status = lsResponseOrder.Status
		lsOrder.Price = lsResponseOrder.Price
		lsOrder.RequestedLots = lsResponseOrder.RequestedLots
		lsOrder.ExecutedLots = lsResponseOrder.ExecutedLots

		rtOrders = append(rtOrders, lsOrder)

	}

	return

}

func (self *Client) CreateLimitOrder(ivTicker string, ivOperation string, ivLots int, ivPrice float64) (rvOrderID string, roError error) {

	rvOrderID, roError = self.createOrder(orderTypeLimit, ivTicker, ivOperation, ivLots, ivPrice)

	return

}

func (self *Client) CreateMarketOrder(ivTicker string, ivOperation string, ivLots int) (rvOrderID string, roError error) {

	rvOrderID, roError = self.createOrder(orderTypeMarket, ivTicker, ivOperation, ivLots, 0)

	return

}

func (self *Client) createOrder(ivType string, ivTicker string, ivOperation string, ivLots int, ivPrice float64) (rvOrderID string, roError error) {

	loInstrument, roError := self.GetInstrumentByTicker(ivTicker)

	if roError != nil {
		return
	}

	loParams := url.Values{}

	loParams.Add("figi", loInstrument.FIGI)

	lvUrl := self.mvUrl + "orders/" + ivType + "-order?" + loParams.Encode()

	type ltsBody struct {
		Operation string  `json:"operation"`
		Lots      int     `json:"lots"`
		Price     float64 `json:"price"`
	}

	lsBody := ltsBody{}

	lsBody.Operation = ivOperation
	lsBody.Lots = ivLots
	lsBody.Price = ivPrice

	lvBodyJsonBytes, roError := json.Marshal(lsBody)

	if roError != nil {
		return
	}

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodPost, lvUrl, bytes.NewBuffer(lvBodyJsonBytes))

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			OrderID       string `json:"orderId"`
			Operation     string `json:"operation"`
			Status        string `json:"status"`
			RejectReason  string `json:"rejectReason"`
			Message       string `json:"message"`
			RequestedLots int    `json:"requestedLots"`
			ExecutedLots  int    `json:"executedLots"`
			Commission    struct {
				Currency string `json:"currency"`
				Value    int    `json:"value"`
			} `json:"commission"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	rvOrderID = lsResponse.Payload.OrderID

	return

}

func (self *Client) CancelOrder(ivOrderID string) (roError error) {

	loParams := url.Values{}

	loParams.Add("orderId", ivOrderID)

	lvUrl := self.mvUrl + "orders/cancel?" + loParams.Encode()

	loClient := http.Client{}

	loRequest, roError := http.NewRequest(http.MethodGet, lvUrl, nil)

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+self.mvToken)

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	lvBodyBytes, roError := io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBodyBytes, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	return

}
