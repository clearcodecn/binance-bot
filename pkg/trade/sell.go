package trade

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
	info        *BoughtInfo
	Reason      SellReason
	PriceChange float64
}
