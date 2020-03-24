package models

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"

	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

func (Order) GetOrders(ctx context.Context, tenantCode string, customerId, storeId, salesmanId int64, orderStatus, saleType, startAt, endAt, ids, outerOrderNo string, skipCount int, maxResultCount int) (int64, []Order, error) {
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
		if customerId == 0 {
			return 0, nil, errors.New("customerId:0")
		}
	}

	var timeStart, timeEnd time.Time
	var errt error
	if startAt != "" {
		if timeStart, errt = DateParseToUtc(startAt); errt != nil {
			return 0, nil, errt
		}
	}
	if endAt != "" {
		if timeEnd, errt = DateParseToUtc(endAt); errt != nil {
			return 0, nil, errt
		}
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
		if storeId != 0 {
			q.And("store_id = ?", storeId)
		}
		if salesmanId != 0 {
			q.And("salesman_id = ?", salesmanId)
		}
		if saleType != "" {
			q.And("sale_type = ?", saleType)
		}
		if userClaim.Iss == auth.IssMembership {
			q.And("customer_id = ?", customerId)
		} else if customerId != 0 {
			q.And("customer_id = ?", customerId)
		}
		if outerOrderNo != "" {
			q.And("outer_order_no = ?", outerOrderNo)
		}
		if ids != "" {
			idsArry := strings.Split(ids, ",")
			orderIds := []int64{}
			for _, item := range idsArry {
				id, _ := strconv.ParseInt(item, 10, 64)
				orderIds = append(orderIds, id)
			}
			q.In("id", orderIds)
		}
		if orderStatus != "" {
			statusArry := strings.Split(orderStatus, ",")
			q.In("status", statusArry)
		}
		if startAt != "" && endAt != "" {
			q.And("created_at>=?", timeStart)
			q.And("created_at<?", timeEnd.AddDate(0, 0, 1))
		}
		return q
	}
	totalCount, _ := queryBuilder().Count(&Order{})
	if totalCount == 0 {
		return 0, []Order{}, nil
	}

	q := queryBuilder().Desc("id")
	if maxResultCount != 0 {
		q.Limit(maxResultCount, skipCount)
	}
	var orders []Order
	if err := q.Find(&orders); err != nil {
		return 0, nil, err
	}

	var (
		orderItems []OrderItem
		orderIds   []int64
	)
	for _, item := range orders {
		orderIds = append(orderIds, item.Id)
	}
	orderItems, err := OrderItem{}.GetOrderItems(ctx, tenantCode, orderIds, nil)
	if err != nil {
		return 0, nil, err
	}
	resultOrders := orders
	// resultOrders, err := GetOrdersBelow(ctx, tenantCode, customerId, orders)
	// if err != nil {
	// 	return 0, nil, err
	// }
	for i := range resultOrders {
		_, refunds, _ := Refund{}.GetRefunds(ctx, tenantCode, customerId, 0, resultOrders[i].Id, 0, "", "", "", "", 0, 0, false)
		for _, orderItem := range orderItems {
			if len(orderItem.ItemSeparates) == 0 {
				for _, refund := range refunds {
					for _, item := range refund.Items {
						if item.OrderId == orderItem.OrderId && item.OrderItemId == orderItem.Id {
							orderItem.RefundItems = append(orderItem.RefundItems, item)
						}
					}
				}
			} else {
				for k := range orderItem.ItemSeparates {
					for _, refund := range refunds {
						for j := range refund.Items {
							if orderItem.ItemSeparates[k].Id == refund.Items[j].SeparateId {
								orderItem.ItemSeparates[k].RefundItem = &refund.Items[j]
							}
						}
					}
				}
			}
			if orderItem.OrderId == resultOrders[i].Id {
				resultOrders[i].Items = append(resultOrders[i].Items, orderItem)
			}
		}
	}

	resultOrders, err = GetOrdersOffer(ctx, tenantCode, customerId, resultOrders)
	if err != nil {
		return 0, nil, err
	}

	return totalCount, resultOrders, nil
}

