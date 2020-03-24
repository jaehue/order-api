package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hublabs/order-api/adapters"
	"github.com/hublabs/order-api/enum"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/sirupsen/logrus"
)

type Refund struct {
	Id                 int64              `json:"id" query:"id"`
	TenantCode         string             `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId            int64              `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId          int64              `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	RefundType         string             `json:"refundType" query:"refundType" xorm:"index VARCHAR(30) default NULL "`
	CustomerId         int64              `json:"customerId" query:"customerId" xorm:"index notnull" validate:"required"`
	SalesmanId         int64              `json:"salesmanId" query:"salesmanId" xorm:"index notnull" validate:"gte=0"`
	OrderId            int64              `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OuterOrderNo       string             `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	TotalListPrice     float64            `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice     float64            `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice float64            `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalRefundPrice   float64            `json:"totalRefundPrice" query:"totalRefundPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	FreightPrice       float64            `json:"freightPrice" query:"freightPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ObtainMileage      float64            `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage            float64            `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	MileagePrice       float64            `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	CashPrice          float64            `json:"cashPrice" query:"cashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status             string             `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	RefundReason       string             `json:"reason" query:"reason" xorm:"VARCHAR(100)"`
	RefuseReason       string             `json:"refuseReason" query:"refuseReason" xorm:"VARCHAR(100)"`
	CustRemark         string             `json:"custRemark" query:"custRemark" xorm:"VARCHAR(500)"`
	IsOutPaid          bool               `json:"isOutPaid" query:"isOutPaid" xorm:"index default false" `
	CreatedId          int64              `json:"createdId" query:"createdId`
	CreatedAt          time.Time          `json:"createdAt" query:"createdAt" xorm:"index created"`
	UpdatedAt          time.Time          `json:"updatedAt" query:"updatedAt" xorm:"index updated"`
	Items              []RefundItem       `json:"items" query:"items" xorm:"-" validate:"required,dive,required"`
	Offers             []OrderOffer       `json:"offers" query:"offers" xorm:"-" validate:"required,dive,required"`
	DeliverableAddress DeliverableAddress `json:"deliverableAddress" query:"deliverableAddress" xorm:"-"`
	Extension          RefundExtension    `json:"extension"  xorm:"-"`
	eventOutBox        []interface{}      `xorm:"-"`
}

type RefundExtension struct {
	Id       int64  `json:"id" query:"id"`
	RefundId int64  `json:"refundId" query:"refundId" xorm:"index notnull"`
	ImgUrl   string `json:"imgUrl" xorm:"VARCHAR(200)" validate:""`
}

func (r *Refund) addEventToOutBox(status string) {
	if r.eventOutBox == nil {
		r.eventOutBox = make([]interface{}, 0)
	}

	r.eventOutBox = append(r.eventOutBox, status)
}

func (r *Refund) clearEventOutBox() {
	r.eventOutBox = nil
}

func (r Refund) PublishEventMessagesInOutBox(ctx context.Context) {
	for _, changedStatus := range r.eventOutBox {
		eventMessage := r.translateEventMessage(ctx, changedStatus.(string))
		key := strconv.FormatInt(r.Id, 10)
		if err := adapters.EventMessagePublisher.Publish(eventMessage, key); err != nil {
			logrus.WithError(err).Error(fmt.Sprintf(`OrderId:%v - publish to message borker`, r.Id))
		}
	}

	r.clearEventOutBox()
}
func (r Refund) RePublishEventMessages(ctx context.Context) map[string]interface{} {
	eventMessage := r.translateEventMessage(ctx, r.Status)
	key := strconv.FormatInt(r.Id, 10)
	if err := adapters.EventMessagePublisher.Publish(eventMessage, key); err != nil {
		logrus.WithError(err).Error(fmt.Sprintf(`OrderId:%v - publish to message borker`, r.Id))
	}
	return eventMessage
}
func (r Refund) translateEventMessage(ctx context.Context, changedStatus string) map[string]interface{} {
	orderEvent := r.MakeRefundOrderEvent(ctx, changedStatus)

	return map[string]interface{}{
		"authToken": behaviorlog.FromCtx(ctx).AuthToken,
		"requestId": behaviorlog.FromCtx(ctx).RequestID,
		"payload": map[string]interface{}{
			"status":     changedStatus,
			"entityType": enum.Refund.String(),
			"payload":    orderEvent,
		},
		"status":    changedStatus,
		"actionId":  behaviorlog.FromCtx(ctx).ActionID,
		"createdAt": time.Now().UTC(),
	}
}

