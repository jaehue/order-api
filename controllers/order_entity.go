package controllers

import (
	"context"
	"errors"
	"strconv"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	"github.com/pangpanglabs/goutils/number"
)

func (c *OrderInput) NewOrderEntity(userClaim auth.UserClaim) (models.Order, error) {
	var order models.Order
	order.TenantCode = userClaim.TenantCode
	order.StoreId = c.StoreId
	order.ChannelId = userClaim.ChannelId
	order.OuterOrderNo = c.OuterOrderNo
	order.SaleType = c.SaleType
	if userClaim.Iss == auth.IssMembership {
		order.CustomerId = userClaim.CustomerId
		order.WxMiniAppOpenId = c.WxMiniAppOpenId
		order.CreatedId = userClaim.CustomerId
	} else if userClaim.Iss == auth.IssColleague {
		order.CustomerId = c.CustomerId
		order.CreatedId = userClaim.ColleagueId
	}
	order.SalesmanId = c.SalesmanId
	order.UserClaimIss = userClaim.Iss
	order.CustRemark = c.CustRemark
	order.ObtainMileage = c.ObtainMileage
	order.Mileage = c.Mileage
	order.MileagePrice = c.MileagePrice
	order.FreightPrice = c.FreightPrice
	order.IsOutPaid = c.IsOutPaid
	order.IsCustDel = false

	err := c.makeOrderItem(&order)
	if err != nil {
		return models.Order{}, err
	}
	err = c.makeOrderOffer(&order)
	if err != nil {
		return models.Order{}, err
	}

	c.makeOrderAddress(&order)

	order.InitializeStatus()
	return order, nil
}

func (c *OrderInput) makeOrderItem(order *models.Order) error {
	orderItems := []models.OrderItem{}
	for _, item := range c.Items {
		orderItem := models.OrderItem{}
		orderItem.TenantCode = order.TenantCode
		orderItem.StoreId = order.StoreId
		orderItem.ChannelId = order.ChannelId
		orderItem.CustomerId = order.CustomerId
		orderItem.OuterOrderNo = order.OuterOrderNo
		orderItem.OuterOrderItemNo = item.OuterOrderItemNo
		orderItem.WxMiniAppOpenId = order.WxMiniAppOpenId
		orderItem.ItemCode = item.ItemCode
		orderItem.ItemName = item.ItemName
		orderItem.ProductId = item.ProductId
		orderItem.SkuId = item.SkuId
		orderItem.SkuImg = item.SkuImg
		orderItem.Option = item.Option
		orderItem.ObtainMileage = item.ObtainMileage
		orderItem.Mileage = item.Mileage
		orderItem.MileagePrice = item.MileagePrice
		orderItem.ListPrice = item.ListPrice
		orderItem.SalePrice = item.SalePrice
		orderItem.Quantity = item.Quantity
		orderItemPrice := models.OrderCalculatePrice(orderItem.ListPrice, orderItem.SalePrice, orderItem.Quantity, float64(0), float64(0))
		orderItem.TotalListPrice = orderItemPrice.TotalListPrice
		orderItem.TotalSalePrice = orderItemPrice.TotalSalePrice
		orderItem.TotalDiscountPrice = orderItemPrice.TotalDiscountPrice
		orderItem.TotalPaymentPrice = orderItemPrice.TotalPaymentPrice
		orderItem.IsDelivery = item.IsDelivery
		orderItem.IsStockChecked = item.IsStockChecked
		orderItem.TotalDistributedCartOfferPrice = item.TotalDistributedCartOfferPrice
		if orderItem.TotalDistributedCartOfferPrice > orderItem.TotalSalePrice {
			return errors.New(orderItem.ItemCode + ":item.TotalDistributedCartOfferPrice=" + strconv.FormatFloat(orderItem.TotalDistributedCartOfferPrice, 'f', 2, 64) +
				",item.TotalSalePrice =" + strconv.FormatFloat(orderItem.TotalSalePrice, 'f', 2, 64))
		}
		distributeOfferPrice := float64(0)
		for _, appliedCartOffer := range item.AppliedCartOffers {
			itemAppliedCartOffer := models.ItemAppliedCartOffer{}
			itemAppliedCartOffer.OfferNo = appliedCartOffer.OfferNo
			itemAppliedCartOffer.Price = appliedCartOffer.DiscountPrice
			itemAppliedCartOffer.IsTarget = appliedCartOffer.IsTarget
			orderItem.AppliedCartOffers = append(orderItem.AppliedCartOffers, itemAppliedCartOffer)
			distributeOfferPrice += appliedCartOffer.DiscountPrice
			if appliedCartOffer.TargetType == "gift" && appliedCartOffer.IsTarget && orderItem.TotalDistributedCartOfferPrice != orderItem.TotalSalePrice {
				return errors.New(orderItem.ItemCode + ":gift item.TotalDistributedCartOfferPrice=" + strconv.FormatFloat(orderItem.TotalDistributedCartOfferPrice, 'f', 2, 64) +
					",item.TotalSalePrice =" + strconv.FormatFloat(orderItem.TotalSalePrice, 'f', 2, 64))
			}
		}
		distributeOfferPrice = number.ToFixed(distributeOfferPrice, nil)
		if distributeOfferPrice != item.TotalDistributedCartOfferPrice {
			return errors.New(item.ItemCode + ":item TotalDistributedCartOfferPrice=" + strconv.FormatFloat(item.TotalDistributedCartOfferPrice, 'f', 2, 64) +
				",Offer applied Items sum price =" + strconv.FormatFloat(distributeOfferPrice, 'f', 2, 64))
		}
		orderResellers := []models.OrderReseller{}
		for _, reseller := range item.Resellers {
			orderReseller := models.OrderReseller{}
			orderReseller.ResellerId = reseller.ResellerId
			orderReseller.ResellerName = reseller.ResellerName
			orderResellers = append(orderResellers, orderReseller)
		}
		orderItem.Resellers = append(orderItem.Resellers, orderResellers...)
		orderItems = append(orderItems, orderItem)
	}

	order.Items = append(order.Items, orderItems...)
	return nil
}

