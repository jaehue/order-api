package models

import (
	"context"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/factory"
)

func (order *Order) Update(ctx context.Context) error {
	err := order.update(ctx)
	if err != nil {
		return err
	}

	userClaim := auth.UserClaim{}.FromCtx(ctx)
	orderHistory := order.NewOrderHistory(userClaim)
	if err := orderHistory.Save(ctx); err != nil {
		return err
	}

	return nil
}

func (order *Order) update(ctx context.Context) (err error) {
	session := factory.DB(ctx).ID(order.Id)
	session.Cols("is_cust_del")

	_, err = session.Update(order)
	if err != nil {
		return err
	}

	return nil
}

func (order *Order) changeOrderStatus(ctx context.Context, orderItemIds []int64, separateIds []int64, status string) error {
	isOrderChange := false

	if order.Status != status {
		isOrderChange = true
		order.Status = status
		if err := order.updateStatus(ctx); err != nil {
			return err
		}
	}

	isOrderItemChange, err := order.changeOrderItemStatus(ctx, orderItemIds, separateIds, status)
	if err != nil {
		return err
	}

	if isOrderChange == true || isOrderItemChange == true {
		userClaim := auth.UserClaim{}.FromCtx(ctx)
		orderHistory := order.NewOrderHistory(userClaim)
		if err := orderHistory.Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (order *Order) changeOrderItemStatus(ctx context.Context, orderItemIds []int64, separateIds []int64, status string) (bool, error) {
	isOrderItemChange := false

	if len(orderItemIds) == 0 {
		for i := range order.Items {
			if order.Items[i].Status != status {
				isOrderItemChange = true
				order.Items[i].Status = status
				if err := order.Items[i].updateStatus(ctx); err != nil {
					return isOrderItemChange, err
				}

				if len(separateIds) == 0 {
					for j := range order.Items[i].ItemSeparates {
						if order.Items[i].ItemSeparates[j].StockDistributionItemId != 0 {
							order.Items[i].ItemSeparates[j].Status = status
							if err := order.Items[i].ItemSeparates[j].updateStatus(ctx); err != nil {
								return isOrderItemChange, err
							}
						}
					}
				}
			}
		}
	} else {
		for i := range order.Items {
			for _, itemId := range orderItemIds {
				if order.Items[i].Id == itemId {
					if order.Items[i].Status != status {
						isOrderItemChange = true
						order.Items[i].Status = status

						if err := order.Items[i].updateStatus(ctx); err != nil {
							return isOrderItemChange, err
						}

					}
					if len(separateIds) == 0 {
						for j := range order.Items[i].ItemSeparates {
							if order.Items[i].ItemSeparates[j].StockDistributionItemId != 0 {
								order.Items[i].ItemSeparates[j].Status = status

								if err := order.Items[i].ItemSeparates[j].updateStatus(ctx); err != nil {
									return isOrderItemChange, err
								}
							}
						}
					} else {
						for j, separate := range order.Items[i].ItemSeparates {
							for _, separateId := range separateIds {
								if separate.OrderItemId == itemId && separate.Id == separateId {
									if separate.Status != status && separate.StockDistributionItemId != 0 {
										order.Items[i].ItemSeparates[j].Status = status

										if err := order.Items[i].ItemSeparates[j].updateStatus(ctx); err != nil {
											return isOrderItemChange, err
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

	return false, nil
}

func (order *Order) updateStatus(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(order.Id).
		Cols("status").
		Update(order); err != nil {
		return err
	}

	return nil
}

func (orderItem *OrderItem) updateStatus(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(orderItem.Id).
		And("order_id = ?", orderItem.OrderId).
		Cols("status").
		Update(orderItem); err != nil {
		return err
	}

	return nil
}

func (orderItem *OrderItem) UpdateFeeRate(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(orderItem.Id).
		And("order_id = ?", orderItem.OrderId).
		Cols("fee_rate").
		Update(orderItem); err != nil {
		return err
	}

	return nil
}
func (itemSeparate *OrderItemSeparate) updateStatus(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(itemSeparate.Id).
		And("order_item_id = ?", itemSeparate.OrderItemId).
		Cols("status").
		Update(itemSeparate); err != nil {
		return err
	}
	return nil
}