func (r Refund) MakeRefundOrderEvent(ctx context.Context, changedStatus string) OrderEvent {
	var orderEvent OrderEvent
	order, _ := Order{}.GetOrder(ctx, "", 0, r.OrderId, nil, "", true)
	orderEvent.Id = r.OrderId
	orderEvent.TenantCode = order.TenantCode
	orderEvent.CreatedAt = order.CreatedAt

	refundEvent := RefundEvent{}
	refundEvent.Id = r.Id
	refundEvent.OuterOrderNo = r.OuterOrderNo
	refundEvent.TenantCode = r.TenantCode
	refundEvent.StoreId = r.StoreId
	refundEvent.ChannelId = r.ChannelId
	refundEvent.RefundType = r.RefundType
	refundEvent.CustomerId = r.CustomerId
	refundEvent.SalesmanId = r.SalesmanId
	refundEvent.TotalListPrice = r.TotalListPrice
	refundEvent.TotalSalePrice = r.TotalSalePrice
	refundEvent.TotalDiscountPrice = r.TotalDiscountPrice
	refundEvent.FreightPrice = r.FreightPrice
	refundEvent.TotalRefundPrice = r.TotalRefundPrice
	refundEvent.ObtainMileage = r.ObtainMileage
	refundEvent.Mileage = r.Mileage
	refundEvent.MileagePrice = r.MileagePrice
	refundEvent.CashPrice = r.CashPrice
	refundEvent.Status = changedStatus
	refundEvent.IsOutPaid = r.IsOutPaid
	refundEvent.RefundReason = r.RefundReason
	refundEvent.RefuseReason = r.RefuseReason
	refundEvent.CustRemark = r.CustRemark
	extension := RefundExtensionEvent{}
	extension.ImgUrl = r.Extension.ImgUrl
	refundEvent.Extension = extension
	refundEvent.CreatedId = r.CreatedId
	refundEvent.CreatedAt = r.CreatedAt
	refundEvent.UpdatedAt = r.UpdatedAt

	refundEvent.Items = r.makeRefundItemEvent(changedStatus)
	if len(refundEvent.Items) == 0 {
		return OrderEvent{}
	}

	refundEvent.Offers = r.makeRefundOfferEvent()
	refundEvent.Address = r.makeRefundAddressEvent()

	orderEvent.Refunds = append(orderEvent.Refunds, refundEvent)

	return orderEvent
}

func (r Refund) makeRefundItemEvent(changedStatus string) []RefundEventItem {
	refundEventItems := make([]RefundEventItem, 0)

	for _, item := range r.Items {
		refundEventItem := RefundEventItem{}
		refundEventItem.Id = item.Id
		refundEventItem.OrderItemId = item.OrderItemId
		refundEventItem.SeparateId = item.SeparateId
		refundEventItem.StockDistributionItemId = item.StockDistributionItemId
		refundEventItem.ItemCode = item.ItemCode
		refundEventItem.ItemName = item.ItemName
		refundEventItem.ProductId = item.ProductId
		refundEventItem.SkuId = item.SkuId
		refundEventItem.SkuImg = item.SkuImg
		refundEventItem.Option = item.Option
		refundEventItem.ObtainMileage = item.ObtainMileage
		refundEventItem.Mileage = item.Mileage
		refundEventItem.MileagePrice = item.MileagePrice
		refundEventItem.ListPrice = item.ListPrice
		refundEventItem.SalePrice = item.SalePrice
		refundEventItem.Quantity = item.Quantity
		refundEventItem.TotalListPrice = item.TotalListPrice
		refundEventItem.TotalSalePrice = item.TotalSalePrice
		refundEventItem.TotalRefundPrice = item.TotalRefundPrice
		refundEventItem.TotalDiscountPrice = item.TotalDiscountPrice
		refundEventItem.TotalDistributedCartOfferPrice = item.TotalDistributedCartOfferPrice
		refundEventItem.ItemFee = item.ItemFee
		refundEventItem.FeeRate = item.FeeRate
		refundEventItem.Status = changedStatus
		refundEventItem.IsDelivery = item.IsDelivery
		refundEventItem.GroupOffers = item.makeRefundItemGroupOfferEvent()
		refundEventItem.CreatedAt = item.CreatedAt
		refundEventItem.UpdatedAt = item.UpdatedAt

		refundEventItems = append(refundEventItems, refundEventItem)
	}
	return refundEventItems
}

