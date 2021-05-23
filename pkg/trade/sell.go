package trade

import (
	"context"
	"time"
)

type SellReason int

const (
	SellReasonUnknown          SellReason = 0
	SellReasonForTakeProfit    SellReason = 1
	SellReasonForStopLoss      SellReason = 2
	SellReasonForForceStopLoss SellReason = 3
)

func (s SellReason) String() string {
	switch s {
	case SellReasonForTakeProfit:
		return "order already reach to take profit price/订单已达到止盈点"
	case SellReasonForStopLoss:
		return "order already reach to stop loss price/订单已达到止损点"
	case SellReasonForForceStopLoss:
		return "order already reach to force stop loss price/订单已达到强制止损点"
	}
	return "unknown"
}

type SellBill struct {
	Info        *BoughtInfo
	Reason      SellReason
	PriceChange float64
}

func (t *Trade) isBlock(symbol string) bool {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()

	val, ok := t.boughtCache.Get(symbol)
	if !ok {
		return false
	}
	ti := val.(time.Time)
	if ti.After(time.Now()) {
		return true
	}
	t.boughtCache.Remove(symbol)
	return false
}

func (t *Trade) addBlock(symbol string) {
	option := t.Option()
	if option.BuyOption.SameCoinBlockDuration == 0 {
		return
	}

	expire := time.Now().Add(option.BuyOption.SameCoinBlockDuration)

	t.cacheMutex.Lock()
	t.boughtCache.Add(symbol, expire)
	t.cacheMutex.Unlock()
}

func (t *Trade) runSell(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sellBill := <-t.sellChan:
			number := FloatTrunc(sellBill.Info.Volume*0.999, sellBill.Info.LotSize)
			_, err := t.Sell(ctx, sellBill.Info.Symbol, number)
			if err != nil {
				t.logger.WithError(err).Errorf("failed to symbol=%s win=%f %s", sellBill.Info.Symbol, sellBill.PriceChange, sellBill.Reason.String())
				continue
			}
			if t.AfterSell != nil {
				go t.AfterSell(sellBill)
			}

			t.boughtMutex.Lock()
			delete(t.boughtInfo, sellBill.Info.Symbol)
			t.boughtMutex.Unlock()

			t.addBlock(sellBill.Info.Symbol)
		}
	}
}
