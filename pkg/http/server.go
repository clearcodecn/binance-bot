package http

import (
	"bytes"
	"github.com/clearcodecn/binance-bot/pkg/trade"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Server struct {
	engine *gin.Engine
	trade  *trade.Trade
}

func (s *Server) Run(stopChan chan struct{}, addr string) error {
	if err := s.trade.Run(stopChan); err != nil {
		return err
	}
	return s.engine.Run(addr)
}

func NewServer(config string) *Server {
	s := new(Server)
	data, err := ioutil.ReadFile(config)
	if err != nil {
		panic(err)
	}

	var opt = new(trade.Option)
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(opt)
	if err != nil {
		panic(err)
	}

	g := gin.Default()
	g.LoadHTMLGlob("web/*.html")
	g.GET("/", s.Index)
	s.engine = g

	// set durations.
	opt.BuyOption.SameCoinBlockDuration = opt.BuyOption.SameCoinBlockDuration * time.Second
	opt.SellOption.Interval = opt.SellOption.Interval * time.Second
	opt.SellOption.StopLossDuration = opt.SellOption.StopLossDuration * time.Second
	opt.BuyOption.Interval = opt.BuyOption.Interval * time.Second

	t := trade.NewTrade(
		trade.WithBuyOption(opt.BuyOption),
		trade.WithSellOption(opt.SellOption),
		trade.WithSystemOption(opt.SystemOption),
	)

	s.trade = t
	return s
}

func (s *Server) Index(ctx *gin.Context) {
	ctx.HTML(200, "index.html", gin.H{})
}
