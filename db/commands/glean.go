package commands

import "exchange"

//go:generate glean -src=../../../ -table=commands -scan

type glean struct {
	exchange.Command
}