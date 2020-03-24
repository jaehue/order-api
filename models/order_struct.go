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

type Order struct {
	Id                 int64              `json:"id" query:"id"`
	TenantCode         string             `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId            int64              `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId          int64              `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	SaleType           string             `json:"saleType" query:"saleType" xorm:"index VARCHAR(30) notnull "`
	CustomerId         int64              `json:"customerId" query:"customerId" xorm:"index notnull" validate:"required"`
	OuterOrderNo       string             `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	WxMiniAppOpenId    string             `json:"wxMiniAppOpenId" query:"wxMiniAppOpenId" xorm:"index VARCHAR(100) notnull" `
	SalesmanId         int64              `json:"salesmanId" query:"salesmanId" xorm:"index notnull" validate:"gte=0"`
	UserClaimIss       string             `json:"userClaimIss" query:"userClaimIss" xorm:"VARCHAR(20) notnull" validate:"required"`
	TotalListPrice     float64            `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice     float64            `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice float64            `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalPaymentPrice  float64            `json:"totalPaymentPrice" query:"totalPaymentPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	FreightPrice       float64            `json:"freightPrice" query:"freightPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ObtainMileage      float64            `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage            float64            `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	MileagePrice       float64            `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	CashPrice          float64            `json:"cashPrice" query:"cashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status             string             `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsOutPaid          bool               `json:"isOutPaid" query:"isOutPaid" xorm:"index default false" `
	IsCustDel          bool               `json:"isCustDel"  xorm:"index default false" `
	CustRemark         string             `json:"custRemark" query:"custRemark" xorm:"VARCHAR(100)"`
	CreatedId          int64              `json:"createdId" query:"createdId" xorm:""`
	CreatedAt          time.Time          `json:"createdAt" query:"createdAt" xorm:"index created"`
	UpdatedAt          time.Time          `json:"updatedAt" query:"updatedAt" xorm:"index updated"`
	Items              []OrderItem        `json:"items" query:"items" xorm:"-" validate:"required,dive,required"`
	Offers             []OrderOffer       `json:"offers" query:"offers" xorm:"-" validate:"required,dive,required"`
	DeliverableAddress DeliverableAddress `json:"deliverableAddress" query:"deliverableAddress" xorm:"-"`

	/* Transaction outbox pattern(https://microservices.io/patterns/data/application-events.html)에서 개념만 채용
	발송 예정인 메시지 이벤트 임시 저장소  */
	eventOutBox []interface{} `xorm:"-"`
}

func (o *Order) InitializeStatus() {
	o.setStatusWithItems(enum.SaleOrderProcessing)
	o.addEventToOutBox(enum.SaleOrderProcessing)
}

func (o *Order) addEventToOutBox(orderType enum.OrderType) {
	if o.eventOutBox == nil {
		o.eventOutBox = make([]interface{}, 0)
	}

	o.eventOutBox = append(o.eventOutBox, orderType)
}

func (o *Order) clearEventOutBox() {
	o.eventOutBox = nil
}

func (o *Order) setStatusWithItems(status enum.OrderType) {
	o.Status = status.String()
	for i := range o.Items {
		o.Items[i].Status = status.String()
	}
}

func (o Order) PublishEventMessagesInOutBox(ctx context.Context) {
	for _, event := range o.eventOutBox {
		eventMessage := o.translateEventMessage(ctx, event.(enum.OrderType))
		logrus.Println(eventMessage)
		if eventMessage != nil {
			key := strconv.FormatInt(o.Id, 10)
			if err := adapters.EventMessagePublisher.Publish(eventMessage, key); err != nil {
				logrus.WithError(err).Error(fmt.Sprintf(`OrderId:%v - publish to message borker`, o.Id))
			}
		}
	}

	o.clearEventOutBox()
}
func (o Order) RePublishEventMessages(ctx context.Context) map[string]interface{} {
	eventMessage := o.translateEventMessage(ctx, enum.FindOrderTypeFromString(o.Status))
	logrus.Println(eventMessage)
	if eventMessage != nil {
		key := strconv.FormatInt(o.Id, 10)
		if err := adapters.EventMessagePublisher.Publish(eventMessage, key); err != nil {
			logrus.WithError(err).Error(fmt.Sprintf(`OrderId:%v - publish to message borker`, o.Id))
		}
	}
	return eventMessage
}
func (o Order) translateEventMessage(ctx context.Context, orderType enum.OrderType) map[string]interface{} {
	orderEvent := o.MakeOrderEvent(orderType)

	return map[string]interface{}{
		"authToken": behaviorlog.FromCtx(ctx).AuthToken,
		"requestId": behaviorlog.FromCtx(ctx).RequestID,
		"payload": map[string]interface{}{
			"status":     orderType.String(),
			"entityType": enum.Order.String(),
			"payload":    orderEvent,
		},
		"status":    orderType.String(),
		"actionId":  behaviorlog.FromCtx(ctx).ActionID,
		"createdAt": time.Now().UTC(),
	}
}

