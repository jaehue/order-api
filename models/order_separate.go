package models

import (
	"context"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"

	"github.com/pangpanglabs/goutils/number"
	"github.com/sirupsen/logrus"
)

type SeparateOrder struct {
	OrderId int64               `json:"orderId" query:"orderId" validate:"required"`
	Items   []SeparateOrderItem `json:"items" query:"items" validate:"required,dive,required"`
}

type SeparateOrderItem struct {
	OrderItemId   int64          `json:"itemid" query:"itemid" validate:"required"`
	ItemSeparates []ItemSeparate `json:"separates" query:"separates" validate:"required,dive,required"`
}

type ItemSeparate struct {
	SeparateId              int64 `json:"separateId,omitempty"`
	StockDistributionItemId int64 `json:"stockDistributionItemId,omitempty"`
	Quantity                int   `json:"quantity" query:"quantity" xorm:"notnull" validate:"gt=0"`
	IsDelete                bool  `json:"isDelete" query:"isDelete"`
}

func (o *SeparateOrder) OrderItemSparateEvent(ctx context.Context, isRefund bool) ([]OrderItemSeparate, error) {
	beforeItemSeparates, orderItemSeparates, err := o.OrderItemSparate(ctx, isRefund)
	if err != nil {
		return nil, err
	}

	resultOrderItemSeparates, err := OrderItemSeparateUpdate(ctx, beforeItemSeparates, orderItemSeparates)
	if err != nil {
		return nil, err
	}

	return resultOrderItemSeparates, nil
}

