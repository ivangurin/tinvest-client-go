package tinvestclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const (
	CurrencyRUB           = "RUB"
	CurrencyUSD           = "USD"
	CurrencyEUR           = "EUR"
	IntervalMin1          = "1min"
	IntervalMin2          = "2min"
	IntervalMin3          = "3min"
	IntervalMin5          = "5min"
	IntervalMin10         = "10min"
	IntervalMin15         = "15min"
	IntervalMin30         = "30min"
	IntervalHour          = "hour"
	IntervalDay           = "day"
	IntervalWeek          = "week"
	IntervalMonth         = "month"
	InstumentTypeCurrency = "Currency"
	InstumentTypeShare    = "Stock"
	InstumentTypeBond     = "Bond"
	InstumentTypeETF      = "Etf"
	CandleTypeGreen       = "Green"
	CandleTypeRed         = "Red"
	TickerTCS             = "TCS"
	TickerTCSG            = "TCSG"
	FigiAAPL              = "BBG000B9XRY4"
	FigiTCS               = "BBG005DXJS36"
	FigiTCSG              = "BBG00QPYJ5H0"
	OperationBuy          = "Buy"
	OperationSell         = "Sell"
	OperationBuyCard      = "BuyCard"
	OperationDividend     = "Dividend"
	OperationTaxDividend  = "TaxDividend"
	OperationCoupon       = "Coupon"
	OperationTaxCoupon    = "TaxCoupon"
	statusError           = "Error"
	statusDone            = "Done"
	orderTypeLimit        = "limit"
	orderTypeMarket       = "market"
)

type Client struct {
	mvUrl     string
	mvToken   string
	mvAccount string
}

type Account struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Instrument struct {
	Type              string  `json:"type"`
	Ticker            string  `json:"ticker"`
	FIGI              string  `json:"figi"`
	ISIN              string  `json:"isin"`
	Text              string  `json:"text"`
	Currency          string  `json:"currency"`
	Lot               int     `json:"lot"`
	MinPriceIncrement float64 `json:"minPriceIncrement"`
}

type Candle struct {
	Time       time.Time `json:"time"`
	High       float64   `json:"high"`
	Open       float64   `json:"open"`
	Close      float64   `json:"close"`
	Low        float64   `json:"low"`
	Volume     float64   `json:"volume"`
	ShadowHigh float64   `json:"shadowHigh"`
	ShadowLow  float64   `json:"shadowLow"`
	Body       float64   `json:"body"`
	Type       string    `json:"type"`
}

type Position struct {
	FIGI     string  `json:"figi"`
	Ticker   string  `json:"ticker"`
	Type     string  `json:"type"`
	Text     string  `json:"text"`
	Quantity float64 `json:"quantity"`
	Blocked  float64 `json:"blocked"`
	Lots     int     `json:"lots"`
	Currency string  `json:"currency"`
	Price    float64 `json:"price"`
	Profit   float64 `json:"profit"`
}

type Operation struct {
	ID         string    `json:"id"`
	Time       time.Time `json:"time"`
	Type       string    `json:"type"`
	FIGI       string    `json:"figi"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	Value      float64   `json:"value"`
	Commission float64   `json:"commission"`
	Currency   string    `json:"currency"`
}

type Order struct {
	ID            string  `json:"id"`
	FIGI          string  `json:"figi"`
	Type          string  `json:"type"`
	Operation     string  `json:"operation"`
	Price         float64 `json:"price"`
	Status        string  `json:"status"`
	RequestedLots int     `json:"requestedLots"`
	ExecutedLots  int     `json:"executedLots"`
}

func (c *Client) Init(token string) {

	c.mvUrl = "https://api-invest.tinkoff.ru/openapi/"
	c.mvToken = token

}

func (c *Client) setAccount(ivId string) {

	c.mvAccount = ivId

}

func (c *Client) GetAccounts() (rtAccounts []Account, roError error) {

	lvBody, roError := c.httpRequest(http.MethodGet, "user/accounts", nil, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
			Accounts []struct {
				BrokerAccountType string `json:"brokerAccountType"`
				BrokerAccountID   string `json:"brokerAccountId"`
			} `json:"accounts"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBody, &lsResponse)

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
		lsAccount.Text = lsResponseAccount.BrokerAccountType

		rtAccounts = append(rtAccounts, lsAccount)

	}

	return

}

