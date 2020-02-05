package exchange

import (
	"github.com/shopspring/decimal"
)

type CommandType int

func (c CommandType) ReflexType() int {
	return int(c)
}

var (
	CommandUnknown CommandType = 0
	CommandPostLimitOrder CommandType = 1
	CommandPostMarketOrder CommandType = 2
	CommandStopOrder CommandType = 3
	kCommandTypeSentinel CommandType = 4
)

type PostLimit struct {
	IsBuy     bool  `json:"is_buy"`
	PostOnly  bool  `json:"post_only"`

	LimitPrice   decimal.Decimal  `json:"limit_price"`
	LimitVolume  decimal.Decimal  `json:"limit_volume"`
}

type PostMarket struct {
	IsBuy  bool  `json:"is_buy"`

	MarketBase    decimal.Decimal `json:"market_base"`
	MarketCounter decimal.Decimal `json:"market_counter"`
}

type StopOrder struct {
	IsBuy bool `json:"is_buy"`
}
