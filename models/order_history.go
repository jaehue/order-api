package models

import (
	"context"
	"time"

	"github.com/hublabs/order-api/factory"
)

type OrderHistory struct {
	Id                 int64              `json:"id" query:"id"`
	OrderId            int64              `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	TenantCode         string             `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId            int64              `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId          int64              `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	SaleType           string             `json:"saleType" query:"saleType" xorm:"index VARCHAR(30) default NULL "`
	CustomerId         int64              `json:"customerId" query:"customerId" xorm:"index notnull" validate:"gte=0"`
	OuterOrderNo       string             `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	WxMiniAppOpenId    string             `json:"wxMiniAppOpenId" query:"wxMiniAppOpenId" xorm:"index VARCHAR(100) notnull" `
	SalesmanId         int64              `json:"salesmanId" query:"salesmanId" xorm:"index notnull" validate:"gte=0"`
	TotalListPrice     float64            `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalSalePrice     float64            `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	TotalDiscountPrice float64            `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalPaymentPrice  float64            `json:"totalPaymentPrice" query:"totalPaymentPrice" xorm:"DECIMAL(18,2) notnull" validate:"required"`
	FreightPrice       float64            `json:"freightPrice" query:"freightPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	CashPrice          float64            `json:"cashPrice" query:"cashPrice" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	ObtainMileage      float64            `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage            float64            `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	MileagePrice       float64            `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	Status             string             `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsOutPaid          bool               `json:"isOutPaid" query:"isOutPaid" xorm:"index default false" `
	IsCustDel          bool               `json:"isCustDel"  xorm:"default false" `
	CustRemark         string             `json:"custRemark" query:"custRemark" xorm:"VARCHAR(100)"`
	CreatedAt          time.Time          `json:"createdAt" query:"createdAt" xorm:"created"`
	CreatedId          int64              `json:"createdId" query:"createdId" xorm:"index notnull" validate:"gte=0"`
	UserClaimIss       string             `json:"userClaimIss" query:"userClaimIss" xorm:"VARCHAR(20) notnull" validate:"required"`
	ItemHistory        []OrderItemHistory `json:"orderItemsHistory" query:"orderItemsHistory" xorm:"-" validate:"required,dive,required"`
}

type OrderItemHistory struct {
	Id                 int64     `json:"id" query:"id"`
	OrderId            int64     `json:"orderId" query:"orderId" xorm:"index notnull" validate:"required"`
	OrderItemId        int64     `json:"orderItemId" query:"orderItemId" xorm:"notnull index" validate:""required"`
	TenantCode         string    `json:"tenantCode" query:"tenantCode" xorm:"index VARCHAR(50) notnull" validate:"required"`
	StoreId            int64     `json:"storeId" query:"storeId" xorm:"index notnull" validate:"gte=0"`
	ChannelId          int64     `json:"channelId" query:"channelId" xorm:"index notnull" validate:"gte=0"`
	CustomerId         int64     `json:"customerId" query:"customerId" xorm:"index notnull" validate:"gte=0"`
	OuterOrderNo       string    `json:"outerOrderNo" query:"outerOrderNo" xorm:"index VARCHAR(100) notnull" `
	OuterOrderItemNo   string    `json:"outerOrderItemNo" query:"outerOrderItemNo" xorm:"index VARCHAR(100) notnull" `
	ItemCode           string    `json:"itemCode" query:"itemCode" xorm:"index notnull" validate:"required"`
	ItemName           string    `json:"itemName" query:"itemName" xorm:"notnull" validate:"required"`
	ProductId          int64     `json:"productId" query:"productId" xorm:"index notnull" validate:"gte=0"`
	SkuId              int64     `json:"skuId" query:"skuId" xorm:"notnull" validate:"gte=0"`
	SkuImg             string    `json:"skuImg" query:"skuImg" xorm:"VARCHAR(200)" validate:""`
	Option             string    `json:"option" query:"option" xorm:"VARCHAR(100)" validate:""`
	ObtainMileage      float64   `json:"obtainMileage" query:"obtainMileage" xorm:"DECIMAL(18,2) default 0" validate:"gte=0"`
	Mileage            float64   `json:"mileage" query:"mileage" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	MileagePrice       float64   `json:"mileagePrice" query:"mileagePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	ListPrice          float64   `json:"listPrice" query:"listPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	SalePrice          float64   `json:"salePrice" query:"salePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	ItemFee            float64   `json:"itemFee" query:"itemFee" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	FeeRate            float64   `json:"feeRate" query:"feeRate" xorm:"DECIMAL(18,2) default 0.00" validate:"gte=0"`
	Quantity           int       `json:"quantity" query:"quantity" xorm:"notnull" validate:"required"`
	TotalListPrice     float64   `json:"totalListPrice" query:"totalListPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalSalePrice     float64   `json:"totalSalePrice" query:"totalSalePrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalDiscountPrice float64   `json:"totalDiscountPrice" query:"totalDiscountPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	TotalPaymentPrice  float64   `json:"totalPaymentPrice" query:"totalPaymentPrice" xorm:"DECIMAL(18,2) notnull" validate:"gte=0"`
	Status             string    `json:"status" query:"status" xorm:"index VARCHAR(30) notnull" validate:"required"`
	IsDelivery         bool      `json:"isDelivery" query:"isDelivery" xorm:"index default true" `
	IsStockChecked     bool      `json:"isStockChecked" query:"isStockChecked" xorm:"index default false" `
	CreatedAt          time.Time `json:"createdAt" query:"createdAt" xorm:"created"`
	CreatedId          int64     `json:"createdId" query:"createdId" xorm:"index notnull" validate:"gte=0"`
	UserClaimIss       string    `json:"userClaimIss" query:"userClaimIss" xorm:"VARCHAR(20) notnull" validate:"required"`
}

