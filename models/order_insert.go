package models

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/factory"
)

func (order *Order) Validate(ctx context.Context) error {

	// Todo : 현재 마일리지 체크는 주문 생성 후 PAY전에 체크

	if err := stockEventHandler.StockValidate(ctx, order); err != nil {
		return err
	}

	if err := calculatorEventHandler.OrderCalculate(ctx, order); err != nil {
		return err
	}
	if err := orderPaymentPriceValidate(order); err != nil {
		return err
	}
	return nil
}

func (o *Order) Save(ctx context.Context) error {
	o.CreatedAt = time.Now().UTC()
	o.UpdatedAt = o.CreatedAt
	if err := o.insert(ctx); err != nil {
		return err
	}

	for i := range o.Items {
		o.Items[i].OrderId = o.Id
		o.Items[i].CreatedAt = o.CreatedAt
		o.Items[i].UpdatedAt = o.CreatedAt
		if err := o.Items[i].insert(ctx); err != nil {
			return err
		}
	}

	for i := range o.Offers {
		o.Offers[i].OrderId = o.Id
		o.Offers[i].CreatedAt = o.CreatedAt
		itemIds := []string{}
		targetItemIds := []string{}
		for _, item := range o.Items {
			for _, itemOffer := range item.AppliedCartOffers {
				if itemOffer.OfferNo == o.Offers[i].OfferNo {
					if itemOffer.IsTarget {
						targetItemIds = append(targetItemIds, strconv.FormatInt(item.Id, 10))
					} else {
						itemIds = append(itemIds, strconv.FormatInt(item.Id, 10))
					}
				}
			}
		}
		o.Offers[i].ItemIds = strings.Join(itemIds, ",")
		o.Offers[i].TargetItemIds = strings.Join(targetItemIds, ",")
		if err := o.Offers[i].Save(ctx); err != nil {
			return err
		}
	}
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	orderHistory := o.NewOrderHistory(userClaim)
	if err := orderHistory.Save(ctx); err != nil {
		return err
	}

	if err := factory.CommitDB(ctx); err != nil {
		return err
	}

	/*
			커밋되고 나서 이벤트 발행 작업은 goroutine을 사용하여 비동기로 처리함.
			동기로 처리하게 되면 아래와 같은 문제가 았을 수 있음
		 		- 메시지 브로커(카프카)쪽 병목이 생기는 경우 클라언트가 응답을 대기하고 있어야함.
		 		- 이벤트를 발행 하면서 오류가 발생하는 경우 주문이 이미 생성(이미 커밋 되었으므로)이 되었지만 클라이언트는 오류를 반환 받음.
	*/
	go func() {
		o.PublishEventMessagesInOutBox(ctx)
	}()

	return nil
}

func (order *Order) insert(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(order)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	return nil
}

func (orderItem *OrderItem) insert(ctx context.Context) error {
	if _, err := factory.DB(ctx).Insert(orderItem); err != nil {
		return err
	}
	for i := range orderItem.AppliedCartOffers {
		orderItem.AppliedCartOffers[i].OrderId = orderItem.OrderId
		orderItem.AppliedCartOffers[i].OrderItemId = orderItem.Id
		orderItem.AppliedCartOffers[i].CreatedAt = orderItem.CreatedAt
		if err := orderItem.AppliedCartOffers[i].Save(ctx); err != nil {
			return err
		}
	}

	for i := range orderItem.Resellers {
		orderItem.Resellers[i].OrderId = orderItem.OrderId
		orderItem.Resellers[i].OrderItemId = orderItem.Id
		orderItem.Resellers[i].CreatedAt = orderItem.CreatedAt
		if err := orderItem.Resellers[i].insert(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (orderReseller *OrderReseller) insert(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(orderReseller)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	return nil
}

func orderPaymentPriceValidate(order *Order) error {
	if order.Mileage > 0 {
		if order.MileagePrice <= 0 {
			return errors.New("Order Mileage error, Order Mileage =" + strconv.FormatFloat(order.Mileage, 'f', 2, 64) + ", Order MileagePrice =" + strconv.FormatFloat(order.MileagePrice, 'f', 2, 64))
		}
		if order.MileagePrice > order.TotalPaymentPrice {
			return errors.New("Order MileagePrice better TotalPaymentPrice, Order MileagePrice =" + strconv.FormatFloat(order.MileagePrice, 'f', 2, 64) + ", Order TotalPaymentPrice =" + strconv.FormatFloat(order.TotalPaymentPrice, 'f', 2, 64))
		}
	}
	if order.TotalPaymentPrice < 0 {
		return errors.New("Order TotalPaymentPrice error, Order TotalPaymentPrice =" + strconv.FormatFloat(order.TotalPaymentPrice, 'f', 2, 64))
	}
	for _, item := range order.Items {
		if item.TotalPaymentPrice < 0 {
			return errors.New("OrderItem TotalPaymentPrice error, OrderItem TotalPaymentPrice =" + strconv.FormatFloat(item.TotalPaymentPrice, 'f', 2, 64))
		}
	}
	return nil
}
