package models

import "github.com/shopspring/decimal"

type Kline struct {
	Timestamp int64
	OpenTime  int64
	CloseTime int64
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
}
