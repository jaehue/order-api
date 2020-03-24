package controllers

import (
	"context"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	"github.com/pangpanglabs/goutils/number"
	"github.com/sirupsen/logrus"
)

func (c *RefundInput) PartialRefund(ctx context.Context, o models.Order) ([]models.OrderItemSeparate, []models.OrderItemSeparate, error) {
	if o.Id == 0 {
		return nil, nil, factory.NewError(api.ErrorMissParameter, "Refund order infomation")
	}
	if o.Status == enum.SaleOrderProcessing.String() || o.Status == enum.SaleOrderCancel.String() {
		return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Refund Create Status not Allowed :"+o.Status)
	}
	newOrderItemSeparates := []models.OrderItemSeparate{}
	oldOrderItemSeparates := []models.OrderItemSeparate{}
	for k, item := range o.Items {
		for r, refundOrderItem := range c.RefundOrderItems {
			if item.Id == refundOrderItem.OrderItemId {
				if len(item.ItemSeparates) == 0 {
					if item.Status == enum.SaleShippingWaiting.String() || item.Status == enum.SaleShippingProcessing.String() {
						return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Refund Create Status not Allowed :"+item.Status)
					}
					if item.Quantity == refundOrderItem.Quantity || refundOrderItem.Quantity == 0 {
						continue
					} else if item.Quantity < refundOrderItem.Quantity {
						return nil, nil, factory.NewError(api.ErrorMissParameter, "refund item Quantity better then Origin Order item Quantity")
					} else {
						separateOrder := models.SeparateOrder{}
						separateOrderItem := models.SeparateOrderItem{}
						separateOrder.OrderId = o.Id
						separateOrderItem.OrderItemId = item.Id
						for i := 0; i < 2; i++ {
							newItemSeparate := models.ItemSeparate{}
							if i == 0 {
								newItemSeparate.Quantity = refundOrderItem.Quantity
							} else {
								newItemSeparate.Quantity = item.Quantity - refundOrderItem.Quantity
							}
							separateOrderItem.ItemSeparates = append(separateOrderItem.ItemSeparates, newItemSeparate)
						}
						separateOrder.Items = append(separateOrder.Items, separateOrderItem)

						resultOrderItemSeparates, err := separateOrder.OrderItemSparateEvent(ctx, true)
						if err != nil {
							return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Refund Create")
						}

						newRefundOrderItemInput := RefundOrderItemInput{}
						newRefundOrderItemInput.OrderItemId = refundOrderItem.OrderItemId
						newRefundOrderItemInput.IsDelivery = refundOrderItem.IsDelivery
						RefundItemSeparateCount := 0
						for _, itemSeperate := range resultOrderItemSeparates {
							newOrderItemSeparates = append(newOrderItemSeparates, itemSeperate)
							o.Items[k].ItemSeparates = append(o.Items[k].ItemSeparates, itemSeperate)
							if itemSeperate.Quantity == refundOrderItem.Quantity && RefundItemSeparateCount == 0 {
								RefundItemSeparateCount++
								newRefundOrderItemSeparate := RefundOrderItemSeparate{}
								newRefundOrderItemSeparate.SeparateId = itemSeperate.Id
								newRefundOrderItemSeparate.Quantity = refundOrderItem.Quantity
								newRefundOrderItemSeparate.IsDelivery = refundOrderItem.IsDelivery
								newRefundOrderItemInput.RefundItemSeparates = append(newRefundOrderItemInput.RefundItemSeparates, newRefundOrderItemSeparate)
							}
						}
						c.RefundOrderItems[r] = newRefundOrderItemInput
					}
				} else {
					for j, itemSeparate := range item.ItemSeparates {
						for l, refundItemSeparate := range refundOrderItem.RefundItemSeparates {
							if itemSeparate.Id == refundItemSeparate.SeparateId {
								if itemSeparate.Status == enum.SaleShippingWaiting.String() || itemSeparate.Status == enum.SaleShippingProcessing.String() {
									return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Refund Create Status not Allowed :"+itemSeparate.Status)
								}
								if itemSeparate.Quantity == refundItemSeparate.Quantity || refundItemSeparate.Quantity == 0 {
									continue
								} else if itemSeparate.Quantity < refundItemSeparate.Quantity {
									return nil, nil, factory.NewError(api.ErrorMissParameter, "refund item Quantity better then Origin Order item Quantity")
								} else {
									separateOrder := models.SeparateOrder{}
									separateOrderItem := models.SeparateOrderItem{}
									separateOrder.OrderId = o.Id
									separateOrderItem.OrderItemId = item.Id
									for i := 0; i < 3; i++ {
										newItemSeparate := models.ItemSeparate{}
										if i == 0 {
											newItemSeparate.IsDelete = true
											newItemSeparate.Quantity = itemSeparate.Quantity
										} else if i == 1 {
											newItemSeparate.IsDelete = false
											newItemSeparate.Quantity = refundItemSeparate.Quantity
										} else if i == 2 {
											newItemSeparate.IsDelete = false
											newItemSeparate.Quantity = itemSeparate.Quantity - refundItemSeparate.Quantity
										}
										newItemSeparate.StockDistributionItemId = itemSeparate.StockDistributionItemId
										separateOrderItem.ItemSeparates = append(separateOrderItem.ItemSeparates, newItemSeparate)
									}
									separateOrder.Items = append(separateOrder.Items, separateOrderItem)

									resultOrderItemSeparates, err := separateOrder.OrderItemSparateEvent(ctx, true)
									if err != nil {
										return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Refund Create tatus not Allowed :"+item.Status)
									}

									oldOrderItemSeparates = append(oldOrderItemSeparates, o.Items[k].ItemSeparates[j])
									o.Items[k].ItemSeparates[j] = models.OrderItemSeparate{}
									for _, itemSeperate := range resultOrderItemSeparates {
										newOrderItemSeparates = append(newOrderItemSeparates, itemSeperate)
										o.Items[k].ItemSeparates = append(o.Items[k].ItemSeparates, itemSeperate)
										if itemSeperate.Quantity == refundItemSeparate.Quantity {
											refundOrderItem.RefundItemSeparates[l].SeparateId = itemSeperate.Id
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return oldOrderItemSeparates, newOrderItemSeparates, nil
}

func OrderItemSeparatesDelete(c context.Context, oldItemSeparates []models.OrderItemSeparate, newItemSeparates []models.OrderItemSeparate) {
	if len(oldItemSeparates) > 0 {
		if err := models.OrderItemSeparatesDelete(c, oldItemSeparates, false); err != nil {
			logrus.WithError(err).WithField("orderItemSeparates", oldItemSeparates)
		}
	}
	if len(newItemSeparates) > 0 {
		if err := models.OrderItemSeparatesDelete(c, newItemSeparates, true); err != nil {
			logrus.WithError(err).WithField("orderItemSeparates", newItemSeparates)
		}
	}
}

func (c *RefundInput) NewRefundEntity(userClaim auth.UserClaim, o models.Order) (models.Refund, error) {
	if o.Id == 0 {
		return models.Refund{}, factory.NewError(api.ErrorMissParameter, "order id")
	}
	if o.Status == enum.SaleOrderProcessing.String() || o.Status == enum.SaleOrderCancel.String() {
		return models.Refund{}, factory.NewError(api.ErrorInvalidStatus, "Refund Create tatus not Allowed :"+o.Status)
	}
	var refund models.Refund
	if userClaim.TenantCode != o.TenantCode {
		return models.Refund{}, api.ErrorPermissionDenied.New(nil)
	}
	if userClaim.Iss == auth.IssMembership {
		if userClaim.CustomerId != o.CustomerId {
			return models.Refund{}, api.ErrorPermissionDenied.New(nil)
		}
		refund.CreatedId = userClaim.CustomerId
	} else {
		refund.CreatedId = userClaim.ColleagueId
	}
	refund.TenantCode = userClaim.TenantCode
	refund.StoreId = o.StoreId
	refund.ChannelId = userClaim.ChannelId
	refund.RefundType = c.RefundType
	refund.CustomerId = o.CustomerId
	if c.SalesmanId != 0 {
		refund.SalesmanId = c.SalesmanId
	} else {
		refund.SalesmanId = o.SalesmanId
	}
	refund.OrderId = o.Id
	refund.OuterOrderNo = o.OuterOrderNo
	refund.RefundReason = c.RefundReason
	refund.CustRemark = c.CustRemark
	// refund.FreightPrice = c.FreightPrice
	refund.Status = enum.RefundOrderRegistered.String()
	refund.IsOutPaid = o.IsOutPaid
	extension := models.RefundExtension{}
	extension.ImgUrl = c.Extension.ImgUrl
	refund.Extension = extension

	err := c.makeRefundItem(&refund, o)
	if err != nil {
		return models.Refund{}, err
	}

	c.makeRefundAddress(&refund)

	refundPrice := models.RefundCalculatePrice(refund.TotalListPrice, refund.TotalSalePrice, int(1), refund.FreightPrice, float64(0))
	refund.TotalListPrice = refundPrice.TotalListPrice
	refund.TotalSalePrice = refundPrice.TotalSalePrice
	refund.TotalDiscountPrice = refundPrice.TotalDiscountPrice
	refund.TotalRefundPrice = refundPrice.TotalRefundPrice

	return refund, nil
}

func (c *RefundInput) makeRefundItem(refund *models.Refund, o models.Order) error {
	refundItems := []models.RefundItem{}
	for _, refundOrderItem := range c.RefundOrderItems {
		refundItem := models.RefundItem{}
		for _, item := range o.Items {
			if item.Id == refundOrderItem.OrderItemId {

				if len(refundOrderItem.RefundItemSeparates) > 0 && len(item.ItemSeparates) == 0 {
					return factory.NewError(api.ErrorMissParameter, "order item Not Separated")
				}
				if len(refundOrderItem.RefundItemSeparates) == 0 && len(item.ItemSeparates) > 0 {
					return factory.NewError(api.ErrorMissParameter, "order item Separated, refund order item need Separate parameter")
				}
				if len(refundOrderItem.RefundItemSeparates) > len(item.ItemSeparates) {
					return factory.NewError(api.ErrorMissParameter, "refund order item Separate parameter too many")
				}
				if item.Quantity < refundOrderItem.Quantity {
					return factory.NewError(api.ErrorMissParameter, "refund order item Quantity better then Origin order item Quantity")
				}
				refundItem.OrderItemId = refundOrderItem.OrderItemId
				refundItem.IsDelivery = refundOrderItem.IsDelivery
				refundItem.TenantCode = refund.TenantCode
				refundItem.StoreId = item.StoreId
				refundItem.ChannelId = refund.ChannelId
				refundItem.CustomerId = item.CustomerId
				refundItem.OrderId = item.OrderId
				refundItem.OuterOrderNo = refund.OuterOrderNo
				refundItem.OuterOrderItemNo = item.OuterOrderItemNo
				refundItem.ItemCode = item.ItemCode
				refundItem.ProductId = item.ProductId
				refundItem.ItemName = item.ItemName
				refundItem.ProductId = item.ProductId
				refundItem.SkuId = item.SkuId
				refundItem.SkuImg = item.SkuImg
				refundItem.Option = item.Option
				refundItem.ListPrice = item.ListPrice
				refundItem.SalePrice = item.SalePrice
				refundItem.ItemFee = item.ItemFee
				refundItem.FeeRate = item.FeeRate
				if len(item.ItemSeparates) > 0 {
					for _, refundItemSeparate := range refundOrderItem.RefundItemSeparates {
						for _, itemSeparate := range item.ItemSeparates {
							if refundItemSeparate.SeparateId == itemSeparate.Id {
								if itemSeparate.Status == enum.SaleShippingWaiting.String() || itemSeparate.Status == enum.SaleShippingProcessing.String() {
									return factory.NewError(api.ErrorInvalidStatus, "Refund Create tatus not Allowed :"+itemSeparate.Status)
								}
								refundItem.SeparateId = itemSeparate.Id
								refundItem.StockDistributionItemId = itemSeparate.StockDistributionItemId
								refundItem.Quantity = itemSeparate.Quantity
								refundItem.TotalListPrice = itemSeparate.TotalListPrice
								refundItem.TotalSalePrice = itemSeparate.TotalSalePrice
								refundItem.TotalDiscountPrice = itemSeparate.TotalDiscountPrice
								refundItem.TotalRefundPrice = itemSeparate.TotalPaymentPrice
								refundItem.Status = refund.Status
								refundItems = append(refundItems, refundItem)
								refund.TotalListPrice += number.ToFixed(refundItem.TotalListPrice, nil)
								refund.TotalSalePrice += number.ToFixed(refundItem.TotalSalePrice, nil)
								refund.TotalDiscountPrice += number.ToFixed(refundItem.TotalDiscountPrice, nil)
								refund.TotalRefundPrice += number.ToFixed(refundItem.TotalRefundPrice, nil)
							}
						}
					}
				} else {
					if item.Status == enum.SaleShippingWaiting.String() || item.Status == enum.SaleShippingProcessing.String() {
						return factory.NewError(api.ErrorInvalidStatus, "Refund Create tatus not Allowed :"+item.Status)
					}
					refundItem.TotalDiscountPrice = item.TotalDiscountPrice
					refundItem.Quantity = item.Quantity
					refundItem.TotalListPrice = item.TotalListPrice
					refundItem.TotalSalePrice = item.TotalSalePrice
					refundItem.Status = refund.Status
					refundItem.TotalRefundPrice = item.TotalPaymentPrice
					refundItems = append(refundItems, refundItem)
					refund.TotalListPrice += number.ToFixed(refundItem.TotalListPrice, nil)
					refund.TotalSalePrice += number.ToFixed(refundItem.TotalSalePrice, nil)
					refund.TotalDiscountPrice += number.ToFixed(refundItem.TotalDiscountPrice, nil)
					refund.TotalRefundPrice += number.ToFixed(refundItem.TotalRefundPrice, nil)
				}
			}
		}
	}

	if len(refundItems) == 0 {
		return factory.NewError(api.ErrorInvalidStatus, "Already Refund Created Item")
	} else {
		refund.Items = append(refund.Items, refundItems...)
	}

	return nil
}

func (c *RefundInput) makeRefundAddress(refund *models.Refund) {
	if c.DeliverableAddress != nil {
		refund.DeliverableAddress.UserName = c.DeliverableAddress.UserName
		refund.DeliverableAddress.PostalCode = c.DeliverableAddress.PostalCode
		refund.DeliverableAddress.ProvinceName = c.DeliverableAddress.ProvinceName
		refund.DeliverableAddress.CityName = c.DeliverableAddress.CityName
		refund.DeliverableAddress.CountyName = c.DeliverableAddress.CountyName
		refund.DeliverableAddress.DetailInfo = c.DeliverableAddress.DetailInfo
		refund.DeliverableAddress.NationalCode = c.DeliverableAddress.NationalCode
		refund.DeliverableAddress.TelNumber = c.DeliverableAddress.TelNumber
	}
}

func (c RefundStatusInput) RefundOrderStatusEntity(ctx context.Context, orderType enum.OrderType) (models.Event, error) {
	refundId := c.RefundId
	refundResult, err := models.Refund{}.GetRefund(ctx, "", 0, refundId, 0, nil, "", true)
	if err != nil {
		return models.Event{}, factory.NewError(api.ErrorMissParameter, "Request Refund is nothing")
	}
	refundResult.DeliverableAddress = c.Address
	refundStatus := refundResult.Status
	//判断当前订单状态是否允许
	flag := true
	switch orderType {
	case enum.RefundOrderProcessing:
		if refundStatus != enum.RefundOrderRegistered.String() {
			flag = false
		}
	case enum.RefundOrderCancel:
		if refundStatus != enum.RefundOrderRegistered.String() && refundStatus != enum.RefundOrderProcessing.String() {
			flag = false
		}
	case enum.RefundRequisiteApprovals:
		if refundStatus == enum.RefundOrderRegistered.String() || refundStatus == enum.RefundRequisiteApprovals.String() || refundStatus == enum.RefundOrderSuccess.String() {
			flag = false
		}
	}
	if !flag {
		return models.Event{}, factory.NewError(api.ErrorMissParameter, "refund status : "+refundStatus)
	}

	orderEvent := refundResult.MakeRefundOrderEvent(ctx, orderType.String())

	var event models.Event
	event.EntityType = "Refund"
	event.Status = orderType.String()
	event.Payload = orderEvent
	event.CreatedAt = time.Now().UTC()

	return event, nil
}
