package services

import "math"

type instance struct {
	leverage float64
	rt       float64
	cs       float64
	ct       float64
}

var ftcCalObj *instance

// New ...
func GetFtxCalInstance() *instance {
	if ftcCalObj == nil {
		ftcCalObj = &instance{
			leverage: 1,
			rt:       1,
		}
	}
	return ftcCalObj
}

func (me *instance) SetLastRawAndRebalance(t, p float64) {
	me.ct = t
	me.cs = p
}

// SetLeverage ...
func (me *instance) SetLeverage(l int) {
	me.leverage = float64(l * 1.0)
}

// SetRebalanceThreshold ...
func (me *instance) SetRebalanceThreshold(r float64) {
	me.rt = r
}

// GetLETFPriceOfCheckpoint ...
func (me *instance) GetLETFPriceOfCheckpoint() float64 {
	return me.cs
}

// GetTargetPriceOfCheckpoint ...
func (me *instance) GetTargetPriceOfCheckpoint() float64 {
	return me.ct
}

// FeedPrice ...
func (me *instance) FeedPrice(t float64) float64 {
	me.wrapper(t)
	return me.price(t)
}

// Rebalance ...
func (me *instance) Rebalance(t float64) {
	me.wrapper(t)
	me.rebalance(t)
}

func (me *instance) wrapper(t float64) {
	delta := t - me.ct
	abs := math.Abs(delta)
	threshold := math.Abs(me.ct * me.rt)
	if abs > threshold {
		me.rebalance(me.ct + delta/abs*threshold)
		me.wrapper(t)
	}
}

func (me *instance) rebalance(t float64) {
	me.cs = me.price(t)
	me.ct = t
}

func (me *instance) price(t float64) float64 {
	pt := me.ct + me.leverage*(t-me.ct)
	ps := me.cs * pt / me.ct
	return math.Max(0, ps)
}
