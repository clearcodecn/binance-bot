// Copyright By git@clearcode.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trade

import (
	"time"
)

type floatWrapper struct {
	f float64
}

var (
	DefaultBuyOption = BuyOption{
		Interval:            1,
		PriceUpChange:       &(&floatWrapper{f: 1}).f,
		PriceDownChange:     nil,
		MaxBuy:              4,
		MoneyPerOrder:       11,
		MainCoin:            "USDT",
		WhiteList:           whiteList,
		BoughtFile:          "trade.json",
		BuySameCoinDuration: 1 * time.Minute,
	}

	DefaultSellOption = SellOption{
		StopLoss:                 1.5,
		StopLossDuration:         0,
		TakeProfit:               2,
		TrailingTakeProfit:       0,
		TrailingStopLoss:         0,
		EnableTrailingTakeProfit: false,
		Interval:                 1 * time.Second,
		ForceStopLoss:            3.0,
	}
)

type Options func(o *Option)

type Option struct {
	BuyOption    BuyOption
	SystemOption SystemOption
	SellOption   SellOption
}

// BuyOption defines options for buy a coin
type BuyOption struct {
	// Interval we will check price every %s time
	Interval time.Duration

	// PriceChange the percent of priceChange from the interval start and interval end time.
	// eg: start at 100, end at 102, and we set 1.0, we will buy it
	PriceUpChange *float64

	// PriceDownChange the percent of priceChange from the interval start and interval end time.
	// eg: start at 100, end at 98, and we set 1.0, we will buy it
	PriceDownChange *float64

	// MaxBuy max coins we will buy
	MaxBuy int

	// MoneyPerOrder the total money we buy each order
	MoneyPerOrder float64

	// MainCoin is the money unit, like: USDT
	MainCoin string

	// WhiteList we will only buy the coins in the list
	WhiteList []string

	// BoughtFile the file we save once we buy a coin.
	BoughtFile string

	// BuySameCoinDuration that means if we just sell a coin, after how long we can buy it again.
	BuySameCoinDuration time.Duration
}

func (b BuyOption) InWhiteList(symbol string) bool {
	for _, w := range b.WhiteList {
		if symbol+b.MainCoin == w {
			return true
		}
	}
	return false
}

// SellOption defines sell option to sell a coin
type SellOption struct {
	// StopLoss once coin's price reach to the price, we will sell it
	// If we set ForceStopLoss, we will decide to wait for a moment.
	StopLoss float64

	// StopLossDuration  once the price reach to (ForceStopLoss, StopLoss]
	// We will wait, once it reach to the time, we will decide to sell it.
	StopLossDuration time.Duration

	// TakeProfit once coin's price reach to the price, we will sell it
	TakeProfit float64

	EnableTrailingTakeProfit bool

	// TrailingTakeProfit once coin's price reach to TakeProfit,
	// we will increase TakeProfit = TakeProfit + TrailingTakeProfit
	TrailingTakeProfit float64

	// TrailingTakeProfit once coin's price reach to TakeProfit,
	// we will increase StopLoss = StopLoss + TrailingStopLoss
	TrailingStopLoss float64

	// Interval how much time we check price
	Interval time.Duration

	// ForceStopLoss
	ForceStopLoss float64
}

// SystemOption defines the options for system to running
type SystemOption struct {
	LogFile   string
	AccessKey string
	SecretKey string
	Debug     bool

	ProxyURL string
}

func WithSellOption(option SellOption) Options {
	return func(o *Option) {
		o.SellOption = option
	}
}

func WithBuyOption(option BuyOption) Options {
	return func(o *Option) {
		o.BuyOption = option
	}
}

func WithDefaultBuyOption() Options {
	return func(o *Option) {
		o.BuyOption = DefaultBuyOption
	}
}

func WithDefaultSellOption() Options {
	return func(o *Option) {
		o.SellOption = DefaultSellOption
	}
}
