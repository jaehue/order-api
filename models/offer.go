package models

import (
	"context"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/factory"

	"github.com/go-xorm/xorm"
)

type OrderOffer struct {
	Id             int64     `json:"id" query:"id" `
	TenantCode     string    `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId        int64     `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId      int64     `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId     int64     `json:"customerId" query:"customerId" xorm:"index notnull" validate:"gte=0"`
	OfferNo        string    `json:"offerNo" query:"offerNo" xorm:"index default" validate:"required"`
	CouponNo       string    `json:"couponNo" query:"couponNo" xorm:"index default" validate:"required"`
	TargetType     string    `json:"targetType" xorm:"VARCHAR(10)"`
	AdditionalType string    `json:"additionalType" xorm:"VARCHAR(10)"`
	ItemIds        string    `json:"itemIds" query:"itemIds" xorm:"TEXT" `
	TargetItemIds  string    `json:"targetItemIds" query:"targetItemIds" xorm:"TEXT" `
	Description    string    `json:"description" query:"description" xorm:"VARCHAR(200) notnull" validate:"required"`
	OrderId        int64     `json:"orderId" query:"orderId" xorm:"index default 0" validate:"required"`
	RefundId       int64     `json:"refundId" query:"refundId" xorm:"index default 0" `
	Price          float64   `json:"price" query:"price" xorm:"DECIMAL(18,2) default 0" validate:"required"`
	CreatedAt      time.Time `json:"createdAt" query:"createdAt" xorm:"created"`
	// ItemCodes       []string  `json:"itemCodes" query:"itemCodes" xorm:"-" `
	// TargetItemCodes []string  `json:"targetItemCodes" query:"targetItemCodes" xorm:"-" `
}
type ItemAppliedCartOffer struct {
	Id             int64     `json:"id" query:"id" `
	OfferNo        string    `json:"offerNo" query:"offerNo" xorm:"index default" validate:"required"`
	CouponNo       string    `json:"couponNo" query:"couponNo" xorm:"index default" validate:"required"`
	AdditionalType string    `json:"additionalType" xorm:"-"`
	TargetType     string    `json:"targetType" xorm:"VARCHAR(10)"`
	IsTarget       bool      `json:"isTarget" query:"isTarget" xorm:"index default false" `
	OrderId        int64     `json:"orderId" query:"orderId" xorm:"index default 0" validate:"required"`
	RefundId       int64     `json:"refundId" query:"refundId" xorm:"index default 0" `
	OrderItemId    int64     `json:"orderItemId" query:"orderItemId" xorm:"notnull index" validate:"required"`
	RefundItemId   int64     `json:"refundItemId" query:"refundItemId" xorm:"notnull index" validate:""`
	Price          float64   `json:"price" query:"price" xorm:"DECIMAL(18,2) default 0" validate:"required"`
	CreatedAt      time.Time `json:"createdAt" query:"createdAt" xorm:"created"`
}

func (OrderOffer) GetOrderOffer(ctx context.Context, tenantCode string, customerId int64, offerNo string, orderId int64, refundId int64) ([]OrderOffer, error) {
	var orderOffer []OrderOffer
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
	}

	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Where("1=1")
		if userClaim.TenantCode != "admin" {
			q.And("tenant_code = ?", userClaim.TenantCode)
		} else {
			if tenantCode != "" {
				q.And("tenant_code = ?", tenantCode)
			}
		}
		if userClaim.Iss == auth.IssMembership {
			q.And("customer_id = ?", customerId)
		} else if customerId != 0 {
			q.And("customer_id = ?", customerId)
		}
		if offerNo != "" {
			q.And("offer_no = ?", offerNo)
		}
		if orderId != 0 {
			q.And("order_id = ?", orderId)
		}
		if refundId != 0 {
			q.And("refund_id = ?", refundId)
		}
		return q
	}

	q := queryBuilder()
	if err := q.Find(&orderOffer); err != nil {
		return nil, err
	}

	return orderOffer, nil
}

func (OrderOffer) GetOrderOffers(ctx context.Context, tenantCode string, customerId int64, offerNo string, orderIds []int64, refundIds []int64) ([]OrderOffer, error) {
	var orderOffers []OrderOffer
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
	}

	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Where("1=1")
		if userClaim.TenantCode != "admin" {
			q.And("tenant_code = ?", userClaim.TenantCode)
		} else {
			if tenantCode != "" {
				q.And("tenant_code = ?", tenantCode)
			}
		}
		if userClaim.Iss == auth.IssMembership {
			q.And("customer_id = ?", customerId)
		} else if customerId != 0 {
			q.And("customer_id = ?", customerId)
		}
		if offerNo != "" {
			q.And("offer_no = ?", offerNo)
		}
		if len(orderIds) > 0 {
			q.In("order_id", orderIds)
		}
		if len(refundIds) > 0 {
			q.In("refund_id", refundIds)
		}
		return q
	}

	q := queryBuilder()
	if err := q.Find(&orderOffers); err != nil {
		return nil, err
	}

	return orderOffers, nil
}
func (ItemAppliedCartOffer) GetItemAppliedOffers(ctx context.Context, offerNo string, orderIds, orderItemIds, refundIds, refundItemIds []int64) ([]ItemAppliedCartOffer, error) {
	var appliedCartOffers []ItemAppliedCartOffer
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Where("1=1")
		if offerNo != "" {
			q.And("offer_no = ?", offerNo)
		}
		if len(orderIds) > 0 {
			q.In("order_id", orderIds)
		}
		if len(orderItemIds) > 0 {
			q.In("order_item_id", orderItemIds)
		}
		if len(refundIds) > 0 {
			q.In("refund_id", refundIds)
		}
		if len(refundItemIds) > 0 {
			q.In("refund_item_id", refundItemIds)
		}
		return q
	}

	q := queryBuilder()
	if err := q.Find(&appliedCartOffers); err != nil {
		return nil, err
	}

	return appliedCartOffers, nil
}
func (offer *OrderOffer) Save(ctx context.Context) error {
	if offer.OfferNo == "" || offer.Price < 0 {
		return nil
	}
	row, err := factory.DB(ctx).Insert(offer)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	userClaim := auth.UserClaim{}.FromCtx(ctx)
	orderofferHistory := offer.NewOrderOfferHistory(userClaim)
	if err := orderofferHistory.Save(ctx); err != nil {
		return err
	}

	return nil
}
func (itemAppliedCartOffer *ItemAppliedCartOffer) Save(ctx context.Context) error {
	if itemAppliedCartOffer.OfferNo == "" || itemAppliedCartOffer.Price < 0 {
		return nil
	}
	row, err := factory.DB(ctx).Insert(itemAppliedCartOffer)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}
	return nil
}