func (r Refund) makeRefundOfferEvent() []Offer {
	offers := []Offer{}

	for _, orderOffer := range r.Offers {
		offer := Offer{}
		offer.OfferNo = orderOffer.OfferNo
		offer.CouponNo = orderOffer.CouponNo
		offer.TargetType = orderOffer.TargetType
		offer.AdditionalType = orderOffer.AdditionalType
		offer.ItemIds = orderOffer.ItemIds
		offer.TargetItemIds = orderOffer.TargetItemIds
		offer.Price = orderOffer.Price
		offers = append(offers, offer)
	}

	return offers
}

func (r RefundItem) makeRefundItemGroupOfferEvent() []ItemOffer {
	offers := []ItemOffer{}
	for _, itemOffer := range r.GroupOffers {
		offer := ItemOffer{}
		offer.OfferNo = itemOffer.OfferNo
		offer.CouponNo = itemOffer.CouponNo
		offer.TargetType = itemOffer.TargetType
		offer.Price = itemOffer.Price
		offers = append(offers, offer)
	}
	return offers
}

func (r Refund) makeRefundAddressEvent() DeliverableAddress {
	address := r.DeliverableAddress
	return address
}

type RefundItem struct {
	Id                             int64                  `json:"id" query:"id"`
	TenantCode                     string                 `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId                        int64                  `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId                      int64                  `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId                     int64                  `json:"customerId" query:"customerId" xorm:"index notnull" validate:"required"`
	RefundId                       int64                  `json:"refundId" query:"refundId" xorm:"index notnull"`
	OrderId                        int64                  `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OrderItemId                    int64                  `json:"orderItemId" query:"orderItemId" xorm:"index notnull" validate:"required"`
	SeparateId                     int64                  `json:"separateId" query:"separateId" xorm:"default 0 index" `
	StockDistributionItemId        int64                  `json:"stockDistributionItemId" query:"stockDistributionItemId" xorm:"default 0 index"`
	OuterOrderNo                   string                 `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	OuterOrderItemNo               string                 `json:"outerOrderItemNo" query:"outerOrderItemNo" xorm:"index VARCHAR(100) notnull" `
	ItemCode                       string                 `json:"itemCode" query:"itemCode" xorm:"index notnull" validate:"required"`
	ItemName                       string                 `json:"itemName" query:"itemName" xorm:"notnull" validate:"required"`
	ItemFee                        float64                `json:"itemFee" query:"itemFee" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	FeeRate                        float64                `json:"feeRate" query:"feeRate" xorm:"DECIMAL(18,4) default 0.00" validate:"gte=0"`
	ProductId                      int64                  `json:"productId" query:"productId" xorm:"index notnull" validate:"gte=0"`
	SkuId                          int64                  `json:"skuId" query:"skuId" xorm:"notnull" validate:"gte=0"`
	SkuImg                         string                 `json:"skuImg" query:"skuImg" xorm:"VARCHAR(200)" validate:""`
	Option                         string                 `json:"option" query:"option" xorm:"VARCHAR(100)" validate:""`
	ObtainMileage                  float64                `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage                        float64                `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	MileagePrice                   float64                `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ListPrice                      float64                `json:"listPrice" query:"listPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	SalePrice                      float64                `json:"salePrice" query:"salePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Quantity                       int                    `json:"quantity" query:"quantity" xorm:"default 0" validate:"gte=0"`
	TotalListPrice                 float64                `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice                 float64                `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice             float64                `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalRefundPrice               float64                `json:"totalRefundPrice" query:"totalRefundPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDistributedCartOfferPrice float64                `json:"totalDistributedCartOfferPrice" query:"totalDistributedCartOfferPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status                         string                 `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsDelivery                     bool                   `json:"isDelivery" query:"isDelivery" xorm:"index default true" `
	GroupOffers                    []ItemAppliedCartOffer `json:"groupOffers" query:"groupOffers" xorm:"-" validate:"required,dive,required"`
	CreatedAt                      time.Time              `json:"createdAt" query:"createdAt" xorm:"index created"`
	UpdatedAt                      time.Time              `json:"updatedAt" query:"updatedAt" xorm:"index updated"`
}