func (c *OrderInput) makeOrderOffer(order *models.Order) error {
	orderOffers := []models.OrderOffer{}
	for _, offer := range c.Offers {
		orderOffer := models.OrderOffer{}
		orderOffer.TenantCode = order.TenantCode
		orderOffer.StoreId = order.StoreId
		orderOffer.ChannelId = order.ChannelId
		orderOffer.CustomerId = order.CustomerId
		orderOffer.OfferNo = offer.OfferNo
		orderOffer.CouponNo = offer.CouponNo
		orderOffer.AdditionalType = offer.AdditionalType
		orderOffer.TargetType = offer.TargetType
		if offer.DiscountPrice < 0 {
			return factory.NewError(api.ErrorMissParameter, "offer.DiscountPrice less than 0")
		}
		for _, offer := range c.Offers {
			distributeOfferPrice := float64(0)
			for i, item := range order.Items {
				for j, appliedCartOffer := range item.AppliedCartOffers {
					if appliedCartOffer.OfferNo == offer.OfferNo {
						distributeOfferPrice += appliedCartOffer.Price
						order.Items[i].AppliedCartOffers[j].TargetType = offer.TargetType
						order.Items[i].AppliedCartOffers[j].AdditionalType = offer.AdditionalType
						order.Items[i].AppliedCartOffers[j].CouponNo = offer.CouponNo
					}
					distributeOfferPrice = number.ToFixed(distributeOfferPrice, nil)
				}
			}
			if distributeOfferPrice != offer.DiscountPrice {
				return errors.New(offer.OfferNo + ":Offer Price=" + strconv.FormatFloat(offer.DiscountPrice, 'f', 2, 64) +
					",Offer applied Items sum price =" + strconv.FormatFloat(distributeOfferPrice, 'f', 2, 64))
			}
		}

		orderOffer.Price = offer.DiscountPrice
		if offer.Description == "" {
			return factory.NewError(api.ErrorMissParameter, "offer.Description is empty")
		}
		orderOffer.Description = offer.Description
		orderOffers = append(orderOffers, orderOffer)
	}
	order.Offers = append(order.Offers, orderOffers...)

	return nil
}
func (c *OrderInput) makeOrderAddress(order *models.Order) {
	if c.DeliverableAddress != nil {
		order.DeliverableAddress.UserName = c.DeliverableAddress.UserName
		order.DeliverableAddress.PostalCode = c.DeliverableAddress.PostalCode
		order.DeliverableAddress.ProvinceName = c.DeliverableAddress.ProvinceName
		order.DeliverableAddress.CityName = c.DeliverableAddress.CityName
		order.DeliverableAddress.CountyName = c.DeliverableAddress.CountyName
		order.DeliverableAddress.DetailInfo = c.DeliverableAddress.DetailInfo
		order.DeliverableAddress.NationalCode = c.DeliverableAddress.NationalCode
		order.DeliverableAddress.TelNumber = c.DeliverableAddress.TelNumber
	}
}

func NewOrderEntity(ctx context.Context, customerId int64, orderId int64, InputEntity interface{}) (bool, models.Order, error) {
	isChange := false
	if orderId == 0 {
		return isChange, models.Order{}, factory.NewError(api.ErrorMissParameter, "order id")
	}
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if customerId == 0 {
		customerId = userClaim.CustomerId
	}
	o, err := models.Order{}.GetOrder(ctx, "", customerId, orderId, nil, "", false)
	if err != nil {
		return isChange, models.Order{}, err
	}
	if o.Id == 0 {
		return isChange, models.Order{}, factory.NewError(api.ErrorMissParameter, "order id")
	}
	if o.IsCustDel == true {
		return isChange, models.Order{}, factory.NewError(api.ErrorInvalidStatus, "Order is Deleted")
	}

	var order models.Order
	order = o

	orderDelete, ok := InputEntity.(OrderDelete)
	if ok {
		if order.IsCustDel == true && orderDelete.IsDelete == false {
			isChange = true
			order.IsCustDel = orderDelete.IsDelete
		} else if order.IsCustDel == false && orderDelete.IsDelete == true {
			isChange = true
			order.IsCustDel = orderDelete.IsDelete
		}
	}

	return isChange, order, nil
}

func SaleOrderStatusEntity(ctx context.Context, orderType enum.OrderType, orderId int64) (*models.Event, error) {
	order, err := models.Order{}.GetOrder(ctx, "", 0, orderId, nil, "", false)
	if err != nil {
		return nil, err
	}
	if order.Id == 0 {
		return nil, factory.NewError(api.ErrorMissParameter, "order id")
	}
	if order.IsCustDel == true {
		return nil, ApiErrorAlreadyDeleted

	}
	if orderType == enum.SaleOrderCancel {
		if order.Status != enum.SaleOrderProcessing.String() {
			return nil, ApiErrorCanNotBeCanceled
		}
	}

	orderEvent := order.MakeOrderEvent(orderType)

	var event models.Event
	event.EntityType = "Order"
	event.Status = orderType.String()
	event.Payload = orderEvent
	event.CreatedAt = time.Now().UTC()

	return &event, nil
}
