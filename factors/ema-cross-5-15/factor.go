package ec5_15

import (
	"sync"

	"github.com/CrazyThursdayV50/indicators/indicators/ema"
	"github.com/CrazyThursdayV50/indicators/kline"
	"github.com/shopspring/decimal"
)

type Factor struct {
	// x权重
	// 金叉信号权重
	xWeight decimal.Decimal
	// k权重
	// k线信号权重
	kWeight decimal.Decimal

	previousEMA5  decimal.Decimal
	previousEMA15 decimal.Decimal

	currentEMA5  decimal.Decimal
	currentEMA15 decimal.Decimal

	// 当前 k 线数据出现 金叉信号
	isCurrentGoldenX bool
	// 前一次 k 线数据出现 金叉信号
	isPreviousGoldenX bool
	// 金叉强度
	goldenXPower decimal.NullDecimal

	currentKline *Kline
	klinePower   decimal.NullDecimal

	ema5  *ema.EMA
	ema15 *ema.EMA
}

type Kline = kline.Model

// SetXWeight 设置 金x/死x 权重
// weight 范围：[0,1]
// 对应的，交叉后的蜡烛收盘权重为 1-weight
func (f *Factor) SetXWeight(weight decimal.Decimal) *Factor {
	f.xWeight = weight
	f.kWeight = decimal.NewFromInt(1).Sub(weight)
	return f
}

// 金叉强度
// speed = diff_current / (diff_current + diff_previous)
// diff_current = (ema5 - ema15)current
// diff_previous = (ema15 - ema5)previous
// 范围：(0,1)
// 越靠近1，越强，中间值：0.5
func (f *Factor) calculateGoldenXPower() decimal.Decimal {
	diffPre := f.previousEMA15.Sub(f.previousEMA5)
	diffCur := f.currentEMA5.Sub(f.currentEMA15)
	return diffCur.Div(diffCur.Add(diffPre))
}

// 使用当前 ema 指标计算是否金叉信号，输出值
func (f *Factor) calcualteCurrentGoldenX() {
	// 如果是 金叉
	if f.previousEMA15.GreaterThan(f.previousEMA5) && f.currentEMA5.GreaterThan(f.currentEMA15) {
		f.isCurrentGoldenX = true
		f.goldenXPower.Decimal = f.calculateGoldenXPower()
		f.goldenXPower.Valid = true
	}
}

func (f *Factor) updateGoldenX(isNextKline bool) {
	// 如果输入的是新时段的k线
	// 此时 Factor 里面的数据还是基于当前k线与前一次k线计算出来的指标
	if isNextKline {
		f.isPreviousGoldenX = f.isCurrentGoldenX

		// 如果当前是金叉
		if f.isCurrentGoldenX {
			// 更新成：上一次是金叉，本次不是金叉
			// 因为不可能连续产生两个金叉
			f.isCurrentGoldenX = false
			f.goldenXPower.Decimal = decimal.Zero
			f.goldenXPower.Valid = false
			return
		}

		// 如果当前不是金叉，那么计算金叉
		f.calcualteCurrentGoldenX()
		return
	}

	// 如果输入的是当前时段k线
	// 那么只需要重新计算即可
	// 如果当前已经是金叉了，那么只更新金叉强度
	if f.isCurrentGoldenX {
		f.goldenXPower.Decimal = f.calculateGoldenXPower()
		return
	}

	// 否则，计算是否金叉
	f.calcualteCurrentGoldenX()
}

func (f *Factor) updateValues() {
	f.previousEMA5 = f.ema5.LastValue()
	f.previousEMA15 = f.ema15.LastValue()
	f.currentEMA5 = f.ema5.CurrentValue()
	f.currentEMA15 = f.ema15.CurrentValue()
}

// k线收阳指标
// (0,1)
// 越靠近1，收阳越强
// 上涨 1% -> 0.009901
// 上涨 2% -> 0.01961
// 上涨 3% -> 0.02913
// 上涨 5% -> 0.04762 (单k线上涨 5% 就已经很厉害了)
// 上涨 10% -> 0.0909
// 上涨 20% -> 0.167
// 上涨 50% -> 0.33
// 上涨 100% -> 0.5
// 上涨 200% -> 0.66
// 上涨 300% -> 0.75
func (f *Factor) updateKlinePower() {
	f.klinePower.Valid = f.currentKline.Close.GreaterThan(f.currentKline.Open)
	if f.klinePower.Valid {
		f.klinePower.Decimal = f.currentKline.Close.Sub(f.currentKline.Open).Div(f.currentKline.Close)
	} else if !f.klinePower.Decimal.IsZero() {
		f.klinePower.Decimal = decimal.Zero
	}
}

func (f *Factor) initWithKlines(klines []*Kline) bool {
	if len(klines) < 15 {
		return false
	}

	ema5 := ema.New(5).Klines(klines[:5]).Build()
	ema15 := ema.New(15).Klines(klines[:15]).Build()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for _, kline := range klines[5:] {
			ema5, _ = ema5.Next(kline)
		}
	}()
	go func() {
		defer wg.Done()
		for _, kline := range klines[15:] {
			ema15, _ = ema15.Next(kline)
		}
	}()
	wg.Wait()

	f.ema5 = ema5
	f.ema15 = ema15
	f.currentKline = klines[len(klines)-1]
	f.updateValues()
	f.updateGoldenX(true)
	return true
}

func (f *Factor) Next(kline *Kline) bool {
	var ok bool
	f.ema5, _ = f.ema5.Next(kline)
	f.ema15, ok = f.ema15.Next(kline)
	f.currentKline = kline
	f.updateValues()
	f.updateGoldenX(ok)
	return ok
}

func New(klines []*Kline) *Factor {
	var f = Factor{
		xWeight: decimal.NewFromFloat(0.7),
		kWeight: decimal.NewFromFloat(0.3),
	}

	ok := f.initWithKlines(klines)
	if !ok {
		return nil
	}
	return &f
}

// 1. 当前收阳
// 2. 已经突破金叉，也即前一次为金叉
func (f *Factor) Value() decimal.Decimal {
	if f.goldenXPower.Valid && f.klinePower.Valid {
		return f.xWeight.Mul(f.goldenXPower.Decimal).Add(f.kWeight.Mul(f.klinePower.Decimal))
	}

	return decimal.Zero
}
