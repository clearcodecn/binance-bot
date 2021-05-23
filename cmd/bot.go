package main

import (
	"bytes"
	"flag"
	"github.com/adshao/go-binance/v2"
	"github.com/clearcodecn/binance-bot/pkg/trade"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	config string
)

func init() {
	flag.StringVar(&config, "c", "config.yaml", "config file")
}

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(config)
	if err != nil {
		panic(err)
	}

	var opt = new(trade.Option)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(opt)
	if err != nil {
		panic(err)
	}

	opt.BuyOption.SameCoinBlockDuration = opt.BuyOption.SameCoinBlockDuration * time.Second
	opt.SellOption.Interval = opt.SellOption.Interval * time.Second
	opt.SellOption.StopLossDuration = opt.SellOption.StopLossDuration * time.Second
	opt.BuyOption.Interval = opt.BuyOption.Interval * time.Second

	t := trade.NewTrade(
		trade.WithBuyOption(opt.BuyOption),
		trade.WithSellOption(opt.SellOption),
		trade.WithSystemOption(opt.SystemOption),
	)

	ch := make(chan os.Signal)
	stopCh := make(chan struct{})

	t.AfterBuy = func(order *binance.Order) {
		logrus.Infof("buy %s", order.Symbol)
	}

	t.AfterSell = func(info *trade.SellBill) {
		logrus.Info("sell %s - %s", info.Info.Symbol, info.Reason)
	}

	go func() {
		if err := t.Run(stopCh); err != nil {
			log.Fatal(err)
		}
	}()

	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	close(stopCh)
	time.Sleep(3 * time.Second)
}