func (o Order) MakeOrderEvent(orderType enum.OrderType) OrderEvent {
	var orderEvent OrderEvent
	orderEvent.Id = o.Id
	orderEvent.OuterOrderNo = o.OuterOrderNo
	orderEvent.TenantCode = o.TenantCode
	orderEvent.StoreId = o.StoreId
	orderEvent.ChannelId = o.ChannelId
	orderEvent.SaleType = o.SaleType
	orderEvent.CustomerId = o.CustomerId
	orderEvent.SalesmanId = o.SalesmanId
	orderEvent.StoreId = o.StoreId
	orderEvent.TotalListPrice = o.TotalListPrice
	orderEvent.TotalSalePrice = o.TotalSalePrice
	orderEvent.TotalDiscountPrice = o.TotalDiscountPrice
	orderEvent.FreightPrice = o.FreightPrice
	orderEvent.TotalPaymentPrice = o.TotalPaymentPrice
	orderEvent.ObtainMileage = o.ObtainMileage
	orderEvent.Mileage = o.Mileage
	orderEvent.MileagePrice = o.MileagePrice
	orderEvent.CashPrice = o.CashPrice
	orderEvent.IsOutPaid = o.IsOutPaid
	orderEvent.Status = orderType.String()
	orderEvent.CreatedId = o.CreatedId
	orderEvent.CreatedAt = o.CreatedAt
	orderEvent.UpdatedAt = o.UpdatedAt

	orderEvent.Items = o.makeOrderItemEvent(orderType)
	if len(orderEvent.Items) == 0 {
		return OrderEvent{}
	}

	orderEvent.Offers = o.makeOrderOfferEvent(orderType)
	orderEvent.Address = o.makeOrderAddressEvent(orderType)

	return orderEvent
}

func (o Order) makeOrderItemEvent(orderType enum.OrderType) []OrderEventItem {
	orderEventItems := make([]OrderEventItem, 0)

	for _, item := range o.Items {
		if orderType == enum.BuyerReceivedConfirmed && item.Status == enum.SaleOrderFinished.String() && item.IsDelivery == true {
			continue
		}
		orderEventItem := OrderEventItem{}
		orderEventItem.Id = item.Id
		orderEventItem.OuterOrderItemNo = item.OuterOrderItemNo
		orderEventItem.ItemCode = item.ItemCode
		orderEventItem.ItemName = item.ItemName
		orderEventItem.ProductId = item.ProductId
		orderEventItem.SkuId = item.SkuId
		orderEventItem.SkuImg = item.SkuImg
		orderEventItem.Option = item.Option
		orderEventItem.ObtainMileage = item.ObtainMileage
		orderEventItem.Mileage = item.Mileage
		orderEventItem.MileagePrice = item.MileagePrice
		orderEventItem.ListPrice = item.ListPrice
		orderEventItem.SalePrice = item.SalePrice
		orderEventItem.Quantity = item.Quantity
		orderEventItem.TotalListPrice = item.TotalListPrice
		orderEventItem.TotalSalePrice = item.TotalSalePrice
		orderEventItem.TotalPaymentPrice = item.TotalPaymentPrice
		orderEventItem.TotalDiscountPrice = item.TotalDiscountPrice
		orderEventItem.TotalDistributedCartOfferPrice = item.TotalDistributedCartOfferPrice
		orderEventItem.ItemFee = item.ItemFee
		orderEventItem.FeeRate = item.FeeRate
		orderEventItem.Status = orderType.String()
		orderEventItem.IsDelivery = item.IsDelivery
		orderEventItem.IsStockChecked = item.IsStockChecked
		orderEventItem.GroupOffers = item.makeOrderItemGroupOfferEvent()
		orderEventItem.CreatedAt = item.CreatedAt
		orderEventItem.UpdatedAt = item.UpdatedAt

		for _, separate := range item.ItemSeparates {
			orderEventItemSeparate := OrderEventItemSeparate{}
			orderEventItemSeparate.Id = separate.Id
			orderEventItemSeparate.StockDistributionItemId = separate.StockDistributionItemId
			orderEventItemSeparate.ListPrice = separate.ListPrice
			orderEventItemSeparate.SalePrice = separate.SalePrice
			orderEventItemSeparate.Quantity = separate.Quantity
			orderEventItemSeparate.TotalListPrice = separate.TotalListPrice
			orderEventItemSeparate.TotalSalePrice = separate.TotalSalePrice
			orderEventItemSeparate.TotalPaymentPrice = separate.TotalPaymentPrice
			orderEventItemSeparate.Status = orderType.String()
			orderEventItem.ItemSeparates = append(orderEventItem.ItemSeparates, orderEventItemSeparate)
		}
		orderEventItems = append(orderEventItems, orderEventItem)
	}

	return orderEventItems
}

