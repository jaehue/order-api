package models

import (
	"context"
	"nomni/utils/auth"
	"time"

	"github.com/hublabs/order-api/factory"
)

type RefundHistory struct {
	Id                 int64               `json:"id" query:"id"`
	RefundId           int64               `json:"refundId" query:"refundId" xorm:"index notnull"`
	TenantCode         string              `json:"tenantCode" query:"tenantCode" xorm:"VARCHAR(50) notnull" validate:"required"`
	StoreId            int64               `json:"storeId" query:"storeId" xorm:"notnull" validate:"gte=0"`
	ChannelId          int64               `json:"channelId" query:"channelId" xorm:"notnull" validate:"gte=0"`
	RefundType         string              `json:"refundType" query:"refundType" xorm:"index VARCHAR(30) default NULL " validate:"required"`
	CustomerId         int64               `json:"customerId" query:"customerId" xorm:"index notnull" validate:"gte=0"`
	OrderId            int64               `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OuterOrderNo       string              `json:"outerOrderNo" query:"outerOrderNo" xorm:"VARCHAR(100) notnull" `
	TotalListPrice     float64             `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalSalePrice     float64             `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalDiscountPrice float64             `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalRefundPrice   float64             `json:"totalRefundPrice" query:"totalRefundPrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	FreightPrice       float64             `json:"freightPrice" query:"freightPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	ObtainMileage      float64             `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage            float64             `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	MileagePrice       float64             `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	CashPrice          float64             `json:"cashPrice" query:"cashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Status             string              `json:"status" query:"status" xorm:"VARCHAR(30) notnull" validate:"required"`
	RefundReason       string              `json:"reason" query:"reason" xorm:"VARCHAR(100)"`
	CustRemark         string              `json:"custRemark" query:"custRemark" xorm:"VARCHAR(500)"`
	IsOutPaid          bool                `json:"isOutPaid" query:"isOutPaid" xorm:"default false" `
	CreatedAt          time.Time           `json:"createdAt" query:"createdAt" xorm:"created"`
	CreatedId          int64               `json:"createdId" query:"createdId" xorm:"notnull" validate:"gte=0"`
	UserClaimIss       string              `json:"userClaimIss" query:"userClaimIss" xorm:"VARCHAR(20) notnull" validate:"required"`
	ItemHistory        []RefundItemHistory `json:"refundItemHistory" query:"refundItemHistory" xorm:"-" validate:"required,dive,required"`
}

type RefundItemHistory struct {
	Id                      int64     `json:"id" query:"id"`
	RefundId                int64     `json:"refundId" query:"refundId" xorm:"index notnull"`
	RefundItemId            int64     `json:"refundItemId" query:"refundItemId" xorm:"index notnull" validate:"required"`
	TenantCode              string    `json:"tenantCode" query:"tenantCode" xorm:"VARCHAR(50) notnull" validate:"required"`
	StoreId                 int64     `json:"storeId" query:"storeId" xorm:"default 0" validate:"gte=0"`
	ChannelId               int64     `json:"channelId" query:"channelId" xorm:"default 0" validate:"gte=0"`
	CustomerId              int64     `json:"customerId" query:"customerId" xorm:"default 0" validate:"gte=0"`
	OrderId                 int64     `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OrderItemId             int64     `json:"orderItemId" query:"orderItemId" xorm:"index notnull" validate:"required"`
	SeparateId              int64     `json:"separateId" query:"separateId" xorm:"default 0" `
	StockDistributionItemId int64     `json:"stockDistributionItemId" query:"stockDistributionItemId" xorm:"default 0"`
	OuterOrderNo            string    `json:"outerOrderNo" query:"outerOrderNo" xorm:"VARCHAR(100) notnull" `
	OuterOrderItemNo        string    `json:"outerOrderItemNo" query:"outerOrderItemNo" xorm:"VARCHAR(100) notnull" `
	ItemCode                string    `json:"itemCode" query:"itemCode" xorm:"index notnull" validate:"required"`
	ItemName                string    `json:"itemName" query:"itemName" xorm:"notnull" validate:"required"`
	ProductId               int64     `json:"productId" query:"productId" xorm:"index notnull" validate:"gte=0"`
	SkuId                   int64     `json:"skuId" query:"skuId" xorm:"notnull" validate:"gte=0"`
	SkuImg                  string    `json:"skuImg" query:"skuImg" xorm:"VARCHAR(200)" validate:""`
	Option                  string    `json:"option" query:"option" xorm:"VARCHAR(100)" validate:""`
	ObtainMileage           float64   `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage                 float64   `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	MileagePrice            float64   `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	ListPrice               float64   `json:"listPrice" query:"listPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	SalePrice               float64   `json:"salePrice" query:"salePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	ItemFee                 float64   `json:"itemFee" query:"itemFee" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	FeeRate                 float64   `json:"feeRate" query:"feeRate" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	Quantity                int       `json:"quantity" query:"quantity" xorm:"notnull" validate:"required"`
	TotalListPrice          float64   `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalSalePrice          float64   `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalDiscountPrice      float64   `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalRefundPrice        float64   `json:"totalRefundPrice" query:"totalRefundPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	Status                  string    `json:"status" query:"status" xorm:"VARCHAR(30) notnull" validate:"required"`
	IsDelivery              bool      `json:"isDelivery" query:"isDelivery" xorm:"index default true" `
	CreatedAt               time.Time `json:"createdAt" query:"createdAt" xorm:"created"`
	CreatedId               int64     `json:"createdId" query:"createdId" xorm:"notnull" validate:"gte=0"`
	UserClaimIss            string    `json:"userClaimIss" query:"userClaimIss" xorm:"VARCHAR(20) notnull" validate:"required"`
}

