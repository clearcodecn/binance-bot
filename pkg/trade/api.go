package trade

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"strconv"
	"time"
)

// SymbolPrice defines symbol price with time
type SymbolPrice struct {
	Symbol string
	Price  float64
	Time   time.Time
}

// GetSymbolPrice get the symbol price
func (t *Trade) GetSymbolPrice(ctx context.Context, symbol string) map[string]*SymbolPrice {
	var prs = make(map[string]*SymbolPrice)
	svc := t.client.NewListPricesService()
	if symbol != "" {
		svc.Symbol(symbol)
	}
	res, err := svc.Do(ctx)
	if err != nil {
		return prs
	}
	for _, sp := range res {
		price, _ := strconv.ParseFloat(sp.Price, 64)
		pr := &SymbolPrice{
			Symbol: sp.Symbol,
			Price:  price,
			Time:   time.Now(),
		}
		prs[sp.Symbol] = pr
	}
	return prs
}

// Buy buy coin with market price with given number
func (t *Trade) Buy(ctx context.Context, number float64, coin string) (*binance.CreateOrderResponse, error) {
	order, err := t.client.NewCreateOrderService().
		Quantity(strconv.FormatFloat(number, 'g', -1, 64)).
		Symbol(coin).
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeMarket).
		Do(ctx, binance.WithRecvWindow(50000))
	if err != nil {
		return nil, err
	}
	return order, nil
}

// Sell sell coin with market price with given number
func (t *Trade) Sell(ctx context.Context, symbol string, number float64) (*binance.CreateOrderResponse, error) {
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeMarket).
		Quantity(strconv.FormatFloat(number, 'g', -1, 64)).
		Do(ctx, binance.WithRecvWindow(50000))
	if err != nil {
		return nil, err
	}
	return order, nil
}

const DefaultRetry = 3

// GetOrder get the order with given id
func (t *Trade) GetOrder(ctx context.Context, symbol string, id int64, retry int) (*binance.Order, error) {
	var (
		order *binance.Order
		err   error
	)
	if retry <= 0 {
		retry = DefaultRetry
	}
	for i := 0; i < retry; i++ {
		order, err = t.client.NewGetOrderService().
			Symbol(symbol).
			OrderID(id).
			Do(ctx, binance.WithRecvWindow(50000))
		if err != nil {
			continue
		}
		return order, nil
	}
	return nil, err
}

func (t *Trade) GetExchangeInfo() {
	t.client.NewExchangeInfoService().Do()
}