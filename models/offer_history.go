package models

import (
	"context"
	"time"

	"github.com/hublabs/order-api/factory"
)

type OrderOfferHistory struct {
	Id             int64     `json:"id" query:"id" `
	OrderOfferId   int64     `json:"orderOfferId" query:"orderOfferId" xorm:"index notnull" validate:"required"`
	TenantCode     string    `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId        int64     `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId      int64     `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId     int64     `json:"customerId" query:"customerId" xorm:"index notnull" validate:"gte=0"`
	OfferNo        string    `json:"offerNo" query:"offerNo" xorm:"index default" validate:"required"`
	CouponNo       string    `json:"couponNo" query:"couponNo" xorm:"index default" validate:"required"`
	TargetType     string    `json:"targetType" xorm:"VARCHAR(10)"`
	AdditionalType string    `json:"additionalType" xorm:"VARCHAR(10)"`
	ItemIds        string    `json:"itemIds" query:"itemIds" xorm:"TEXT"`
	TargetItemIds  string    `json:"targetItemIds" query:"targetItemIds" xorm:"TEXT" `
	Description    string    `json:"description" query:"description" xorm:"VARCHAR(200) notnull" validate:"required"`
	OrderId        int64     `json:"orderId" query:"orderId" xorm:"index default 0" validate:"required"`
	RefundId       int64     `json:"refundId" query:"refundId" xorm:"index default 0" `
	Price          float64   `json:"price" query:"price" xorm:"DECIMAL(18,2) default 0" validate:"required"`
	CreatedAt      time.Time `json:"createdAt" query:"createdAt" xorm:"created"`
}

func (o *OrderOffer) NewOrderOfferHistory(userClaim UserClaim) OrderOfferHistory {
	var orderOfferHistory OrderOfferHistory
	orderOfferHistory.OrderOfferId = o.Id
	orderOfferHistory.TenantCode = userClaim.tenantCode()
	orderOfferHistory.StoreId = o.StoreId
	orderOfferHistory.ChannelId = userClaim.channelId()
	orderOfferHistory.CustomerId = o.CustomerId
	orderOfferHistory.OfferNo = o.OfferNo
	orderOfferHistory.CouponNo = o.CouponNo
	orderOfferHistory.TargetType = o.TargetType
	orderOfferHistory.AdditionalType = o.AdditionalType
	orderOfferHistory.ItemIds = o.ItemIds
	orderOfferHistory.TargetItemIds = o.TargetItemIds
	orderOfferHistory.Description = o.Description
	orderOfferHistory.OrderId = o.OrderId
	orderOfferHistory.RefundId = o.RefundId
	orderOfferHistory.Price = o.Price

	return orderOfferHistory
}

func (m *OrderOfferHistory) Save(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(m)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	return nil
}
