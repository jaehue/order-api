package models

import (
	"context"
	"fmt"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"

	"github.com/pkg/errors"
)

func (Order) ChangeStatus(ctx context.Context, orderId int64, orderItemIds []int64, separateIds []int64, status string) (*Order, error) {
	order, err := Order{}.GetOrder(ctx, "", 0, orderId, orderItemIds, "", false)
	if err != nil {
		return nil, err
	}

	if order.Id == 0 {
		return nil, api.ErrorInvalidStatus.New(nil, "Order is Not Exist or Deleted")
	}

	orderType := enum.FindOrderTypeFromString(status)
	if orderType == enum.Unknown {
		return nil, errors.New(fmt.Sprintf(`unknown status: %v`, status))
	}

	order.addEventToOutBox(orderType)

	if err := order.changeOrderStatus(ctx, orderItemIds, separateIds, status); err != nil {
		return nil, err
	}

	orderNotDeliveryItems := []int64{}
	for _, item := range order.Items {
		if item.IsDelivery == false {
			orderNotDeliveryItems = append(orderNotDeliveryItems, item.Id)
		}
	}
	if status == enum.SaleOrderFinished.String() && len(orderNotDeliveryItems) > 0 {
		if len(order.Items) == len(orderNotDeliveryItems) {
			orderNotDeliveryItems = nil
		}
		order.addEventToOutBox(enum.BuyerReceivedConfirmed)
		if _, err := order.ChangeStatus(ctx, orderId, orderNotDeliveryItems, separateIds, enum.BuyerReceivedConfirmed.String()); err != nil {
			return nil, err
		}
	}

	return &order, nil
}
