package enum

type EventType int

const (
	Order EventType = 1 + iota
	StockDistribution
	Refund
	OrderDelivery
	RefundDelivery
)

var eventTypes = [...]string{
	"Order",
	"StockDistribution",
	"Refund",
	"OrderDelivery",
	"RefundDelivery",
}

func (o EventType) String() string { return eventTypes[o-1] }
