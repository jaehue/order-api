package controllers

import "github.com/hublabs/order-api/models"

type OrderInput struct {
	IsOutPaid          bool                       `json:"isOutPaid" query:"isOutPaid" `
	SaleType           string                     `json:"saleType" query:"saleType"  validate:"required"`
	OuterOrderNo       string                     `json:"outerOrderNo" query:"outerOrderNo" `
	ObtainMileage      float64                    `json:"obtainMileage" query:"obtainMileage" validate:"gte=0"`
	Mileage            float64                    `json:"mileage" query:"mileage" `
	MileagePrice       float64                    `json:"mileagePrice" query:"mileagePrice" `
	FreightPrice       float64                    `json:"freightPrice" query:"freightPrice" validate:"gte=0"`
	CustRemark         string                     `json:"custRemark" query:"custRemark" `
	CustomerId         int64                      `json:"customerId" query:"customerId" `
	StoreId            int64                      `json:"storeId" query:"storeId" `
	SalesmanId         int64                      `json:"salesmanId" query:"salesmanId" `
	WxMiniAppOpenId    string                     `json:"wxMiniAppOpenId" query:"wxMiniAppOpenId" `
	Offers             []OfferInput               `json:"offers" query:"offers"`
	Items              []ItemInput                `json:"items" query:"items" validate:"required,dive,required"`
	DeliverableAddress *models.DeliverableAddress `json:"deliverableAddress" query:"deliverableAddress" `
}

type OrderDelete struct {
	CustomerId int64 `json:"customerId" query:"customerId" `
	IsDelete   bool  `json:"isDelete" query:"isDelete" `
}

type ItemInput struct {
	OuterOrderItemNo               string                      `json:"outerOrderItemNo" query:"outerOrderItemNo" `
	ItemCode                       string                      `json:"itemCode" query:"itemCode" validate:"required"`
	ItemName                       string                      `json:"itemName" query:"itemName" validate:"required"`
	ProductId                      int64                       `json:"productId" query:"productId" validate:"gte=0"`
	SkuId                          int64                       `json:"skuId" query:"skuId" validate:"gte=0"`
	SkuImg                         string                      `json:"skuImg" query:"skuImg" validate:""`
	ObtainMileage                  float64                     `json:"obtainMileage" query:"obtainMileage" validate:"gte=0"`
	Mileage                        float64                     `json:"mileage" query:"mileage" `
	MileagePrice                   float64                     `json:"mileagePrice" query:"mileagePrice" `
	Option                         string                      `json:"option" query:"option" validate:""`
	ListPrice                      float64                     `json:"listPrice" query:"listPrice" validate:"gte=0"`
	SalePrice                      float64                     `json:"salePrice" query:"salePrice" validate:"gte=0"`
	TotalDistributedCartOfferPrice float64                     `json:"totalDistributedCartOfferPrice" query:"totalDistributedCartOfferPrice" validate:"gte=0"`
	Quantity                       int                         `json:"quantity" query:"quantity" validate:"gt=0"`
	IsDelivery                     bool                        `json:"isDelivery" query:"isDelivery"`
	IsStockChecked                 bool                        `json:"isStockChecked" query:"isStockChecked" `
	AppliedCartOffers              []ItemAppliedCartOfferInput `json:"appliedCartOffers" query:"appliedCartOffers"`
	Resellers                      []ResellerInput             `json:"resellers" query:"resellers"`
}

type ResellerInput struct {
	ResellerId   int64  `json:"resellerId" query:"resellerId" xorm:"VARCHAR(100)" validate:"required"`
	ResellerName string `json:"resellerName" query:"resellerName" xorm:"VARCHAR(100)" `
}

type RefundInput struct {
	OrderId            int64                      `json:"orderId" query:"orderId" validate:"required"`
	RefundType         string                     `json:"refundType" query:"refundType" validate:"required"`
	CustomerId         int64                      `json:"customerId" query:"customerId"`
	SalesmanId         int64                      `json:"salesmanId" query:"salesmanId" `
	FreightPrice       float64                    `json:"freightPrice" query:"freightPrice" validate:"gte=0"`
	CustRemark         string                     `json:"custRemark" query:"custRemark"`
	RefundReason       string                     `json:"refundReason" query:"refundReason" validate:"required"`
	RefundOrderItems   []RefundOrderItemInput     `json:"items" query:"items" validate:"required,dive,required"`
	DeliverableAddress *models.DeliverableAddress `json:"deliverableAddress" query:"deliverableAddress"`
	Extension          RefundExtension            `json:"extension" query:"extension"`
}

type RefundOrderItemInput struct {
	OrderItemId         int64                     `json:"id" query:"id"`
	Quantity            int                       `json:"quantity" query:"quantity"`
	IsDelivery          bool                      `json:"isDelivery" query:"isDelivery"`
	RefundItemSeparates []RefundOrderItemSeparate `json:"itemSeparates" query:"itemSeparate"`
}

type RefundExtension struct {
	ImgUrl string `json:"imgUrl" query:"imgUrl"`
}

type RefundOrderItemSeparate struct {
	SeparateId int64 `json:"separateId" query:"separateId" validate:"required"`
	Quantity   int   `json:"quantity" query:"quantity" xorm:"notnull" validate:"gte=0"`
	IsDelivery bool  `json:"isDelivery" query:"isDelivery"`
}

type OfferInput struct {
	OfferNo        string  `json:"offerNo" query:"offerNo" validate:"required"`
	CouponNo       string  `json:"couponNo" query:"couponNo" `
	TargetType     string  `json:"targetType"`
	AdditionalType string  `json:"additionalType"`
	DiscountPrice  float64 `json:"discountPrice" query:"discountPrice" validate:"required"`
	Description    string  `json:"description" query:"description" validate:"required"`
}
type ItemAppliedCartOfferInput struct {
	OfferNo       string  `json:"offerNo" query:"offerNo" validate:"required"`
	DiscountPrice float64 `json:"discountPrice" query:"discountPrice" validate:"required"`
	TargetType    string  `json:"targetType"`
	IsTarget      bool    `json:"isTarget"`
}
type RefundStatusInput struct {
	RefundId          int64                     `json:"id" query:"id" validate:"required"`
	RefuseReason      string                    `json:"refuseReason" query:"refuseReason" `
	RefundItemsStatus []RefundItemStatusInput   `json:"items" query:"items" validate:"dive,required"`
	Address           models.DeliverableAddress `json:"address" query:"address" `
}

type RefundItemStatusInput struct {
	RefundItemId int64 `json:"id" query:"id" validate:"required"`
	IsDelivery   bool  `json:"isDelivery" query:"isDelivery"`
}