func (o Order) makeOrderOfferEvent(orderType enum.OrderType) []Offer {
	offers := []Offer{}
	for _, orderOffer := range o.Offers {
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

func (o OrderItem) makeOrderItemGroupOfferEvent() []ItemOffer {
	offers := []ItemOffer{}
	for _, itemOffer := range o.GroupOffers {
		offer := ItemOffer{}
		offer.OfferNo = itemOffer.OfferNo
		offer.CouponNo = itemOffer.CouponNo
		offer.TargetType = itemOffer.TargetType
		offer.IsTarget = itemOffer.IsTarget
		offer.Price = itemOffer.Price
		offers = append(offers, offer)
	}
	return offers
}

func (o Order) makeOrderAddressEvent(orderType enum.OrderType) DeliverableAddress {
	address := o.DeliverableAddress

	return address
}

type OrderItem struct {
	Id                             int64                  `json:"id" query:"id"`
	OrderId                        int64                  `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	TenantCode                     string                 `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId                        int64                  `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId                      int64                  `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId                     int64                  `json:"customerId" query:"customerId" xorm:"index notnull" validate:"required"`
	OuterOrderNo                   string                 `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	OuterOrderItemNo               string                 `json:"outerOrderItemNo" query:"outerOrderItemNo" xorm:"index VARCHAR(100) notnull" `
	WxMiniAppOpenId                string                 `json:"wxMiniAppOpenId" query:"wxMiniAppOpenId" xorm:"index VARCHAR(100) notnull" `
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
	Quantity                       int                    `json:"quantity" query:"quantity" xorm:"notnull" validate:"required"`
	TotalListPrice                 float64                `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice                 float64                `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice             float64                `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalPaymentPrice              float64                `json:"totalPaymentPrice" query:"totalPaymentPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDistributedCartOfferPrice float64                `json:"totalDistributedCartOfferPrice" query:"totalDistributedCartOfferPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status                         string                 `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsDelivery                     bool                   `json:"isDelivery" query:"isDelivery" xorm:"index default true" `
	IsStockChecked                 bool                   `json:"isStockChecked" query:"isStockChecked" xorm:"index default false" `
	IsDel                          bool                   `json:"isDel"  xorm:"index default false" `
	CreatedAt                      time.Time              `json:"createdAt" query:"createdAt" xorm:"index created"`
	UpdatedAt                      time.Time              `json:"updatedAt" query:"updatedAt" xorm:"index updated"`
	GroupOffers                    []ItemAppliedCartOffer `json:"groupOffers" query:"groupOffers" xorm:"-" validate:"required,dive,required"`
	Resellers                      []OrderReseller        `json:"reseller" query:"reseller" xorm:"-" validate:"required,dive,required"`
	ItemSeparates                  []OrderItemSeparate    `json:"itemSeparates" query:"itemSeparates" xorm:"-" validate:"required,dive,required"`
	RefundItems                    []RefundItem           `json:"refundItems" query:"refundItems" xorm:"-" validate:"required,dive,required"`
	AppliedCartOffers              []ItemAppliedCartOffer `json:"AppliedCartOffers" query:"AppliedCartOffers" xorm:"-" validate:"required,dive,required"`
}

type OrderItemSeparate struct {
	Id                      int64       `json:"id" query:"id"`
	OrderItemId             int64       `json:"orderItemId" query:"orderItemId" xorm:"notnull index" validate:"required"`
	StockDistributionItemId int64       `json:"stockDistributionItemId" query:"stockDistributionItemId" xorm:"notnull index" validate:"gte=0"`
	SalePrice               float64     `json:"salePrice" query:"salePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ListPrice               float64     `json:"listPrice" query:"listPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Quantity                int         `json:"quantity" query:"quantity" xorm:"notnull" validate:"required"`
	TotalListPrice          float64     `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalSalePrice          float64     `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalDiscountPrice      float64     `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	TotalPaymentPrice       float64     `json:"totalPaymentPrice" query:"totalPaymentPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status                  string      `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsDelete                bool        `json:"isDelete" query:"isDelete" xorm:"index default false" `
	RefundItem              *RefundItem `json:"refundItem" query:"refundItem" xorm:"-" validate:"required,dive,required"`
	CreatedAt               time.Time   `json:"createdAt" query:"createdAt" xorm:"index created"`
	UpdatedAt               time.Time   `json:"updatedAt" query:"updatedAt" xorm:"index updated"`
}

type OrderReseller struct {
	Id           int64     `json:"id" query:"id" `
	OrderId      int64     `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OrderItemId  int64     `json:"orderItemId" query:"orderItemId" xorm:"notnull index" validate:"required"`
	ResellerId   int64     `json:"resellerId" query:"resellerId" xorm:"notnull index" validate:"required"`
	ResellerName string    `json:"resellerName" query:"resellerName" xorm:"VARCHAR(100)" `
	CreatedAt    time.Time `json:"createdAt" query:"createdAt" xorm:"index created"`
}
