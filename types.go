package exchange

type CommandType int

const (
	CommandTypeUnknown   CommandType = 0
	CommandTypePostOrder CommandType = 1
	CommandTypeStopOrder CommandType = 2
	kCommandTypeSentinel CommandType = 3
)

type CommandEvent int

func (c CommandEvent) ReflexType() int {
	return int(c)
}

const (
	EventUnknown CommandEvent = 0
	EventCommandCreated CommandEvent = 1
	kEventSentinel CommandEvent = 2
)

type Command struct {
	ID int64
	Type CommandType
	OrderId int64
}
