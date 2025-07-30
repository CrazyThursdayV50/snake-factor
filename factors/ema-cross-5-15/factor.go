package ec5_15

import (
	"github.com/CrazyThursdayV50/snake-factor/models"
	"github.com/shopspring/decimal"
)

type Factor struct {
	xWeight decimal.Decimal
	kWeight decimal.Decimal

	previousEMA5  decimal.Decimal
	previousEMA15 decimal.Decimal

	currentEMA5  decimal.Decimal
	currentEMA15 decimal.Decimal

	currentKline *models.Kline
}

// SetXWeight 设置 金x/死x 权重
// weight 范围：[0,1]
// 对应的，交叉后的蜡烛收盘权重为 1-weight
func (f *Factor) SetXWeight(weight decimal.Decimal) *Factor {
	f.xWeight = weight
	f.kWeight = decimal.NewFromInt(1).Sub(weight)
	return f
}

// SetPreviousEMA 设置前一个 EMA 值
func (f *Factor) SetPreviousEMA(ema5, ema15 decimal.Decimal) *Factor {
	f.previousEMA5, f.previousEMA15 = ema5, ema15
	return f
}

// SetCurrentEMA 设置前一个 EMA 值
func (f *Factor) SetCurrentEMA(ema5, ema15 decimal.Decimal) *Factor {
	f.currentEMA5, f.currentEMA15 = ema5, ema15
	return f
}

// SetCurrentKline 设置当前 Kline
func (f *Factor) SetCurrentKline(kline *models.Kline) *Factor {
	f.currentKline = kline
	return f
}

func New() *Factor {
	return &Factor{
		xWeight: decimal.NewFromFloat(0.5),
		kWeight: decimal.NewFromFloat(0.5),
	}
}

func (f *Factor) Factor() float64 {

}
