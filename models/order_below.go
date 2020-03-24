package models

import (
	"context"
	"strconv"
	"strings"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"

	"github.com/go-xorm/xorm"
)

func (OrderItem) GetOrderItems(ctx context.Context, tenantCode string, orderIds []int64, orderItemIds []int64) ([]OrderItem, error) {
	var orderItems []OrderItem

	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx)
		if len(orderIds) > 0 {
			q.In("order_id", orderIds)
		}
		if len(orderItemIds) > 0 {
			q.In("id", orderItemIds)
		}
		return q
	}
	q := queryBuilder()
	if err := q.Find(&orderItems); err != nil {
		return nil, err
	}

	if len(orderItems) == 0 {
		return nil, NotFoundDataError
	}

	items, err := GetOrderItemsBelow(ctx, tenantCode, orderItems, nil)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (OrderItem) GetOrderItemsWithRefundItem(ctx context.Context, tenantCode string, customerId, storeId int64, orderStatus, ids, itemIds, startAt, endAt string, skipCount int, maxResultCount int, isRemoveRefund bool) (int64, []OrderItem, error) {
	userClaim := auth.UserClaim{}.FromCtx(ctx)
	if userClaim.Iss == auth.IssMembership {
		customerId = userClaim.CustomerId
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

	orderIds := []int64{}
	orderItemIds := []int64{}
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx).Table("order_item").Select("order_item.*, order_item_separate.* ")
		q.Join("LEFT", "order_item_separate", "order_item.id = order_item_separate.order_item_id AND order_item_separate.is_delete=0")
		q.Join("LEFT", "refund_item", "order_item.order_id = refund_item.order_id AND order_item.id = refund_item.order_item_id AND refund_item.separate_id = IFNULL(order_item_separate.id, 0) AND refund_item.status not in ('"+enum.RefundOrderCancel.String()+"')")
		q.Where("1=1")
		if userClaim.TenantCode != "admin" {
			q.And("order_item.tenant_code = ?", userClaim.TenantCode)
		} else {
			if tenantCode != "" {
				q.And("order_item.tenant_code = ?", tenantCode)
			}
		}
		if isRemoveRefund == true {
			q.And("refund_item.id is null")
		}
		if ids != "" {
			idsArry := strings.Split(ids, ",")
			for _, item := range idsArry {
				id, _ := strconv.ParseInt(item, 10, 64)
				orderIds = append(orderIds, id)
			}
			q.In("order_item.order_id", orderIds)
		}
		if itemIds != "" {
			itemIdsArry := strings.Split(itemIds, ",")
			for _, item := range itemIdsArry {
				itemId, _ := strconv.ParseInt(item, 10, 64)
				orderItemIds = append(orderItemIds, itemId)
			}
			q.In("order_item.id", orderItemIds)
		}
		if storeId != 0 {
			q.And("order_item.store_id = ?", storeId)
		}
		if userClaim.Iss == auth.IssMembership {
			q.And("order_item.customer_id = ?", customerId)
		} else if customerId != 0 {
			q.And("order_item.customer_id = ?", customerId)
		}
		if orderStatus != "" {
			statusArry := strings.Split(orderStatus, ",")
			q.And("(order_item.status IN ('" + strings.Join(statusArry, "','") + "') AND order_item_separate.status is null) OR (order_item_separate.status IN ('" + strings.Join(statusArry, "','") + "'))")
		}
		if startAt != "" && endAt != "" {
			q.And("order_item.created_at >= ?", timeStart)
			q.And("order_item.created_at < ?", timeEnd.AddDate(0, 0, 1))
		}

		return q
	}

	var count []OrderItem
	if err := queryBuilder().Find(&count); err != nil {
		return 0, nil, err
	}
	totalCount := int64(len(count))
	if totalCount == 0 {
		return 0, nil, nil
	}
	q := queryBuilder().Desc("order_item.order_id").Desc("order_item.id")
	var results []struct {
		OrderItem         `xorm:"extends"`
		OrderItemSeparate `xorm:"extends"`
	}
	if maxResultCount != 0 {
		q.Limit(maxResultCount, skipCount)
	}
	if err := q.Find(&results); err != nil {
		return 0, nil, err
	}

	if len(results) == 0 {
		return 0, nil, NotFoundDataError
	}

	var orderItems []OrderItem
	var orderItemSeparates []OrderItemSeparate
	if orderItemIds != nil {
		orderItemIds = []int64{}
	}
	for i, result := range results {
		if i > 0 {
			if result.OrderItem.Id != results[i-1].OrderItem.Id {
				orderItems = append(orderItems, result.OrderItem)
			}
		} else {
			orderItems = append(orderItems, result.OrderItem)
		}
		orderItemIds = append(orderItemIds, result.OrderItem.Id)
		orderItemSeparates = append(orderItemSeparates, result.OrderItemSeparate)
	}

	items, err := GetOrderItemsBelow(ctx, tenantCode, orderItems, orderItemSeparates)
	if err != nil {
		return 0, nil, err
	}

	if isRemoveRefund == false {
		refundItems, _ := RefundItem{}.GetRefundItems(ctx, tenantCode, nil, nil, nil, orderItemIds)
		for i, orderItem := range items {
			if len(orderItem.ItemSeparates) == 0 {
				for _, item := range refundItems {
					if item.OrderId == orderItem.OrderId && item.OrderItemId == orderItem.Id {
						items[i].RefundItems = append(items[i].RefundItems, item)
					}
				}
			} else {
				for k := range orderItem.ItemSeparates {
					for j := range refundItems {
						if orderItem.ItemSeparates[k].Id == refundItems[j].SeparateId {
							items[i].ItemSeparates[k].RefundItem = &refundItems[j]
						}
					}
				}
			}
		}
	}

	return totalCount, items, nil
}

func GetOrderItemsBelow(ctx context.Context, tenantCode string, orderItems []OrderItem, itemSeparates []OrderItemSeparate) ([]OrderItem, error) {
	var orderItemIds []int64
	for _, orderItem := range orderItems {
		orderItemIds = append(orderItemIds, orderItem.Id)
	}

	var err error
	if len(itemSeparates) == 0 {
		itemSeparates, err = OrderItemSeparate{}.GetOrderItemSeparates(ctx, orderItemIds)
		if err != nil {
			return nil, err
		}
	}

	orderResellers, err := OrderReseller{}.GetOrderResellers(ctx, orderItemIds)
	if err != nil {
		return nil, err
	}

	for i := range orderItems {
		for _, itemSeparate := range itemSeparates {
			if itemSeparate.OrderItemId == orderItems[i].Id {
				orderItems[i].ItemSeparates = append(orderItems[i].ItemSeparates, itemSeparate)
			}
		}
		for _, orderReseller := range orderResellers {
			if orderReseller.OrderItemId == orderItems[i].Id {
				orderItems[i].Resellers = append(orderItems[i].Resellers, orderReseller)
			}
		}
	}
	return orderItems, nil
}