func (Order) GetOrder(ctx context.Context, tenantCode string, customerId int64, orderId int64, orderItemIds []int64, orderStatus string, IsCancelInclude bool) (Order, error) {
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
	}
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Where("id = ?", orderId)
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
		if orderStatus != "" {
			q.And("status = ?", orderStatus)
		}
		if IsCancelInclude == false {
			q.And("status != ?", enum.SaleOrderCancel.String())
		}
		return q
	}

	var order Order
	if exist, err := queryBuilder().Get(&order); err != nil {
		return Order{}, err
	} else if !exist {
		return Order{}, nil
	}

	var (
		orderItems []OrderItem
		orderIds   []int64
	)
	orderIds = append(orderIds, order.Id)

	orderItems, err := OrderItem{}.GetOrderItems(ctx, tenantCode, orderIds, orderItemIds)
	if err != nil {
		return Order{}, err
	}
	resultOrder := order
	// resultOrder, err := GetOrderBelow(ctx, tenantCode, customerId, order)
	// if err != nil {
	// 	return Order{}, err
	// }
	_, refunds, _ := Refund{}.GetRefunds(ctx, tenantCode, customerId, 0, orderId, 0, "", "", "", "", 0, 0, false)

	for _, orderItem := range orderItems {
		if len(orderItem.ItemSeparates) == 0 {
			for _, refund := range refunds {
				for _, item := range refund.Items {
					if item.OrderId == orderItem.OrderId && item.OrderItemId == orderItem.Id {
						orderItem.RefundItems = append(orderItem.RefundItems, item)
					}
				}
			}
		} else {
			for i := range orderItem.ItemSeparates {
				for _, refund := range refunds {
					for j := range refund.Items {
						if orderItem.ItemSeparates[i].Id == refund.Items[j].SeparateId {
							orderItem.ItemSeparates[i].RefundItem = &refund.Items[j]
						}
					}
				}
			}
		}
		resultOrder.Items = append(resultOrder.Items, orderItem)
	}

	resultOrder, err = GetOrderOffer(ctx, tenantCode, customerId, resultOrder)
	if err != nil {
		return Order{}, err
	}

	return resultOrder, nil
}

func (Order) GetOrdersByItem(ctx context.Context, tenantCode string, customerId, storeId int64, orderStatus, ids, orderItemIds, startAt, endAt string, skipCount int, maxResultCount int, isRemoveRefund bool) (int64, []Order, error) {
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
		if customerId == 0 {
			return 0, nil, errors.New("customerid:0")
		}
	}

	totalCount, orderItemResults, err := OrderItem{}.GetOrderItemsWithRefundItem(ctx, tenantCode, customerId, storeId, orderStatus, ids, orderItemIds, startAt, endAt, skipCount, maxResultCount, isRemoveRefund)
	if err != nil {
		return 0, nil, err
	}
	if totalCount == 0 {
		return 0, nil, nil
	}

	var orderIds []int64
	for _, item := range orderItemResults {
		orderIds = append(orderIds, item.OrderId)
	}

	resultOrders, err := Order{}.GetOnlyOrders(ctx, tenantCode, customerId, orderIds)
	if err != nil {
		return 0, nil, err
	}

	for i, order := range resultOrders {
		for _, orderItem := range orderItemResults {
			if order.Id == orderItem.OrderId {
				resultOrders[i].Items = append(resultOrders[i].Items, orderItem)
			}
		}
	}

	return totalCount, resultOrders, nil
}

func (Order) GetOnlyOrders(ctx context.Context, tenantCode string, customerId int64, orderIds []int64) ([]Order, error) {
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
		if len(orderIds) > 0 {
			q.In("id", orderIds)
		}
		return q
	}
	q := queryBuilder().Desc("id")

	var orders []Order
	if err := q.Find(&orders); err != nil {
		return nil, err
	}
	// resultOrders, err := GetOrdersBelow(ctx, tenantCode, customerId, orders)
	// if err != nil {
	// 	return nil, err
	// }

	resultOrders, err := GetOrdersOffer(ctx, tenantCode, customerId, orders)
	if err != nil {
		return nil, err
	}

	return resultOrders, nil
}