func (c *Client) GetCurrencies() (rtCurrencies []Instrument, roError error) {

	rtCurrencies, roError = c.getInstruments("currencies")

	return

}

func (c *Client) GetShares() (rtShares []Instrument, roError error) {

	rtShares, roError = c.getInstruments("stocks")

	return

}

func (c *Client) GetBonds() (rtBonds []Instrument, roError error) {

	rtBonds, roError = c.getInstruments("bonds")

	return

}

func (c *Client) GetETFs() (rtETFs []Instrument, roError error) {

	rtETFs, roError = c.getInstruments("etfs")

	return

}

func (c *Client) GetInstruments() (rtInstruments []Instrument, roError error) {

	ltCurrencies, roError := c.GetCurrencies()

	if roError != nil {
		return
	}

	ltShares, roError := c.GetShares()

	if roError != nil {
		return
	}

	ltBonds, roError := c.GetBonds()

	if roError != nil {
		return
	}

	ltETFs, roError := c.GetETFs()

	if roError != nil {
		return
	}

	rtInstruments = append(rtInstruments, ltCurrencies...)
	rtInstruments = append(rtInstruments, ltShares...)
	rtInstruments = append(rtInstruments, ltBonds...)
	rtInstruments = append(rtInstruments, ltETFs...)

	return

}

