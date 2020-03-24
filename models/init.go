package models

import (
	"github.com/hublabs/order-api/config"

	"github.com/go-xorm/xorm"
)

var (
	eventHandler           EventHandler
	stockEventHandler      StockEventHandler
	calculatorEventHandler CalculatorEventHandler
	benefitHandler         BenefitHandler
)

func Init(db *xorm.Engine) error {
	stockEventHandler.StockApiUrl = config.Config().Services.StockApiUrl
	calculatorEventHandler.CalculatorApiUrl = config.Config().Services.CalculatorApiUrl
	benefitHandler.BenefitApiUrl = config.Config().Services.BenefitApiUrl
	return db.Sync2(new(Order),
		new(OrderItem),
		new(OrderItemSeparate),
		new(OrderReseller),
		new(OrderHistory),
		new(OrderItemHistory),
		new(Refund),
		new(RefundItem),
		new(RefundHistory),
		new(RefundItemHistory),
		new(OrderOffer),
		new(OrderOfferHistory),
		new(ItemAppliedCartOffer),
		new(RefundExtension),
	)
}
func DropTables(db *xorm.Engine) error {
	return db.DropTables(new(Order),
		new(OrderItem),
		new(OrderItemSeparate),
		new(OrderReseller),
		new(OrderHistory),
		new(OrderItemHistory),
		new(Refund),
		new(RefundItem),
		new(RefundHistory),
		new(RefundItemHistory),
		new(OrderOffer),
		new(OrderOfferHistory),
		new(ItemAppliedCartOffer),
		new(RefundExtension),
	)
}
