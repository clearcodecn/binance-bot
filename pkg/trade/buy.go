package trade

import (
	"context"
	"math"
	"time"
)

type BoughtInfo struct {
	Symbol                   string    `json:"symbol"`
	OrderId                  int64     `json:"orderId"`
	Time                     time.Time `json:"time"`
	Price                    float64   `json:"price"`
	StopLoss                 *float64  `json:"stopLoss"`
	ForceStopLoss            *float64  `json:"forceStopLoss"`
	TakeProfit               float64   `json:"takeProfit"`
	Volume                   float64   `json:"volume"`
	ExecutedQuantity         float64   `json:"executedQuantity"`
	CummulativeQuoteQuantity float64   `json:"cummulativeQuoteQuantity"`
	LotSize                  int       `json:"lotSize"`
}

func (b *BoughtInfo) GetPrice() float64 {
	if b.ExecutedQuantity != 0 {
		return b.CummulativeQuoteQuantity / b.ExecutedQuantity
	}

	return math.MaxFloat64
}

func (t *Trade) buyCoin(ctx context.Context, symbol string, change float64) error {
}
