package enum

type OrderType int

const (
	SaleOrderProcessing OrderType = 1 + iota
	SaleOrderCancel
	SaleOrderFinished
	StockDistributed
	SaleShippingWaiting
	SaleShippingProcessing
	SaleShippingFinished
	BuyerReceivedConfirmed
	SaleOrderSuccess
	RefundOrderRegistered
	SellerRefundAgree
	RefundOrderCancel
	RefundOrderProcessing
	RefundShippingWaiting
	RefundShippingProcessing
	RefundShippingFinished
	RefundRequisiteApprovals
	RefundOrderSuccess
	Unknown
)

var orderTypes = [...]string{
	"SaleOrderProcessing",
	"SaleOrderCancel",
	"SaleOrderFinished",
	"StockDistributed",
	"SaleShippingWaiting",
	"SaleShippingProcessing",
	"SaleShippingFinished",
	"BuyerReceivedConfirmed",
	"SaleOrderSuccess",
	"RefundOrderRegistered",
	"SellerRefundAgree",
	"RefundOrderCancel",
	"RefundOrderProcessing",
	"RefundShippingWaiting",
	"RefundShippingProcessing",
	"RefundShippingFinished",
	"RefundRequisiteApprovals",
	"RefundOrderSuccess",
}

func (o OrderType) String() string { return orderTypes[o-1] }

func FindOrderTypeFromString(str string) OrderType {
	switch str {
	case "SaleOrderProcessing":
		return SaleOrderProcessing
	case "SaleOrderCancel":
		return SaleOrderCancel
	case "SaleOrderFinished":
		return SaleOrderFinished
	case "StockDistributed":
		return StockDistributed
	case "SaleShippingWaiting":
		return SaleShippingWaiting
	case "SaleShippingProcessing":
		return SaleShippingProcessing
	case "SaleShippingFinished":
		return SaleShippingFinished
	case "BuyerReceivedConfirmed":
		return BuyerReceivedConfirmed
	case "SaleOrderSuccess":
		return SaleOrderSuccess
	case "RefundOrderRegistered":
		return RefundOrderRegistered
	case "SellerRefundAgree":
		return SellerRefundAgree
	case "RefundOrderCancel":
		return RefundOrderCancel
	case "RefundOrderProcessing":
		return RefundOrderProcessing
	case "RefundShippingWaiting":
		return RefundShippingWaiting
	case "RefundShippingProcessing":
		return RefundShippingProcessing
	case "RefundShippingFinished":
		return RefundShippingFinished
	case "RefundRequisiteApprovals":
		return RefundRequisiteApprovals
	case "RefundOrderSuccess":
		return RefundOrderSuccess
	default:
		return Unknown
	}
}