func (o *Order) NewOrderHistory(userClaim UserClaim) OrderHistory {
	var orderHistory OrderHistory
	orderHistory.OrderId = o.Id
	orderHistory.TenantCode = userClaim.tenantCode()
	orderHistory.StoreId = o.StoreId
	orderHistory.ChannelId = userClaim.channelId()
	orderHistory.SaleType = o.SaleType
	orderHistory.CustomerId = o.CustomerId
	orderHistory.OuterOrderNo = o.OuterOrderNo
	orderHistory.WxMiniAppOpenId = o.WxMiniAppOpenId
	orderHistory.SalesmanId = o.SalesmanId
	orderHistory.CustRemark = o.CustRemark
	orderHistory.FreightPrice = o.FreightPrice
	orderHistory.TotalDiscountPrice = o.TotalDiscountPrice
	orderHistory.TotalListPrice = o.TotalListPrice
	orderHistory.TotalSalePrice = o.TotalSalePrice
	orderHistory.TotalPaymentPrice = o.TotalPaymentPrice
	orderHistory.ObtainMileage = o.ObtainMileage
	orderHistory.Mileage = o.Mileage
	orderHistory.MileagePrice = o.MileagePrice
	orderHistory.CashPrice = o.CashPrice
	orderHistory.Status = o.Status
	orderHistory.IsCustDel = o.IsCustDel
	orderHistory.IsOutPaid = o.IsOutPaid
	orderHistory.CreatedAt = o.UpdatedAt
	if userClaim.isCustomer() {
		orderHistory.CreatedId = userClaim.customerId()
	} else {
		orderHistory.CreatedId = userClaim.ColleagueId
	}
	orderHistory.UserClaimIss = userClaim.Issuer
	orderItemHistorys := []OrderItemHistory{}
	for _, item := range o.Items {
		orderItemHistory := OrderItemHistory{}
		orderItemHistory.OrderId = item.OrderId
		orderItemHistory.OrderItemId = item.Id
		orderItemHistory.TenantCode = item.TenantCode
		orderItemHistory.StoreId = item.StoreId
		orderItemHistory.ChannelId = item.ChannelId
		orderItemHistory.OuterOrderNo = item.OuterOrderNo
		orderItemHistory.OuterOrderItemNo = item.OuterOrderItemNo
		orderItemHistory.CustomerId = item.CustomerId
		orderItemHistory.ItemCode = item.ItemCode
		orderItemHistory.ItemName = item.ItemName
		orderItemHistory.ProductId = item.ProductId
		orderItemHistory.SkuId = item.SkuId
		orderItemHistory.SkuImg = item.SkuImg
		orderItemHistory.Option = item.Option
		orderItemHistory.ObtainMileage = item.ObtainMileage
		orderItemHistory.Mileage = item.Mileage
		orderItemHistory.MileagePrice = item.MileagePrice
		orderItemHistory.ListPrice = item.ListPrice
		orderItemHistory.SalePrice = item.SalePrice
		orderItemHistory.Quantity = item.Quantity
		orderItemHistory.TotalListPrice = item.TotalListPrice
		orderItemHistory.TotalSalePrice = item.TotalSalePrice
		orderItemHistory.TotalDiscountPrice = item.TotalDiscountPrice
		orderItemHistory.TotalPaymentPrice = item.TotalPaymentPrice
		orderItemHistory.ItemFee = item.ItemFee
		orderItemHistory.FeeRate = item.FeeRate
		orderItemHistory.Status = item.Status
		orderItemHistory.IsDelivery = item.IsDelivery
		orderItemHistory.IsStockChecked = item.IsStockChecked
		orderItemHistory.CreatedAt = item.UpdatedAt
		if userClaim.isCustomer() {
			orderItemHistory.CreatedId = userClaim.customerId()
		} else {
			orderItemHistory.CreatedId = userClaim.ColleagueId
		}
		orderItemHistory.UserClaimIss = userClaim.Issuer
		orderItemHistorys = append(orderItemHistorys, orderItemHistory)
	}
	orderHistory.ItemHistory = append(orderHistory.ItemHistory, orderItemHistorys...)

	return orderHistory
}

func (o *OrderHistory) Save(ctx context.Context) error {
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

func (orderHistory *OrderHistory) insert(ctx context.Context) error {
	_, err := factory.DB(ctx).Insert(orderHistory)
	if err != nil {
		return nil
	}

	return nil
}

func (orderItemHistory *OrderItemHistory) insert(ctx context.Context) error {
	if _, err := factory.DB(ctx).Insert(orderItemHistory); err != nil {
		return nil
	}

	return nil
}
