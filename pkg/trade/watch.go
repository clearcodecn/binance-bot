package trade

import (
	"context"
	"sort"
	"time"
)

func (t *Trade) watchPrice(ctx context.Context) {
	var (
		option          = t.Option()
		sellCheckTicker = time.NewTicker(option.SellOption.Interval)
		buyCheckTicker  = time.NewTicker(option.BuyOption.Interval)
		lastPrice       map[string]*SymbolPrice
	)
	for {
		select {
		case <-ctx.Done():
			return
		case <-buyCheckTicker.C:
			t.mu.Lock()
			option := t.option
			t.mu.Unlock()
			nowPrice, change := t.checkingPrice(ctx, option, lastPrice)
			if nowPrice != nil {
				lastPrice = nowPrice
			}
			if change != nil {
				t.buyChan <- change
			}
		case <-sellCheckTicker.C:
			t.mu.Lock()
			option := t.option
			t.mu.Unlock()

			if err := t.checkingTPSL(ctx, option); err != nil {
				t.logger.WithError(err).Error("failed to checking TP/SL")
			}
		}
	}
}

type symbolPriceChange struct {
	symbol    string
	lastPrice float64
	nowPrice  float64
	change    float64
	volume    float64
}

func (t *Trade) checkingPrice(ctx context.Context, option Option, lastPrice map[string]*SymbolPrice) (map[string]*SymbolPrice, []*symbolPriceChange) {
	var changes []*symbolPriceChange
	nowPrice := t.GetSymbolPrice(ctx, "")
	for symbol, pr := range lastPrice {
		// check if symbol is in white list
		if !option.BuyOption.InWhiteList(symbol) {
			continue
		}
		// check if symbol in buy blocks
		if t.isBlock(symbol) {
			continue
		}

		now, ok := nowPrice[symbol]
		if !ok {
			continue
		}
		var shouldBuy bool
		change := (now.Price - pr.Price) / pr.Price * 100
		if option.BuyOption.PriceUpChange != nil {
			up := *option.BuyOption.PriceUpChange
			if change > up {
				shouldBuy = true
			}
		}
		if shouldBuy {
			changes = append(changes, &symbolPriceChange{
				symbol:    symbol,
				lastPrice: pr.Price,
				nowPrice:  now.Price,
				change:    change,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].change > changes[j].change
	})

	return nowPrice, changes
}

func (t *Trade) checkingTPSL(ctx context.Context, option Option) error {
	symbolPrice := t.GetSymbolPrice(ctx, "")
	bought := t.getBoughtInfo()

	for coin, info := range bought {
		sp, ok := symbolPrice[coin]
		if !ok {
			continue
		}

		var (
			price              = info.GetPrice()
			takeProfitPrice    = price * (1 + info.TakeProfit)
			stopLossPrice      = -1.0
			forceStopLossPrice = -1.0
			lastPrice          = sp.Price
			priceChange        = (lastPrice - price) / price * 100 // tp/sl
			shouldSell         bool
			sellReason         SellReason
		)

		if info.StopLoss != nil {
			stopLossPrice = price * (1 + *info.StopLoss)
		}

		if info.ForceStopLoss != nil {
			forceStopLossPrice = price * (1 + *info.ForceStopLoss)
		}

		// 止盈点.
		if lastPrice >= takeProfitPrice {
			// 持续止盈 止损
			if option.SellOption.EnableTrailingTakeProfit {
				info.TakeProfit += option.SellOption.TrailingTakeProfit
				t.OnTrailingTakeProfit(info)
				if info.StopLoss != nil {
					sl := info.TakeProfit - option.SellOption.TrailingStopLoss + *info.StopLoss
					info.StopLoss = &sl
				}
				continue
			}
			shouldSell = true
			sellReason = SellReasonForTakeProfit
		}

		if lastPrice <= forceStopLossPrice {
			shouldSell = true
			sellReason = SellReasonForForceStopLoss
		} else {
			if lastPrice < stopLossPrice {
				if info.Time.Add(option.SellOption.StopLossDuration).After(time.Now()) {
					shouldSell = true
					sellReason = SellReasonForStopLoss
				}
			}
		}

		if !shouldSell {
			continue
		}

		sellInfo := &SellBill{
			Info:        info,
			Reason:      sellReason,
			PriceChange: priceChange,
		}

		select {
		case <-ctx.Done():
			return nil
		case t.sellChan <- sellInfo:
		}
	}

	return nil
}