func (o *SeparateOrder) OrderItemSparate(ctx context.Context, isRefund bool) ([]OrderItemSeparate, []OrderItemSeparate, error) {
	order, err := Order{}.GetOrder(ctx, "", 0, o.OrderId, nil, "", false)
	if err != nil {
		return nil, nil, err
	}
	if order.Id == 0 {
		return nil, nil, factory.NewError(api.ErrorMissParameter, "Order Separate Order Id or Item Id")
	}

	var orderItemSeparates []OrderItemSeparate
	var removedItemSeparates []OrderItemSeparate
	for _, oItem := range o.Items {
		isItemExist := false
		var beforeItemSeparateQuantity int
		var beforeItemSeparateStatus string
		var totalQuantity int

		for _, item := range order.Items {
			if item.Id == oItem.OrderItemId {
				beforeItemSeparates := []OrderItemSeparate{}
				for _, itemSeparate := range item.ItemSeparates {
					if itemSeparate.RefundItem == nil {
						itemSeparatebyId, err := OrderItemSeparate{}.GetOrderItemSeparate(ctx, itemSeparate.Id)
						if err != nil {
							return nil, nil, err
						}
						beforeItemSeparates = append(beforeItemSeparates, itemSeparatebyId)
					}
				}

				for _, beforeItemSeparate := range beforeItemSeparates {
					if isRefund == false {
						if beforeItemSeparate.Status != enum.SaleOrderFinished.String() && beforeItemSeparate.Status != enum.StockDistributed.String() {
							return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Create Order Separate Status not Allowed : "+item.Status)
						}
					}
					if isRefund == true {
						if beforeItemSeparate.Status == enum.SaleShippingWaiting.String() || beforeItemSeparate.Status == enum.SaleShippingProcessing.String() {
							return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Create Order Separate Status not Allowed : "+item.Status)
						}
					}
				}

				isItemExist = true
				beforeItemSeparateQuantity = 0
				beforeItemSeparateStatus = ""
				logrus.WithField("ItemSeparates", oItem.ItemSeparates).Info("oItem")
				for _, itemSeparate := range oItem.ItemSeparates {
					if itemSeparate.IsDelete == true {
						if beforeItemSeparateQuantity > 0 {
							return nil, nil, factory.NewError(api.ErrorMissParameter, "Create Order Same Separate Delete Item over then 1")
						}
						for _, beforeItemSeparate := range beforeItemSeparates {
							if beforeItemSeparateQuantity == 0 {
								if itemSeparate.SeparateId > 0 || itemSeparate.StockDistributionItemId > 0 {
									if itemSeparate.SeparateId == beforeItemSeparate.Id || itemSeparate.StockDistributionItemId == beforeItemSeparate.StockDistributionItemId {
										beforeItemSeparateQuantity = beforeItemSeparate.Quantity
										beforeItemSeparateStatus = beforeItemSeparate.Status
										removedItemSeparates = append(removedItemSeparates, beforeItemSeparate)
									}
								}
							}
						}
					}
					if itemSeparate.IsDelete == false {
						if isRefund == false {
							if item.Status != enum.SaleOrderFinished.String() && item.Status != enum.StockDistributed.String() {
								return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Create Order Separate Status not Allowed :"+item.Status)
							}
						}
						if isRefund == true {
							if item.Status == enum.SaleShippingWaiting.String() || item.Status == enum.SaleShippingProcessing.String() {
								return nil, nil, factory.NewError(api.ErrorInvalidStatus, "Create Order Separate Status not Allowed :"+item.Status)
							}
						}
						itemSeparateQuantity := itemSeparate.Quantity
						totalQuantity += itemSeparateQuantity

						var orderItemSeparate OrderItemSeparate
						orderItemSeparate.OrderItemId = item.Id
						orderItemSeparate.StockDistributionItemId = itemSeparate.StockDistributionItemId
						orderItemSeparate.ListPrice = item.ListPrice
						orderItemSeparate.SalePrice = item.SalePrice
						orderItemSeparate.Quantity = itemSeparateQuantity
						orderItemSeparate.TotalListPrice = number.ToFixed(orderItemSeparate.ListPrice*float64(itemSeparateQuantity), nil)
						orderItemSeparate.TotalSalePrice = number.ToFixed(orderItemSeparate.SalePrice*float64(itemSeparateQuantity), nil)
						orderItemSeparate.TotalDiscountPrice = number.ToFixed(orderItemSeparate.TotalListPrice-orderItemSeparate.TotalSalePrice, nil)
						orderItemSeparate.TotalPaymentPrice = number.ToFixed(orderItemSeparate.TotalSalePrice, nil)

						if isRefund == true {
							if beforeItemSeparateStatus == enum.StockDistributed.String() {
								orderItemSeparate.Status = enum.SaleOrderFinished.String()
							} else if beforeItemSeparateStatus == "" {
								if item.Status == enum.StockDistributed.String() {
									orderItemSeparate.Status = enum.SaleOrderFinished.String()
								} else {
									orderItemSeparate.Status = item.Status
								}
							} else {
								orderItemSeparate.Status = beforeItemSeparateStatus
							}
						} else {
							orderItemSeparate.Status = item.Status
						}
						orderItemSeparates = append(orderItemSeparates, orderItemSeparate)
					}
				}

				if beforeItemSeparateQuantity == 0 {
					for _, beforeItemSeparate := range beforeItemSeparates {
						beforeItemSeparateQuantity += beforeItemSeparate.Quantity
						beforeItemSeparateStatus = beforeItemSeparate.Status
						removedItemSeparates = append(removedItemSeparates, beforeItemSeparate)
					}
				}

				logrus.WithField("Quantity", item.Quantity).Info("order.Item")
				logrus.WithField("Quantity", totalQuantity).Info("order.Item.ItemSeparates")
				logrus.WithField("Quantity", beforeItemSeparateQuantity).Info("beforeItemSeparateQuantity")
				if beforeItemSeparateQuantity == 0 && totalQuantity != item.Quantity {
					return nil, nil, factory.NewError(api.ErrorMissParameter, "Order Item Separate Quantity Incorrect")
				}
				if beforeItemSeparateQuantity != 0 && totalQuantity != beforeItemSeparateQuantity {
					return nil, nil, factory.NewError(api.ErrorMissParameter, "Order Item ReSeparate Quantity Incorrect")
				}
				totalQuantity = 0
			}
		}
		if isItemExist == false {
			return nil, nil, factory.NewError(api.ErrorMissParameter, "Order Separate ItemId Incorrect")
		}
		isItemExist = false
	}

	return removedItemSeparates, orderItemSeparates, nil
}

