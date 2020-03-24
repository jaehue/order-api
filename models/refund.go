package models

import (
	"context"
	"nomni/utils/auth"
	"strings"
	"time"

	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"

	"github.com/go-xorm/xorm"
)

func (Refund) GetRefunds(ctx context.Context, tenantCode string, customerId int64, refundId int64, orderId, storeId int64, outerOrderNo, refundStatus, startAt, endAt string, skipCount int, maxResultCount int, IsCancelInclude bool) (int64, []Refund, error) {
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
		if refundId != 0 {
			q.And("id = ?", refundId)
		}
		if orderId != 0 {
			q.And("order_id = ?", orderId)
		}
		if storeId != 0 {
			q.And("store_id = ?", storeId)
		}
		if outerOrderNo != "" {
			q.And("outer_order_no = ?", outerOrderNo)
		}
		if refundStatus != "" {
			statusArry := strings.Split(refundStatus, ",")
			q.In("status", statusArry)
		}
		if IsCancelInclude == false {
			q.And("status != ?", enum.RefundOrderCancel.String())
		}
		if startAt != "" && endAt != "" {
			q.And("created_at >= ?", timeStart)
			q.And("created_at < ?", timeEnd.AddDate(0, 0, 1))
		}
		return q
	}
	totalCount, _ := queryBuilder().Count(&Refund{})
	if totalCount == 0 {
		return 0, []Refund{}, nil
	}
	q := queryBuilder().Desc("id")
	if maxResultCount != 0 {
		q.Limit(maxResultCount, skipCount)
	}
	var refunds []Refund
	if err := q.Find(&refunds); err != nil {
		return 0, nil, err
	}
	var (
		refundItems []RefundItem
		refundIds   []int64
	)
	for _, item := range refunds {
		refundIds = append(refundIds, item.Id)
	}
	refundItems, err := RefundItem{}.GetRefundItems(ctx, tenantCode, refundIds, nil, nil, nil)
	if err != nil {
		return 0, nil, err
	}

	refundOffers, oerr := OrderOffer{}.GetOrderOffers(ctx, tenantCode, customerId, "", nil, refundIds)
	if oerr != nil {
		return 0, nil, oerr
	}
	extensions, err := getRefundExtensions(ctx, refundIds)
	if err != nil {
		return 0, nil, err
	}
	itemOffers, err := ItemAppliedCartOffer{}.GetItemAppliedOffers(ctx, "", nil, nil, refundIds, nil)
	if err != nil {
		return 0, nil, err
	}
	for i := range refunds {
		for _, refundOffer := range refundOffers {
			if refundOffer.OrderId == refunds[i].OrderId && refundOffer.RefundId == refunds[i].Id {
				refunds[i].Offers = append(refunds[i].Offers, refundOffer)
			}
		}
		for _, refundItem := range refundItems {
			if refundItem.RefundId == refunds[i].Id {
				for _, itemOffer := range itemOffers {
					if itemOffer.RefundItemId == refundItem.Id {
						refundItem.GroupOffers = append(refundItem.GroupOffers, itemOffer)
					}
				}
				refunds[i].Items = append(refunds[i].Items, refundItem)
			}
		}
		for _, extension := range extensions {
			if extension.RefundId == refunds[i].Id {
				refunds[i].Extension = extension
			}
		}
	}

	return totalCount, refunds, nil
}

func (Refund) GetRefund(ctx context.Context, tenantCode string, customerId int64, refundId int64, orderId int64, refundItemIds []int64, refundStatus string, IsCancelInclude bool) (Refund, error) {
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
		if refundId != 0 {
			q.And("id = ?", refundId)
		}
		if orderId != 0 {
			q.And("order_id = ?", orderId)
		}
		if refundStatus != "" {
			q.And("status = ?", refundStatus)
		}
		if IsCancelInclude == false {
			q.And("status != ?", enum.RefundOrderCancel.String())
		}
		return q
	}

	var refund Refund
	if exist, err := queryBuilder().Get(&refund); err != nil {
		return Refund{}, err
	} else if !exist {
		return Refund{}, nil
	}

	var (
		refundItems []RefundItem
		refundIds   []int64
	)

	refundIds = append(refundIds, refund.Id)
	refundItems, err := RefundItem{}.GetRefundItems(ctx, tenantCode, refundIds, refundItemIds, nil, nil)
	if err != nil {
		return Refund{}, err
	}
	extensions, err := getRefundExtensions(ctx, refundIds)
	if err != nil {
		return Refund{}, err
	}
	if len(extensions) > 0 {
		refund.Extension = extensions[0]
	}
	for _, refundItem := range refundItems {
		refund.Items = append(refund.Items, refundItem)
	}

	refundOffers, oerr := OrderOffer{}.GetOrderOffer(ctx, tenantCode, customerId, "", 0, refund.Id)
	if oerr != nil {
		return Refund{}, oerr
	}
	for _, refundOffer := range refundOffers {
		if refundOffer.OrderId == refund.OrderId && refundOffer.RefundId == refund.Id {
			refund.Offers = append(refund.Offers, refundOffer)
		}
	}
	itemOffers, err := ItemAppliedCartOffer{}.GetItemAppliedOffers(ctx, "", nil, nil, refundIds, nil)
	for j, item := range refund.Items {
		for _, itemOffer := range itemOffers {
			if itemOffer.RefundItemId == item.Id {
				refund.Items[j].GroupOffers = append(refund.Items[j].GroupOffers, itemOffer)
			}
		}
	}
	return refund, nil
}

func (RefundItem) GetRefundItems(ctx context.Context, tenantCode string, refundIds []int64, refundItemIds []int64, orderIds []int64, orderItemIds []int64) ([]RefundItem, error) {
	var refundItems []RefundItem
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx)
		if len(refundIds) > 0 {
			q.In("refund_id", refundIds)
		}
		if len(refundItemIds) > 0 {
			q.In("id", refundItemIds)
		}
		if len(orderIds) > 0 {
			q.In("order_id", orderIds)
		}
		if len(orderItemIds) > 0 {
			q.In("order_item_id", orderItemIds)
		}
		return q
	}
	q := queryBuilder()
	if err := q.Find(&refundItems); err != nil {
		return nil, err
	}
	if len(refundItems) == 0 {
		return nil, NotFoundDataError
	}

	return refundItems, nil
}
func getRefundExtensions(ctx context.Context, refundIds []int64) (extensions []RefundExtension, err error) {
	queryBuilder := func() xorm.Interface {
		q := factory.DB(ctx)
		if len(refundIds) > 0 {
			q.In("refund_id", refundIds)
		}
		return q
	}
	q := queryBuilder()
	if err := q.Find(&extensions); err != nil {
		return nil, err
	}
	return
}
