package exchange

type CommandType int

func (c CommandType) ReflexType() int {
	return int(c)
}

var (
	CommandTypeUnknown CommandType = 0
	CommandTypePostOrder CommandType = 1
	CommandTypeStopOrder CommandType = 2
	kCommandTypeSentinel CommandType = 3
)