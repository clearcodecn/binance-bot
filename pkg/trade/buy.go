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
	for {
		select {
		case <-ctx.Done():
			return
		case change := <-t.buyChan:
			option := t.Option()
			for _, c := range change {
				t.boughtMutex.Lock()
				count := len(t.boughtInfo)
				t.boughtMutex.Unlock()
				if count >= option.BuyOption.MaxBuy {
					continue
				}
				if order, err := t.buySymbol(ctx, c.symbol, c.lastPrice); err != nil {
					t.logger.WithError(err).Errorf("failed to buy symbol=%s price=%f tp=%f", c.symbol, c.lastPrice, c.change)
				} else {
					price, _ := strconv.ParseFloat(order.CummulativeQuoteQuantity, 64)
					number, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
					var (
						stopLoss      *float64
						forceStopLoss *float64
					)
					if option.SellOption.StopLoss != 0 {
						v := -1 * option.SellOption.StopLoss
						stopLoss = &v
					}
					if option.SellOption.ForceStopLoss != 0 {
						v := -1 * option.SellOption.ForceStopLoss
						forceStopLoss = &v
					}
					info := &BoughtInfo{
						Symbol:                   c.symbol,
						OrderId:                  order.OrderID,
						Time:                     time.Now(),
						StopLoss:                 stopLoss,
						ForceStopLoss:            forceStopLoss,
						TakeProfit:               option.SellOption.TakeProfit,
						Volume:                   order.Number,
						ExecutedQuantity:         number,
						CummulativeQuoteQuantity: price,
						LotSize:                  order.LotSize,
					}
					t.boughtMutex.Lock()
					t.boughtInfo[c.symbol] = info
					t.boughtMutex.Unlock()
				}
			}
			t.save()
		}
	}
}

type Order struct {
	*binance.Order
	Number  float64
	LotSize int
}

// buySymbol buy a symbol
func (t *Trade) buySymbol(ctx context.Context, symbol string, lastPrice float64) (*Order, error) {
	option := t.Option()
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

	//t.BeforeBuy(symbol, number, lastPrice)

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

	if t.AfterBuy != nil {
		go t.AfterBuy(order)
	}
	return &Order{
		Order:   order,
		Number:  number,
		LotSize: lotSize,
	}, nil
}
