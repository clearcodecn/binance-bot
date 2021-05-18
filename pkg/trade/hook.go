package trade

import "github.com/adshao/go-binance/v2"

func (t *Trade) OnTrailingTakeProfit(info *BoughtInfo) {

}

func (t *Trade) BeforeSell(info *BoughtInfo) {

}

func (t *Trade) BeforeBuy(symbol string, number float64, price float64) {

}

func (t *Trade) AfterBuy(order *binance.Order) {

}