func (o *Refund) NewRefundHistory(userClaim auth.UserClaim) RefundHistory {
	var refundHistory RefundHistory
	refundHistory.RefundId = o.Id
	refundHistory.OrderId = o.OrderId
	refundHistory.OuterOrderNo = o.OuterOrderNo
	refundHistory.TenantCode = userClaim.TenantCode
	refundHistory.StoreId = o.StoreId
	refundHistory.ChannelId = userClaim.ChannelId
	refundHistory.RefundType = o.RefundType
	refundHistory.CustomerId = o.CustomerId
	refundHistory.CustRemark = o.CustRemark
	refundHistory.RefundReason = o.RefundReason
	refundHistory.FreightPrice = o.FreightPrice
	refundHistory.ObtainMileage = o.ObtainMileage
	refundHistory.Mileage = o.Mileage
	refundHistory.MileagePrice = o.MileagePrice
	refundHistory.CashPrice = o.CashPrice
	refundHistory.TotalListPrice = o.TotalSalePrice
	refundHistory.TotalSalePrice = o.TotalSalePrice
	refundHistory.TotalDiscountPrice = o.TotalDiscountPrice
	refundHistory.TotalRefundPrice = o.TotalRefundPrice
	refundHistory.Status = o.Status
	refundHistory.IsOutPaid = o.IsOutPaid
	refundHistory.CreatedAt = o.UpdatedAt
	if userClaim.Iss == auth.IssMembership {
		refundHistory.CreatedId = userClaim.CustomerId
	} else {
		refundHistory.CreatedId = userClaim.ColleagueId
	}
	refundHistory.UserClaimIss = userClaim.Iss
	refundItemHistorys := []RefundItemHistory{}
	for _, item := range o.Items {
		refundItemHistory := RefundItemHistory{}
		refundItemHistory.OrderId = item.OrderId
		refundItemHistory.OrderItemId = item.OrderItemId
		refundItemHistory.SeparateId = item.SeparateId
		refundItemHistory.StockDistributionItemId = item.StockDistributionItemId
		refundItemHistory.OuterOrderNo = item.OuterOrderNo
		refundItemHistory.OuterOrderItemNo = item.OuterOrderItemNo
		refundItemHistory.TenantCode = item.TenantCode
		refundItemHistory.StoreId = item.StoreId
		refundItemHistory.ChannelId = item.ChannelId
		refundItemHistory.CustomerId = item.CustomerId
		refundItemHistory.RefundId = item.RefundId
		refundItemHistory.RefundItemId = item.Id
		refundItemHistory.ItemCode = item.ItemCode
		refundItemHistory.ItemName = item.ItemName
		refundItemHistory.ProductId = item.ProductId
		refundItemHistory.SkuId = item.SkuId
		refundItemHistory.SkuImg = item.SkuImg
		refundItemHistory.Option = item.Option
		refundItemHistory.ObtainMileage = item.ObtainMileage
		refundItemHistory.Mileage = item.Mileage
		refundItemHistory.MileagePrice = item.MileagePrice
		refundItemHistory.ListPrice = item.ListPrice
		refundItemHistory.SalePrice = item.SalePrice
		refundItemHistory.Quantity = item.Quantity
		refundItemHistory.TotalListPrice = item.TotalSalePrice
		refundItemHistory.TotalSalePrice = item.TotalSalePrice
		refundItemHistory.TotalDiscountPrice = item.TotalDiscountPrice
		refundItemHistory.TotalRefundPrice = item.TotalRefundPrice
		refundItemHistory.ItemFee = item.ItemFee
		refundItemHistory.FeeRate = item.FeeRate
		refundItemHistory.Status = item.Status
		refundItemHistory.IsDelivery = item.IsDelivery
		refundItemHistory.CreatedAt = item.UpdatedAt
		if userClaim.Iss == auth.IssMembership {
			refundItemHistory.CreatedId = userClaim.CustomerId
		} else {
			refundItemHistory.CreatedId = userClaim.ColleagueId
		}
		refundItemHistory.UserClaimIss = userClaim.Iss
		refundItemHistorys = append(refundItemHistorys, refundItemHistory)
	}
	refundHistory.ItemHistory = append(refundHistory.ItemHistory, refundItemHistorys...)

	return refundHistory
}

func (o *RefundHistory) Save(ctx context.Context) error {
	if err := o.insert(ctx); err != nil {
		return err
	}

	for i := range o.ItemHistory {
		if err := o.ItemHistory[i].insert(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (refundHistory *RefundHistory) insert(ctx context.Context) error {
	_, err := factory.DB(ctx).Insert(refundHistory)
	if err != nil {
		return nil
	}

	return nil
}

func (refundItemHistory *RefundItemHistory) insert(ctx context.Context) error {
	if _, err := factory.DB(ctx).Insert(refundItemHistory); err != nil {
		return nil
	}

	return nil
}
