package models

import (
	"context"
	"time"

	"github.com/hublabs/order-api/enum"
)

type EventHandler struct {
	OsmApi string
}

type Event struct {
	EntityType string     `json:"entityType"` // Order, Refund, OrderDelivery, RefundDelivery, StockDistribution
	Status     string     `json:"status"`
	Payload    OrderEvent `json:"payload"`
	CreatedAt  time.Time  `json:"createdAt" xorm:"created"`
}

type OrderEvent struct {
	Id                 int64              `json:"id,omitempty"`
	OuterOrderNo       string             `json:"outerOrderNo"`
	TenantCode         string             `json:"tenantCode"`
	StoreId            int64              `json:"storeId" `
	ChannelId          int64              `json:"channelId" `
	SaleType           string             `json:"saleType" `
	CustomerId         int64              `json:"customerId"`
	SalesmanId         int64              `json:"salesmanId" `
	TotalListPrice     float64            `json:"totalListPrice"`
	TotalSalePrice     float64            `json:"totalSalePrice"`
	TotalDiscountPrice float64            `json:"totalDiscountPrice"`
	FreightPrice       float64            `json:"freightPrice"`
	TotalPaymentPrice  float64            `json:"totalPaymentPrice"`
	ObtainMileage      float64            `json:"obtainMileage"`
	Mileage            float64            `json:"mileage"`
	MileagePrice       float64            `json:"mileagePrice"`
	CashPrice          float64            `json:"cashPrice"`
	IsOutPaid          bool               `json:"isOutPaid" `
	Status             string             `json:"status"`
	Items              []OrderEventItem   `json:"items,omitempty"`
	Deliveries         []Delivery         `json:"deliveries,omitempty"`
	Refunds            []RefundEvent      `json:"refunds,omitempty"`
	Offers             []Offer            `json:"offers,omitempty"`
	Address            DeliverableAddress `json:"address,omitempty"`
	CreatedId          int64              `json:"createdId,omitempty"`
	CreatedAt          time.Time          `json:"createdAt,omitempty" xorm:"created"`
	UpdatedAt          time.Time          `json:"updatedAt,omitempty" xorm:"updated"`
}

type Offer struct {
	OfferNo        string  `json:"offerNo"`
	CouponNo       string  `json:"couponNo"`
	TargetType     string  `json:"targetType"`
	AdditionalType string  `json:"additionalType"`
	ItemIds        string  `json:"itemIds" `
	TargetItemIds  string  `json:"targetItemIds" `
	Price          float64 `json:"price"`
}
type ItemOffer struct {
	OfferNo    string  `json:"offerNo"`
	CouponNo   string  `json:"couponNo"`
	TargetType string  `json:"targetType"`
	IsTarget   bool    `json:"isTarget" `
	Price      float64 `json:"price"`
}
type OrderEventItem struct {
	Id                             int64                    `json:"id,omitempty"`
	OuterOrderItemNo               string                   `json:"outerOrderItemNo"`
	ItemCode                       string                   `json:"itemCode"`
	ItemName                       string                   `json:"itemName"`
	ProductId                      int64                    `json:"productId"`
	SkuId                          int64                    `json:"skuId"`
	SkuImg                         string                   `json:"skuImg"`
	ObtainMileage                  float64                  `json:"obtainMileage"`
	Mileage                        float64                  `json:"mileage"`
	MileagePrice                   float64                  `json:"mileagePrice"`
	Option                         string                   `json:"option"`
	ListPrice                      float64                  `json:"listPrice"`
	SalePrice                      float64                  `json:"salePrice"`
	Quantity                       int                      `json:"quantity"`
	TotalListPrice                 float64                  `json:"totalListPrice"`
	TotalSalePrice                 float64                  `json:"totalSalePrice"`
	TotalDiscountPrice             float64                  `json:"totalDiscountPrice"`
	TotalPaymentPrice              float64                  `json:"totalPaymentPrice"`
	TotalDistributedCartOfferPrice float64                  `json:"totalDistributedCartOfferPrice"`
	ItemFee                        float64                  `json:"itemFee"`
	FeeRate                        float64                  `json:"feeRate"`
	Status                         string                   `json:"status"`
	IsDelivery                     bool                     `json:"isDelivery" `
	IsStockChecked                 bool                     `json:"isStockChecked" `
	GroupOffers                    []ItemOffer              `json:"groupOffers,omitempty"`
	ItemSeparates                  []OrderEventItemSeparate `json:"separates,omitempty"`
	StockDistributionItems         []StockDistributionItem  `json:"stockDistributionItems,omitempty"`
	CreatedAt                      time.Time                `json:"createdAt"`
	UpdatedAt                      time.Time                `json:"updatedAt"`
}