func (c *Client) getInstruments(ivType string) (rtInstruments []Instrument, roError error) {

	lvBody, roError := c.httpRequest(http.MethodGet, "market/"+ivType, nil, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code        string `json:"code"`
			Message     string `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

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

func (c *Client) GetInstrumentByTicker(ivTicker string) (rsInstrument Instrument, roError error) {

	loParams := url.Values{}

	loParams.Add("ticker", ivTicker)

	lvBody, roError := c.httpRequest(http.MethodGet, "market/search/by-ticker", loParams, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code        string `json:"code"`
			Message     string `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

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

func (c *Client) GetInstrumentByFIGI(ivFIGI string) (rsInstrument Instrument, roError error) {

	loParams := url.Values{}

	loParams.Add("figi", ivFIGI)

	lvBody, roError := c.httpRequest(http.MethodGet, "market/search/by-figi", loParams, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code              string  `json:"code"`
			Message           string  `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

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

func (c *Client) GetCandles(ivFIGI string, ivInterval string, ivFrom time.Time, ivTo time.Time) (rtCandles []Candle, roError error) {

	loParams := url.Values{}

	loParams.Add("figi", ivFIGI)
	loParams.Add("interval", ivInterval)
	loParams.Add("from", ivFrom.Format(time.RFC3339))
	loParams.Add("to", ivTo.Format(time.RFC3339))

	lvBody, roError := c.httpRequest(http.MethodGet, "market/candles", loParams, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

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

		if lsCandle.Open < lsCandle.Close {
			lsCandle.Type = CandleTypeGreen
			lsCandle.ShadowHigh = lsCandle.High - lsCandle.Close
			lsCandle.Body = lsCandle.Close - lsCandle.Open
			lsCandle.ShadowLow = lsCandle.Open - lsCandle.Low
		} else {
			lsCandle.Type = CandleTypeRed
			lsCandle.ShadowHigh = lsCandle.High - lsCandle.Open
			lsCandle.Body = lsCandle.Open - lsCandle.Close
			lsCandle.ShadowLow = lsCandle.Close - lsCandle.Low
		}

		rtCandles = append(rtCandles, lsCandle)

	}

	return

}

func (c *Client) GetPositions() (rtPositions []Position, roError error) {

	lvBody, roError := c.httpRequest(http.MethodGet, "portfolio", nil, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			Positions []struct {
				Figi           string  `json:"figi"`
				Ticker         string  `json:"ticker"`
				Isin           string  `json:"isin"`
				InstrumentType string  `json:"instrumentType"`
				Balance        float64 `json:"balance"`
				Blocked        float64 `json:"blocked"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

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
		lsPosition.Blocked = lsResponsePosition.Blocked
		lsPosition.Lots = lsResponsePosition.Lots
		lsPosition.Currency = lsResponsePosition.AveragePositionPrice.Currency
		lsPosition.Price = lsResponsePosition.AveragePositionPrice.Value
		lsPosition.Profit = lsResponsePosition.ExpectedYield.Value

		rtPositions = append(rtPositions, lsPosition)

	}

	return

}

func (c *Client) GetOperations(ivFIGI string, ivFrom time.Time, ivTo time.Time) (rtOperations []Operation, roError error) {

	loParams := url.Values{}

	loParams.Add("from", ivFrom.Format(time.RFC3339))
	loParams.Add("to", ivTo.Format(time.RFC3339))

	if ivFIGI != "" {

		lvFIGI := ivFIGI

		// Workaround for TCSG share
		if lvFIGI == FigiTCSG {
			lvFIGI = FigiTCS
		}

		loParams.Add("figi", lvFIGI)

	}

	lvBody, roError := c.httpRequest(http.MethodGet, "operations", loParams, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code       string `json:"code"`
			Message    string `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	for _, lsResponseOperation := range lsResponse.Payload.Operations {

		// Workaround for TCS share
		if lsResponseOperation.Figi == FigiTCS &&
			lsResponseOperation.Currency == CurrencyRUB {
			lsResponseOperation.Figi = FigiTCSG
		}

		if ivFIGI != "" {

			// Workaround for TCS share
			if ivFIGI == FigiTCS &&
				lsResponseOperation.Figi != FigiTCS {
				continue
			}

			// Workaround for TCSG share
			if ivFIGI == FigiTCSG &&
				lsResponseOperation.Figi != FigiTCSG {
				continue
			}

		}

		if lsResponseOperation.OperationType != OperationBuy &&
			lsResponseOperation.OperationType != OperationBuyCard &&
			lsResponseOperation.OperationType != OperationSell &&
			lsResponseOperation.OperationType != OperationDividend &&
			lsResponseOperation.OperationType != OperationTaxDividend &&
			lsResponseOperation.OperationType != OperationCoupon &&
			lsResponseOperation.OperationType != OperationTaxCoupon {
			continue
		}

		if lsResponseOperation.Status != statusDone {
			continue
		}

		lsOperation := Operation{}

		lsOperation.ID = lsResponseOperation.ID
		lsOperation.Type = lsResponseOperation.OperationType
		lsOperation.FIGI = lsResponseOperation.Figi
		lsOperation.Currency = lsResponseOperation.Currency
		lsOperation.Time = lsResponseOperation.Date
		lsOperation.Quantity = lsResponseOperation.QuantityExecuted
		lsOperation.Price = math.Abs(lsResponseOperation.Price)
		lsOperation.Value = math.Abs(lsResponseOperation.Payment)
		lsOperation.Commission = math.Abs(lsResponseOperation.Commission.Value)

		rtOperations = append(rtOperations, lsOperation)

	}

	sort.Slice(rtOperations, func(i, j int) bool {
		return rtOperations[i].Time.Before(rtOperations[j].Time)
	})

	return

}

func (c *Client) GetOrders() (rtOrders []Order, roError error) {

	lvBody, roError := c.httpRequest(http.MethodGet, "orders", nil, nil)

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
			Code    string `json:"code"`
			Message string `json:"message"`
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

	roError = json.Unmarshal(lvBody, &lsResponseGeneric)

	if roError != nil {
		return
	}

	if lsResponseGeneric.Status == statusError {

		lsResponseError := ltsResponseError{}

		roError = json.Unmarshal(lvBody, &lsResponseError)

		if roError != nil {
			return
		}

		roError = errors.New(lsResponseError.Payload.Message)

		return

	}

	lsResponseResult := ltsResponseResult{}

	roError = json.Unmarshal(lvBody, &lsResponseResult)

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

func (c *Client) CreateLimitOrder(ivFIGI string, ivOperation string, ivLots int, ivPrice float64) (rvOrderID string, roError error) {

	rvOrderID, roError = c.createOrder(orderTypeLimit, ivFIGI, ivOperation, ivLots, ivPrice)

	return

}

func (c *Client) CreateMarketOrder(ivFIGI string, ivOperation string, ivLots int) (rvOrderID string, roError error) {

	rvOrderID, roError = c.createOrder(orderTypeMarket, ivFIGI, ivOperation, ivLots, 0)

	return

}

func (c *Client) createOrder(ivType string, ivFIGI string, ivOperation string, ivLots int, ivPrice float64) (rvOrderID string, roError error) {

	loParams := url.Values{}

	loParams.Add("figi", ivFIGI)

	type ltsBody struct {
		Operation string  `json:"operation"`
		Lots      int     `json:"lots"`
		Price     float64 `json:"price"`
	}

	lsBody := ltsBody{}

	lsBody.Operation = ivOperation
	lsBody.Lots = ivLots
	lsBody.Price = ivPrice

	lvBody, roError := json.Marshal(lsBody)

	if roError != nil {
		return
	}

	lvBody, roError = c.httpRequest(http.MethodPost, "orders/"+ivType+"-order", loParams, lvBody)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code          string `json:"code"`
			Message       string `json:"message"`
			OrderID       string `json:"orderId"`
			Operation     string `json:"operation"`
			RejectReason  string `json:"rejectReason"`
			RequestedLots int    `json:"requestedLots"`
			ExecutedLots  int    `json:"executedLots"`
			Commission    struct {
				Currency string `json:"currency"`
				Value    int    `json:"value"`
			} `json:"commission"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBody, &lsResponse)

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

func (c *Client) CancelOrder(ivOrderID string) (roError error) {

	loParams := url.Values{}

	loParams.Add("orderId", ivOrderID)

	lvBody, roError := c.httpRequest(http.MethodPost, "orders/cancel", loParams, nil)

	if roError != nil {
		return
	}

	type ltsResponse struct {
		TrackingID string `json:"trackingId"`
		Status     string `json:"status"`
		Payload    struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"payload"`
	}

	lsResponse := ltsResponse{}

	roError = json.Unmarshal(lvBody, &lsResponse)

	if roError != nil {
		return
	}

	if lsResponse.Status == statusError {
		roError = errors.New(lsResponse.Payload.Message)
		return
	}

	return

}

func (c *Client) httpRequest(ivMethod string, ivPath string, ioParams url.Values, ivBody []byte) (rvBody []byte, roError error) {

	lvUrl := c.mvUrl + ivPath

	if c.mvAccount != "" {
		ioParams.Add("brokerAccountId", c.mvAccount)
	}

	if ioParams != nil {
		lvUrl = lvUrl + "?" + ioParams.Encode()
	}

	loRequest, roError := http.NewRequest(ivMethod, lvUrl, bytes.NewBuffer(ivBody))

	if roError != nil {
		return
	}

	loRequest.Header.Add("Authorization", "Bearer "+c.mvToken)

	loClient := http.Client{}

	loResponse, roError := loClient.Do(loRequest)

	if roError != nil {
		return
	}

	defer loResponse.Body.Close()

	rvBody, roError = io.ReadAll(loResponse.Body)

	if roError != nil {
		return
	}

	if loResponse.StatusCode != http.StatusOK {

		lvErrorText := ""

		if len(rvBody) == 0 {

			lvErrorText = fmt.Sprintf("%v (%v)", http.StatusText(loResponse.StatusCode), loResponse.StatusCode)

		} else {

			type ltsResponse struct {
				TrackingID string `json:"trackingId"`
				Status     string `json:"status"`
				Payload    struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"payload"`
			}

			lsResponse := ltsResponse{}

			roError = json.Unmarshal(rvBody, &lsResponse)

			if roError == nil {
				lvErrorText = fmt.Sprintf("%v (%v)", lsResponse.Payload.Message, lsResponse.Payload.Code)
			} else {
				lvErrorText = string(rvBody)
			}

		}

		roError = errors.New(lvErrorText)

		return
	}

	return

}
