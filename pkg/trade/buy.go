package trade

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"math"
	"strconv"
	"strings"
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

func (t *Trade) runBuy(ctx context.Context) {
	for change := range t.buyChan {
		for _, c := range change {
			if order, err := t.buySymbol(ctx, c.symbol, c.lastPrice); err != nil {
				t.logger.WithError(err).Errorf("failed to buy symbol=%s price=%f tp=%f", c.symbol, c.lastPrice, c.change)
			} else {
				t.boughtMutex.Lock()

				executedQuantity, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
				cummulativeQuoteQuantity, _ := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)

				//cummulativeQuoteQuantity / executedQuantity

				t.boughtInfo[c.symbol] = &BoughtInfo{
					Symbol:                   c.symbol,
					OrderId:                  order.OrderID,
					Time:                     time.Now(),
					Price:                    order.Price,
					StopLoss:                 nil,
					ForceStopLoss:            nil,
					TakeProfit:               0,
					Volume:                   0,
					ExecutedQuantity:         executedQuantity,
					CummulativeQuoteQuantity: cummulativeQuoteQuantity,
					LotSize:                  0,
				}
				t.boughtMutex.Unlock()
			}
		}
	}
}

// buySymbol buy a symbol
func (t *Trade) buySymbol(ctx context.Context, symbol string, lastPrice float64) (*binance.Order, error) {
	t.mu.Lock()
	option := t.option
	t.mu.Unlock()

	info, err := t.GetExchangeInfo(ctx, symbol)
	if err != nil {
		return nil, err
	}

	var symbolInfo binance.Symbol
	for _, s := range info.Symbols {
		if s.Symbol == symbol {
			symbolInfo = s
			break
		}
	}

	if symbolInfo.Symbol == "" {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	if len(symbolInfo.Filters) < 3 {
		return nil, fmt.Errorf("not found stepSize in symbol: %s", symbol)
	}

	stepSize, ok := symbolInfo.Filters[2]["stepSize"]
	if !ok {
		return nil, fmt.Errorf("not found stepSize in symbol: %s", symbol)
	}
	stepSizeString, ok := stepSize.(string)
	if !ok {
		return nil, fmt.Errorf("not found stepSize in symbol: %s", symbol)
	}
	lotSize := strings.Index(stepSizeString, "1") - 1
	if lotSize < 0 {
		lotSize = 0
	}

	// calculate number:  use total money / current price.
	number := option.BuyOption.MoneyPerOrder / lastPrice

	if lotSize == 0 {
		number = float64(int(number))
	} else {
		number = FloatTrunc(number, lotSize)
	}

	t.BeforeBuy(symbol, number, lastPrice)

	// buy.
	resp, err := t.Buy(ctx, number, symbol)
	if err != nil {
		return nil, err
	}

	// query order.
	order, err := t.GetOrder(ctx, symbol, resp.OrderID, math.MaxInt64)
	if err != nil {
		return nil, err
	}

	t.AfterBuy(order)

	return order, nil
}