type OrderEventItemSeparate struct {
	Id                      int64     `json:"id,omitempty"`
	StockDistributionItemId int64     `json:"stockDistributionItemId"`
	ListPrice               float64   `json:"listPrice"`
	SalePrice               float64   `json:"salePrice"`
	Quantity                int       `json:"quantity"`
	TotalListPrice          float64   `json:"totalListPrice"`
	TotalSalePrice          float64   `json:"totalSalePrice"`
	TotalPaymentPrice       float64   `json:"totalPaymentPrice"`
	Status                  string    `json:"status"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type StockDistributionItem struct {
	StockDistributionItemId int64 `json:"id,omitempty"`
	Quantity                int   `json:"quantity"`
	IsCanceled              bool  `json:"isCanceled"`
}

type RefundEvent struct {
	Id                 int64                `json:"id"`
	OuterOrderNo       string               `json:"outerOrderNo"`
	TenantCode         string               `json:"tenantCode"`
	StoreId            int64                `json:"storeId" `
	ChannelId          int64                `json:"channelId" `
	RefundType         string               `json:"refundType"`
	CustomerId         int64                `json:"customerId"`
	SalesmanId         int64                `json:"salesmanId" `
	TotalListPrice     float64              `json:"totalListPrice"`
	TotalSalePrice     float64              `json:"totalSalePrice"`
	TotalDiscountPrice float64              `json:"totalDiscountPrice"`
	FreightPrice       float64              `json:"freightPrice"`
	TotalRefundPrice   float64              `json:"totalRefundPrice"`
	ObtainMileage      float64              `json:"obtainMileage"`
	Mileage            float64              `json:"mileage"`
	MileagePrice       float64              `json:"mileagePrice"`
	CashPrice          float64              `json:"cashPrice"`
	Status             string               `json:"status"`
	IsOutPaid          bool                 `json:"isOutPaid" `
	CustRemark         string               `json:"custRemark" `
	Items              []RefundEventItem    `json:"items,omitempty"`
	Address            DeliverableAddress   `json:"address,omitempty"`
	Extension          RefundExtensionEvent `json:"extension,omitempty"`
	RefundReason       string               `json:"refundReason"`
	RefuseReason       string               `json:"refuseReason"`
	Offers             []Offer              `json:"offers,omitempty"`
	Deliveries         []Delivery           `json:"deliveries,omitempty"`
	CreatedId          int64                `json:"createdId,omitempty"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
}
type RefundExtensionEvent struct {
	ImgUrl string `json:"imgUrl"`
}
type RefundEventItem struct {
	Id                             int64       `json:"id"`
	OuterOrderItemNo               string      `json:"outerOrderItemNo"`
	OrderItemId                    int64       `json:"orderItemId"`
	SeparateId                     int64       `json:"separateId"`
	StockDistributionItemId        int64       `json:"stockDistributionItemId"`
	ItemCode                       string      `json:"itemCode"`
	ItemName                       string      `json:"itemName"`
	ProductId                      int64       `json:"productId"`
	SkuId                          int64       `json:"skuId"`
	SkuImg                         string      `json:"skuImg"`
	Option                         string      `json:"option"`
	ObtainMileage                  float64     `json:"obtainMileage"`
	Mileage                        float64     `json:"mileage"`
	MileagePrice                   float64     `json:"mileagePrice"`
	ListPrice                      float64     `json:"listPrice"`
	SalePrice                      float64     `json:"salePrice"`
	Quantity                       int         `json:"quantity"`
	TotalListPrice                 float64     `json:"totalListPrice"`
	TotalSalePrice                 float64     `json:"totalSalePrice"`
	TotalDiscountPrice             float64     `json:"totalDiscountPrice"`
	TotalRefundPrice               float64     `json:"totalRefundPrice"`
	TotalDistributedCartOfferPrice float64     `json:"totalDistributedCartOfferPrice"`
	ItemFee                        float64     `json:"itemFee"`
	FeeRate                        float64     `json:"feeRate"`
	Status                         string      `json:"status"`
	IsDelivery                     bool        `json:"isDelivery" `
	GroupOffers                    []ItemOffer `json:"groupOffers,omitempty"`
	CreatedAt                      time.Time   `json:"createdAt"`
	UpdatedAt                      time.Time   `json:"updatedAt"`
}
type Delivery struct {
	Id           int64     `json:"id,omitempty"`
	OrderItemId  int64     `json:"orderItemId"`
	SeparateId   int64     `json:"separateId"`
	RefundItemId int64     `json:"refundItemId"`
	Quantity     int       `json:"quantity"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (EventHandler) HandleEvent(ctx context.Context, event Event) error {
	entityType := event.EntityType
	statusType := event.Status
	orderId := event.Payload.Id
	orderItemIds := []int64{}
	separateIds := []int64{}

	if entityType == enum.Order.String() || entityType == enum.OrderDelivery.String() {
		if entityType == enum.Order.String() {
			for _, item := range event.Payload.Items {
				orderItemIds = append(orderItemIds, item.Id)
				for _, separate := range item.ItemSeparates {
					if separate.Id != 0 {
						separateIds = append(separateIds, separate.Id)
					}
				}
			}
		} else if entityType == enum.OrderDelivery.String() {
			for _, item := range event.Payload.Deliveries {
				orderItemIds = append(orderItemIds, item.OrderItemId)
				if item.SeparateId != 0 {
					separateIds = append(separateIds, item.SeparateId)
				}
			}
		}
		if err := OrderChangeStatusEvent(ctx, statusType, orderId, orderItemIds, separateIds, DeliverableAddress{}); err != nil {
			return err
		}
	} else if entityType == enum.StockDistribution.String() {
		var separateOrder SeparateOrder
		separateOrder.OrderId = orderId
		for _, item := range event.Payload.Items {
			var separateOrderItem SeparateOrderItem
			separateOrderItem.OrderItemId = item.Id
			orderItemIds = append(orderItemIds, item.Id)
			for _, stockDistributionItem := range item.StockDistributionItems {
				var itemSeparate ItemSeparate
				itemSeparate.StockDistributionItemId = stockDistributionItem.StockDistributionItemId
				itemSeparate.Quantity = stockDistributionItem.Quantity
				itemSeparate.IsDelete = stockDistributionItem.IsCanceled
				separateOrderItem.ItemSeparates = append(separateOrderItem.ItemSeparates, itemSeparate)
			}
			separateOrder.Items = append(separateOrder.Items, separateOrderItem)
		}

		orderItemSeparates, err := separateOrder.OrderItemSparateEvent(ctx, false)
		if err != nil {
			return err
		}
		for _, OrderItemSeparate := range orderItemSeparates {
			separateIds = append(separateIds, OrderItemSeparate.Id)
		}
		if err := OrderChangeStatusEvent(ctx, statusType, orderId, orderItemIds, separateIds, DeliverableAddress{}); err != nil {
			return err
		}
	} else if entityType == enum.Refund.String() || entityType == enum.RefundDelivery.String() {
		for _, refund := range event.Payload.Refunds {
			refundId := refund.Id
			refundItemIds := []int64{}
			if entityType == enum.Refund.String() {
				for _, item := range refund.Items {
					refundItemIds = append(refundItemIds, item.Id)
				}
			} else if entityType == enum.RefundDelivery.String() {
				for _, item := range refund.Deliveries {
					refundItemIds = append(refundItemIds, item.RefundItemId)
				}
			}
			if err := RefundChangeStatusEvent(ctx, statusType, refundId, refundItemIds, refund.RefuseReason, refund.Address); err != nil {
				return err
			}
		}
	}

	return nil
}

func OrderChangeStatusEvent(ctx context.Context, statusType string, orderId int64, orderItemIds []int64, separateIds []int64, address DeliverableAddress) error {
	switch statusType {
	case enum.SaleOrderProcessing.String(), enum.SaleOrderFinished.String(), enum.SaleOrderCancel.String():
		order, err := (Order{}).ChangeStatus(ctx, orderId, []int64{}, []int64{}, statusType)
		if err != nil {
			return err
		} else {
			go func() {
				// 변경된 상태를 메시지 이벤트로 발행
				order.PublishEventMessagesInOutBox(ctx)
			}()
		}
	case enum.StockDistributed.String(), enum.SaleShippingWaiting.String(), enum.SaleShippingProcessing.String(), enum.SaleShippingFinished.String(), enum.BuyerReceivedConfirmed.String(), enum.SaleOrderSuccess.String():
		if order, err := (Order{}).ChangeStatus(ctx, orderId, orderItemIds, separateIds, statusType); err != nil {
			return err
		} else {
			go func() {
				// 변경된 상태를 메시지 이벤트로 발행
				order.PublishEventMessagesInOutBox(ctx)
			}()
		}
	}

	return nil
}

func RefundChangeStatusEvent(ctx context.Context, statusType string, refundId int64, refundItemIds []int64, refuseReason string, address DeliverableAddress) error {
	r, err := Refund{}.GetRefund(ctx, "", 0, refundId, 0, refundItemIds, "", true)
	if err != nil {
		return err
	}
	if r.Id == 0 {
		return NotFoundDataError
	}
	refund := &r
	refund.DeliverableAddress = address
	refund.RefuseReason = refuseReason
	switch statusType {
	case enum.RefundOrderRegistered.String(), enum.RefundOrderCancel.String(), enum.RefundOrderProcessing.String():
		if _, err := refund.ChangeStatus(ctx, statusType); err != nil {
			return err
		} else {
			// go func() {
			// 변경된 상태를 메시지 이벤트로 발행
			refund.PublishEventMessagesInOutBox(ctx)
			// }()
		}

	case enum.RefundShippingWaiting.String(), enum.RefundShippingProcessing.String(), enum.RefundShippingFinished.String(), enum.RefundRequisiteApprovals.String(), enum.RefundOrderSuccess.String():
		if _, err := refund.ChangeStatus(ctx, statusType); err != nil {
			return err
		} else {
			// go func() {
			// 변경된 상태를 메시지 이벤트로 발행
			refund.PublishEventMessagesInOutBox(ctx)
			// }()
		}
	}

	return nil
}