func (OrderItemSeparate) GetOrderItemSeparates(ctx context.Context, orderItemIds []int64) ([]OrderItemSeparate, error) {
	var orderItemSeparates []OrderItemSeparate
	if err := factory.DB(ctx).In("order_item_id", orderItemIds).And("is_delete = ?", false).Find(&orderItemSeparates); err != nil {
		return nil, err
	}

	return orderItemSeparates, nil
}

func (OrderItemSeparate) GetOrderItemSeparate(ctx context.Context, itemSeparateId int64) (OrderItemSeparate, error) {
	var orderItemSeparate OrderItemSeparate
	if exist, err := factory.DB(ctx).Where("id = ?", itemSeparateId).And("is_delete = ?", false).Get(&orderItemSeparate); err != nil {
		return OrderItemSeparate{}, err
	} else if !exist {
		return OrderItemSeparate{}, nil
	}

	return orderItemSeparate, nil
}

func (OrderItemSeparate) GetOrderItemSeparateByStockDistributedItemId(ctx context.Context, orderItemId int64, stockDistributedItemId int64) (OrderItemSeparate, error) {
	var orderItemSeparate OrderItemSeparate
	if exist, err := factory.DB(ctx).Where("order_item_id = ?", orderItemId).And("stock_distribution_item_id = ?", stockDistributedItemId).And("is_delete = ?", false).Get(&orderItemSeparate); err != nil {
		return OrderItemSeparate{}, err
	} else if !exist {
		return OrderItemSeparate{}, nil
	}

	return orderItemSeparate, nil
}

func (OrderReseller) GetOrderResellers(ctx context.Context, orderItemIds []int64) ([]OrderReseller, error) {
	var orderResellers []OrderReseller
	if err := factory.DB(ctx).In("order_item_id", orderItemIds).Find(&orderResellers); err != nil {
		return nil, err
	}

	return orderResellers, nil
}

func OrderItemSeparateUpdate(ctx context.Context, beforeItemSeparates []OrderItemSeparate, newItemSeparates []OrderItemSeparate) ([]OrderItemSeparate, error) {
	for _, beforeItemSeparate := range beforeItemSeparates {
		beforeItemSeparate.IsDelete = true
		if err := beforeItemSeparate.Update(ctx); err != nil {
			return nil, err
		}
	}

	orderItemSeparates, err := OrderItemSeparateSave(ctx, newItemSeparates)
	if err != nil {
		return nil, err
	}

	return orderItemSeparates, nil
}

func OrderItemSeparateSave(ctx context.Context, orderItemSeparates []OrderItemSeparate) ([]OrderItemSeparate, error) {
	for i := range orderItemSeparates {
		orderItemSeparates[i].CreatedAt = time.Now().UTC()
		orderItemSeparates[i].UpdatedAt = orderItemSeparates[i].CreatedAt
		if err := orderItemSeparates[i].insert(ctx); err != nil {
			return nil, err
		}
	}

	return orderItemSeparates, nil
}

func OrderItemSeparatesDelete(ctx context.Context, beforeItemSeparates []OrderItemSeparate, isDelete bool) error {
	for i := range beforeItemSeparates {
		beforeItemSeparates[i].IsDelete = isDelete
		if err := beforeItemSeparates[i].Update(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (orderItemSeparate *OrderItemSeparate) insert(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(orderItemSeparate)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	return nil
}

func (orderItemSeparate *OrderItemSeparate) Update(ctx context.Context) (err error) {
	session := factory.DB(ctx).ID(orderItemSeparate.Id)
	session.Cols("is_delete")

	_, err = session.Update(orderItemSeparate)
	if err != nil {
		return err
	}

	return nil
}