func (Order) GetOrderStatusCount(ctx context.Context, tenantCode string, customerId int64, isOrder bool) ([]StatusCount, error) {
	statusCounts := []StatusCount{}
	statusList := []string{
		enum.SaleOrderProcessing.String(),
		enum.SaleOrderCancel.String(),
		enum.SaleOrderFinished.String(),
		enum.StockDistributed.String(),
		enum.SaleShippingWaiting.String(),
		enum.SaleShippingProcessing.String(),
		enum.SaleShippingFinished.String(),
		enum.BuyerReceivedConfirmed.String(),
		enum.SaleOrderSuccess.String(),
	}
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
		if customerId == 0 {
			return nil, errors.New("customerId:0")
		}
	}
	totalCount, orderResults, err := Order{}.GetOrdersByItem(ctx, tenantCode, customerId, 0, "", "", "", "", "", 0, 0, true)
	if err != nil {
		return nil, err
	}
	if totalCount == 0 {
		return nil, nil
	}
	var statusTotalCount int64
	var orderId int64
	for _, orderStatus := range statusList {
		logrus.WithField("orderStatus", orderStatus).Info("statusList")
		if isOrder == true {
			for _, order := range orderResults {
				for _, orderItem := range order.Items {
					if orderId == orderItem.OrderId {
						break
					}
					logrus.WithField("Id", orderItem.Id).Info("orderItem")
					if len(orderItem.ItemSeparates) == 0 {
						if orderItem.Status == orderStatus {
							logrus.WithField("Status", orderItem.Status).Info("orderItem")
							statusTotalCount++
							orderId = orderItem.OrderId
							break
						}
					} else {
						for _, itemSeparate := range orderItem.ItemSeparates {
							if itemSeparate.Status == orderStatus {
								logrus.WithField("Status", itemSeparate.Status).Info("itemSeparate")
								statusTotalCount++
								orderId = orderItem.OrderId
								break
							}
						}
					}
				}
			}
		} else {
			for _, order := range orderResults {
				for _, orderItem := range order.Items {
					logrus.WithField("Id", orderItem.Id).Info("orderItem")
					if len(orderItem.ItemSeparates) == 0 {
						if orderItem.Status == orderStatus {
							logrus.WithField("Status", orderItem.Status).Info("orderItem")
							statusTotalCount++
						}
					} else {
						for _, itemSeparate := range orderItem.ItemSeparates {
							if itemSeparate.Status == orderStatus {
								logrus.WithField("Status", itemSeparate.Status).Info("itemSeparate")
								statusTotalCount++
							}
						}
					}
				}
			}
		}
		statusCount := StatusCount{
			Status:     orderStatus,
			TotalCount: statusTotalCount,
		}
		orderId = 0
		statusTotalCount = 0
		statusCounts = append(statusCounts, statusCount)
	}

	return statusCounts, nil
}

func GetOrderBelow(ctx context.Context, tenantCode string, customerId int64, order Order) (Order, error) {
	var orders []Order
	orders = append(orders, order)
	resultOrders, err := GetOrdersBelow(ctx, tenantCode, customerId, orders)
	if err != nil {
		return Order{}, err
	}

	return resultOrders[0], nil
}
func GetOrdersBelow(ctx context.Context, tenantCode string, customerId int64, orders []Order) ([]Order, error) {
	var orderIds []int64
	for _, item := range orders {
		orderIds = append(orderIds, item.Id)
	}
	return orders, nil
}

func GetOrderOffer(ctx context.Context, tenantCode string, customerId int64, order Order) (Order, error) {
	var orders []Order
	orders = append(orders, order)
	resultOrders, err := GetOrdersOffer(ctx, tenantCode, customerId, orders)
	if err != nil {
		return Order{}, err
	}

	return resultOrders[0], nil
}
func GetOrdersOffer(ctx context.Context, tenantCode string, customerId int64, resultOrders []Order) ([]Order, error) {
	var orderIds []int64
	for _, item := range resultOrders {
		orderIds = append(orderIds, item.Id)
	}

	orderOffers, err := OrderOffer{}.GetOrderOffers(ctx, tenantCode, customerId, "", orderIds, nil)
	if err != nil {
		return nil, err
	}
	for i, _ := range resultOrders {
		for _, orderOffer := range orderOffers {
			if orderOffer.OrderId == resultOrders[i].Id && orderOffer.RefundId == int64(0) {
				resultOrders[i].Offers = append(resultOrders[i].Offers, orderOffer)
			}
		}
	}
	itemOffers, err := ItemAppliedCartOffer{}.GetItemAppliedOffers(ctx, "", orderIds, nil, nil, nil)
	for i, _ := range resultOrders {
		for j, item := range resultOrders[i].Items {
			for _, itemOffer := range itemOffers {
				if itemOffer.OrderItemId == item.Id && itemOffer.RefundId == 0 {
					resultOrders[i].Items[j].GroupOffers = append(resultOrders[i].Items[j].GroupOffers, itemOffer)
				}
			}
		}
	}
	return resultOrders, nil
}

func (Order) OrderCheckAndCancel(ctx context.Context, id, customerId int64) {
	oldOrder, _ := Order{}.GetOrder(ctx, "", customerId, id, []int64{}, "", true)
	if oldOrder.Id != 0 && oldOrder.Status == enum.SaleOrderProcessing.String() {
		if order, err := (Order{}).ChangeStatus(ctx, id, []int64{}, []int64{}, enum.SaleOrderCancel.String()); err != nil {
			logrus.WithField("OrderCheckAndCancel", err).Info("SaleOrderCancel")
		} else {
			go func() {
				// 변경된 상태를 메시지 이벤트로 발행
				order.PublishEventMessagesInOutBox(ctx)
			}()
		}

	}
}
