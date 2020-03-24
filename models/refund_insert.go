package models

import (
	"context"
	"errors"
	"fmt"
	"math"
	"nomni/utils/auth"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"

	"github.com/pangpanglabs/goutils/number"
	"github.com/sirupsen/logrus"
)

func (refund *Refund) RefundValidate(ctx context.Context, tenantCode string, order Order) (Order, error) {
	_, preRefunds, err := Refund{}.GetRefunds(ctx, tenantCode, order.CustomerId, 0, refund.OrderId, 0, "", "", "", "", 0, 0, false)
	if err != nil {
		return Order{}, err
	}

	remainOrderRemovePreRefund, err := refund.OrderByPreRefundRemoved(ctx, preRefunds, order)
	if err != nil {
		return Order{}, err
	}

	remainOrderRemoveNewRefund, err := refund.OrderByNewRefundRemoved(ctx, remainOrderRemovePreRefund)
	if err != nil {
		return Order{}, err
	}

	if err := refund.RefundCalculate(ctx, order, remainOrderRemoveNewRefund); err != nil {
		return Order{}, err
	}

	if err := refund.DistributeCartOfferPrice(ctx, order); err != nil {
		return Order{}, err
	}

	if err := refund.RefundMileageCalculate(ctx, order, remainOrderRemoveNewRefund); err != nil {
		return Order{}, err
	}

	if err := RefundPriceValidate(refund); err != nil {
		return Order{}, err
	}

	if err := CompareRefundPriceValidate(ctx, refund, preRefunds, order); err != nil {
		return Order{}, err
	}

	return remainOrderRemoveNewRefund, nil
}

func (refund *Refund) OrderByPreRefundRemoved(ctx context.Context, preRefunds []Refund, order Order) (Order, error) {
	remainOrderRemovePreRefund := Order{}
	remainOrderRemovePreRefund.CustomerId = order.CustomerId
	remainOrderRemovePreRefund.TotalListPrice = number.ToFixed(order.TotalListPrice, nil)
	remainOrderRemovePreRefund.TotalSalePrice = number.ToFixed(order.TotalSalePrice, nil)
	remainOrderRemovePreRefund.TotalDiscountPrice = number.ToFixed(order.TotalDiscountPrice, nil)
	remainOrderRemovePreRefund.TotalPaymentPrice = number.ToFixed(order.TotalPaymentPrice, nil)
	remainOrderRemovePreRefund.FreightPrice = number.ToFixed(order.FreightPrice, nil)
	remainOrderRemovePreRefund.Mileage = number.ToFixed(order.Mileage, nil)
	remainOrderRemovePreRefund.MileagePrice = number.ToFixed(order.MileagePrice, nil)
	remainOrderRemovePreRefund.CashPrice = number.ToFixed(order.CashPrice, nil)

	if len(preRefunds) == 0 {
		for _, offer := range order.Offers {
			remainOrderRemovePreRefund.Offers = append(remainOrderRemovePreRefund.Offers, offer)
		}
		for _, item := range order.Items {
			remainOrderRemovePreRefund.Items = append(remainOrderRemovePreRefund.Items, item)
		}
		return remainOrderRemovePreRefund, nil
	}

	for _, orderItem := range order.Items {
		for _, preRefund := range preRefunds {
			for _, refundItem := range preRefund.Items {
				if orderItem.OrderId == refundItem.OrderId && orderItem.Id == refundItem.OrderItemId {
					refundItemPrice := RefundCalculatePrice(refundItem.ListPrice, refundItem.SalePrice, refundItem.Quantity, float64(0), float64(0))
					orderItem.Quantity -= refundItemPrice.Quantity
					orderItem.TotalListPrice = number.ToFixed(orderItem.TotalListPrice-refundItemPrice.TotalListPrice, nil)
					orderItem.TotalSalePrice = number.ToFixed(orderItem.TotalSalePrice-refundItemPrice.TotalSalePrice, nil)
					orderItem.TotalDiscountPrice = number.ToFixed(orderItem.TotalDiscountPrice-refundItemPrice.TotalDiscountPrice, nil)
					orderItem.TotalPaymentPrice = number.ToFixed(orderItem.TotalPaymentPrice-refundItemPrice.TotalRefundPrice, nil)
				}
			}

		}
		if orderItem.Quantity > 0 {
			remainOrderRemovePreRefund.Items = append(remainOrderRemovePreRefund.Items, orderItem)
		}
	}
	for _, offer := range order.Offers {
		remainOfferPrice := offer.Price
		logrus.WithField("Offer", "OfferNo="+offer.OfferNo+", OfferPrice="+strconv.FormatFloat(offer.Price, 'f', 2, 64)).Info("OriginOrder")
		for _, preRefund := range preRefunds {
			for _, refundOffer := range preRefund.Offers {
				if offer.OrderId == refundOffer.OrderId && offer.OfferNo == refundOffer.OfferNo {
					remainOfferPrice -= refundOffer.Price
					remainOfferPrice = number.ToFixed(remainOfferPrice, nil)
				}
			}
		}
		//生成订单时有使用促销退货时也要添加促销，即使促销金额是0
		offerApply := false
		for _, item := range remainOrderRemovePreRefund.Items {
			if IsItemIdInOffer(item.Id, offer) {
				offerApply = true
			}
		}
		if offerApply {
			offer.Id = 0
			offer.Price = remainOfferPrice
			remainOrderRemovePreRefund.Offers = append(remainOrderRemovePreRefund.Offers, offer)
		}
	}
	for _, preRefund := range preRefunds {
		remainOrderRemovePreRefund.FreightPrice = number.ToFixed(remainOrderRemovePreRefund.FreightPrice-preRefund.FreightPrice, nil)
		remainOrderRemovePreRefund.TotalListPrice = number.ToFixed(remainOrderRemovePreRefund.TotalListPrice-preRefund.TotalListPrice, nil)
		remainOrderRemovePreRefund.TotalSalePrice = number.ToFixed(remainOrderRemovePreRefund.TotalSalePrice-preRefund.TotalSalePrice, nil)
		remainOrderRemovePreRefund.TotalDiscountPrice = number.ToFixed(remainOrderRemovePreRefund.TotalDiscountPrice-preRefund.TotalDiscountPrice, nil)
		remainOrderRemovePreRefund.TotalPaymentPrice = number.ToFixed(remainOrderRemovePreRefund.TotalPaymentPrice-preRefund.TotalRefundPrice, nil)
		remainOrderRemovePreRefund.CashPrice = number.ToFixed(remainOrderRemovePreRefund.CashPrice-preRefund.CashPrice, nil)
		if order.Mileage > 0 {
			remainOrderRemovePreRefund.Mileage = number.ToFixed(remainOrderRemovePreRefund.Mileage-preRefund.Mileage, nil)
			remainOrderRemovePreRefund.MileagePrice = number.ToFixed(remainOrderRemovePreRefund.MileagePrice-preRefund.MileagePrice, nil)
		}
	}
	return remainOrderRemovePreRefund, nil
}

func (newRefund *Refund) OrderByNewRefundRemoved(ctx context.Context, remainOrderRemovePreRefund Order) (Order, error) {
	remainOrderRemoveNewRefund := Order{}
	remainOrderRemoveNewRefund.CustomerId = remainOrderRemovePreRefund.CustomerId
	remainOrderRemoveNewRefund.TotalListPrice = number.ToFixed(remainOrderRemovePreRefund.TotalListPrice, nil)
	remainOrderRemoveNewRefund.TotalSalePrice = number.ToFixed(remainOrderRemovePreRefund.TotalSalePrice, nil)
	remainOrderRemoveNewRefund.TotalDiscountPrice = number.ToFixed(remainOrderRemovePreRefund.TotalDiscountPrice, nil)
	remainOrderRemoveNewRefund.TotalPaymentPrice = number.ToFixed(remainOrderRemovePreRefund.TotalPaymentPrice, nil)
	remainOrderRemoveNewRefund.FreightPrice = number.ToFixed(remainOrderRemovePreRefund.FreightPrice, nil)
	remainOrderRemoveNewRefund.Mileage = number.ToFixed(remainOrderRemovePreRefund.Mileage, nil)
	remainOrderRemoveNewRefund.MileagePrice = number.ToFixed(remainOrderRemovePreRefund.MileagePrice, nil)
	remainOrderRemoveNewRefund.CashPrice = number.ToFixed(remainOrderRemovePreRefund.CashPrice, nil)
	remainOrderRemoveNewRefund.Offers = remainOrderRemovePreRefund.Offers

	for i := range remainOrderRemoveNewRefund.Offers {
		remainOrderRemoveNewRefund.Offers[i].Id = 0
	}

	for _, orderItem := range remainOrderRemovePreRefund.Items {
		remainItemSeparates := []OrderItemSeparate{}
		for _, refundItem := range newRefund.Items {
			refundItemPrice := RefundCalculatePrice(refundItem.ListPrice, refundItem.SalePrice, refundItem.Quantity, float64(0), float64(0))
			if refundItem.OrderId == orderItem.OrderId && refundItem.OrderItemId == orderItem.Id {
				if len(orderItem.ItemSeparates) == 0 {
					orderItem.Quantity -= refundItemPrice.Quantity
					remainOrderRemoveNewRefund.TotalListPrice = number.ToFixed(remainOrderRemoveNewRefund.TotalListPrice-refundItemPrice.TotalListPrice, nil)
				} else {
					for _, separate := range orderItem.ItemSeparates {
						if separate.Id == refundItem.SeparateId {
							orderItem.Quantity -= refundItemPrice.Quantity
							orderItem.TotalListPrice = number.ToFixed(orderItem.TotalListPrice-refundItemPrice.TotalListPrice, nil)
							orderItem.TotalSalePrice = number.ToFixed(orderItem.TotalSalePrice-refundItemPrice.TotalSalePrice, nil)
							orderItem.TotalDiscountPrice = number.ToFixed(orderItem.TotalDiscountPrice-refundItemPrice.TotalDiscountPrice, nil)
							orderItem.TotalPaymentPrice = number.ToFixed(orderItem.TotalPaymentPrice-refundItemPrice.TotalRefundPrice, nil)
							remainOrderRemoveNewRefund.TotalListPrice = number.ToFixed(remainOrderRemoveNewRefund.TotalListPrice-refundItemPrice.TotalListPrice, nil)
						} else {
							remainItemSeparates = append(remainItemSeparates, separate)
						}
					}
				}
			}
		}
		if orderItem.Quantity > 0 {
			orderItem.ItemSeparates = nil
			orderItem.ItemSeparates = remainItemSeparates
			remainOrderRemoveNewRefund.Items = append(remainOrderRemoveNewRefund.Items, orderItem)
		}
	}

	if remainOrderRemoveNewRefund.TotalListPrice == 0 {
		newRefund.TotalListPrice = remainOrderRemovePreRefund.TotalListPrice
		newRefund.TotalSalePrice = remainOrderRemovePreRefund.TotalSalePrice
		newRefund.TotalDiscountPrice = remainOrderRemovePreRefund.TotalDiscountPrice
		newRefund.TotalRefundPrice = remainOrderRemovePreRefund.TotalPaymentPrice
		newRefund.CashPrice = remainOrderRemovePreRefund.CashPrice
		newRefund.Offers = remainOrderRemovePreRefund.Offers
	}
	return remainOrderRemoveNewRefund, nil
}

func (refund *Refund) DistributeCartOfferPrice(ctx context.Context, order Order) error {
	if len(refund.Offers) == 0 {
		return nil
	}
	getItemIds := func(offer OrderOffer) ([]string, []string) {
		itemIds := make([]string, 0)
		targetItemIds := make([]string, 0)
		if offer.TargetItemIds != "" {
			targetItemIds = strings.Split(offer.TargetItemIds, ",")
		}
		if offer.ItemIds != "" {
			itemIds = strings.Split(offer.ItemIds, ",")
		}
		return itemIds, targetItemIds
	}
	getTotalSalePrice := func(offer OrderOffer, itemIds []string, targetItemIds []string) float64 {
		totalSaleprice := float64(0)
		for _, refundItem := range refund.Items {
			if offer.TargetType != "" {
				if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), targetItemIds) {
					totalSaleprice += refundItem.TotalSalePrice
				}
			} else {
				if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), itemIds) {
					totalSaleprice += refundItem.TotalSalePrice
				}
			}
		}
		return number.ToFixed(totalSaleprice, nil)
	}
	isLastItem := func(i int, IsTarget bool, itemIds, targetItemIds []string) bool {
		if i == len(refund.Items)-1 {
			return true
		}
		remainItems := refund.Items[i:]
		for _, refundItem := range remainItems {
			if IsTarget {
				if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), targetItemIds) {
					return false
				}
			} else {
				if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), itemIds) {
					return false
				}
			}
		}
		return true
	}
	getRemainOfferPrice := func(i int, offer OrderOffer) (remainPrice float64) {
		remainPrice = offer.Price
		if i == 0 {
			return
		}
		beforeItems := refund.Items[:i]
		for _, refundItem := range beforeItems {
			for _, appliedOffer := range refundItem.GroupOffers {
				if appliedOffer.OfferNo == offer.OfferNo {
					remainPrice = number.ToFixed(remainPrice-appliedOffer.Price, nil)
				}
			}
		}
		return
	}
	numberSetting, _, err := GetRoundSetting(ctx, refund.StoreId)
	if err != nil {
		return err
	}
	for i, refundItem := range refund.Items {
		totalDistributedCartOfferPrice := float64(0)
		for _, offer := range refund.Offers {
			itemIds, targetItemIds := getItemIds(offer)
			if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), append(itemIds, targetItemIds...)) {
				itemAppliedCartOffer := ItemAppliedCartOffer{}
				itemAppliedCartOffer.OfferNo = offer.OfferNo
				itemAppliedCartOffer.TargetType = offer.TargetType
				itemAppliedCartOffer.OrderId = offer.OrderId
				itemAppliedCartOffer.OrderItemId = refundItem.OrderItemId
				totalSalePrice := getTotalSalePrice(offer, itemIds, targetItemIds)
				offerPrice := float64(0)
				if offer.TargetType == "" {
					if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), itemIds) {
						if isLastItem(i, false, itemIds, targetItemIds) {
							offerPrice = getRemainOfferPrice(i, offer)
						} else if totalSalePrice != 0 {
							offerPrice = number.ToFixed(offer.Price*refundItem.TotalSalePrice/totalSalePrice, numberSetting)
						}
					}
				} else {
					if StringInArr(strconv.FormatInt(refundItem.OrderItemId, 10), targetItemIds) {
						if isLastItem(i, true, itemIds, targetItemIds) {
							offerPrice = getRemainOfferPrice(i, offer)
						} else if totalSalePrice != 0 {
							offerPrice = number.ToFixed(offer.Price*refundItem.TotalSalePrice/totalSalePrice, numberSetting)
						}
						itemAppliedCartOffer.IsTarget = true
					}
				}
				itemAppliedCartOffer.Price = offerPrice
				totalDistributedCartOfferPrice += offerPrice
				refund.Items[i].GroupOffers = append(refund.Items[i].GroupOffers, itemAppliedCartOffer)
			}
		}
		refund.Items[i].TotalDistributedCartOfferPrice = number.ToFixed(totalDistributedCartOfferPrice, nil)
	}
	return nil
}

func (refund *Refund) RefundMileageCalculate(ctx context.Context, order Order, remainOrder Order) error {
	param := Mileage{}
	param.TenantCode = order.TenantCode
	param.StoreId = order.StoreId
	param.TradeNo = strconv.FormatInt(order.Id, 10)
	param.MemberId = order.CustomerId
	for _, item := range refund.Items {
		dtl := MileageItem{}
		dtl.ItemId = item.OrderItemId
		dtl.TotalListAmount = item.TotalListPrice
		param.Items = append(param.Items, dtl)
	}
	mileage, err := benefitHandler.GetRefundMileage(ctx, param)
	if err != nil {
		return err
	}
	refund.Mileage = math.Abs(mileage.Point)
	refund.MileagePrice = math.Abs(mileage.PointPrice)
	refund.ObtainMileage = math.Abs(mileage.ObtainPoint)
	refund.CashPrice = number.ToFixed(refund.TotalRefundPrice-refund.MileagePrice, nil)
	if remainOrder.CashPrice < refund.CashPrice {
		return errors.New("Refund Cash Price Greate than Remain CashPrice, Remain CashPrice=" + strconv.FormatFloat(remainOrder.CashPrice, 'f', 2, 64) + ", newRefundCashPrice=" + strconv.FormatFloat(refund.CashPrice, 'f', 2, 64))
	}
	for i, _ := range refund.Items {
		for _, item := range mileage.Items {
			if refund.Items[i].OrderItemId == item.ItemId && refund.Items[i].TotalListPrice == item.TotalListAmount {
				refund.Items[i].ObtainMileage = math.Abs(item.ObtainPoint)
				refund.Items[i].Mileage = math.Abs(item.Point)
				refund.Items[i].MileagePrice = math.Abs(item.PointPrice)
				break
			}
		}
	}
	return nil
}
func RefundPriceValidate(refund *Refund) error {
	totalRefundDiscountPrice := float64(0.00)
	totalRefundPrice := float64(0.00)
	refundFreightPrice := refund.FreightPrice
	fmt.Println("RefundPriceValidate refund==", JsonToString(refund))
	logrus.WithField("FreightPrice", refundFreightPrice).Info("refund")
	for i := range refund.Items {
		totalRefundItemPrice := number.ToFixed(refund.Items[i].SalePrice*float64(refund.Items[i].Quantity), nil)
		totalRefundItemDiscountPrice := number.ToFixed((refund.Items[i].ListPrice-refund.Items[i].SalePrice)*float64(refund.Items[i].Quantity), nil)
		if refund.Items[i].TotalRefundPrice != totalRefundItemPrice {
			return RefundPriceError
		}
		if refund.Items[i].TotalDiscountPrice != totalRefundItemDiscountPrice {
			return RefundPriceError
		}
		totalRefundPrice += totalRefundItemPrice
		totalRefundDiscountPrice += totalRefundItemDiscountPrice
	}
	totalRefundDiscountPrice = number.ToFixed(totalRefundDiscountPrice, nil)
	logrus.WithField("totalRefundDiscountPrice", totalRefundDiscountPrice).Info("totalRefundDiscountPrice")
	totalRefundPrice = number.ToFixed(totalRefundPrice, nil)
	logrus.WithField("totalRefundPrice", totalRefundPrice).Info("totalRefundPrice")
	totalRefundPrice = number.ToFixed(totalRefundPrice+refundFreightPrice, nil)
	logrus.WithField("totalRefundPrice", totalRefundPrice).Info("totalRefundPrice")
	for _, offer := range refund.Offers {
		totalRefundDiscountPrice = number.ToFixed(totalRefundDiscountPrice+offer.Price, nil)
		totalRefundPrice = number.ToFixed(totalRefundPrice-offer.Price, nil)
	}
	logrus.WithField("refund.TotalDiscountPrice", refund.TotalDiscountPrice).Info("totalDiscountPrice")
	logrus.WithField("totalRefundDiscountPrice", totalRefundDiscountPrice).Info("totalRefundDiscountPrice")
	if refund.TotalDiscountPrice != totalRefundDiscountPrice {
		return RefundPriceError
	}

	logrus.WithField("refund.totalRefundPrice", refund.TotalRefundPrice).Info("refund.TotalRefundPrice")
	logrus.WithField("totalRefundPrice", totalRefundPrice).Info("totalRefundPrice")
	if refund.TotalRefundPrice != totalRefundPrice {
		return RefundPriceError
	}

	return nil
}

func CompareRefundPriceValidate(ctx context.Context, newRefund *Refund, preRefunds []Refund, order Order) error {
	orderItemTotalPaymentPrice := float64(0.00)
	orderItemTotalDiscountPrice := float64(0.00)
	for _, orderItem := range order.Items {
		orderItemTotalPaymentPrice += number.ToFixed(orderItem.TotalPaymentPrice, nil)
		orderItemTotalDiscountPrice += number.ToFixed(orderItem.TotalDiscountPrice, nil)
	}
	orderItemTotalPaymentPrice = number.ToFixed(orderItemTotalPaymentPrice, nil)
	orderItemTotalDiscountPrice = number.ToFixed(orderItemTotalDiscountPrice, nil)
	newTotalRefundPrice := number.ToFixed(newRefund.TotalRefundPrice, nil)
	newTotalDiscountPrice := number.ToFixed(newRefund.TotalDiscountPrice, nil)
	newRefundItemPrice := float64(0.00)
	newRefundItemTotalDiscountPrice := float64(0.00)
	for _, item := range newRefund.Items {
		newRefundItemPrice += number.ToFixed(item.TotalRefundPrice, nil)
		newRefundItemTotalDiscountPrice += number.ToFixed(item.TotalDiscountPrice, nil)
	}
	newTotalRefundPrice = number.ToFixed(newTotalRefundPrice, nil)
	newTotalDiscountPrice = number.ToFixed(newTotalDiscountPrice, nil)
	newRefundItemPrice = number.ToFixed(newRefundItemPrice, nil)
	newRefundItemTotalDiscountPrice = number.ToFixed(newRefundItemTotalDiscountPrice, nil)
	if orderItemTotalPaymentPrice < number.ToFixed(newRefundItemPrice, nil) {
		return errors.New("Refund Item Price Greate then equal Order Item Price, orderItemTotalPaymentPrice=" + strconv.FormatFloat(orderItemTotalPaymentPrice, 'f', 2, 64) + ", newRefundItemPrice=" + strconv.FormatFloat(newRefundItemPrice, 'f', 2, 64))
	}
	if orderItemTotalDiscountPrice < number.ToFixed(newRefundItemTotalDiscountPrice, nil) {
		return errors.New("Refund Item Offer Price Greate then equal Order Item Offer Price, orderItemTotalDiscountPrice=" + strconv.FormatFloat(orderItemTotalDiscountPrice, 'f', 2, 64) + ", newRefundItemTotalDiscountPrice=" + strconv.FormatFloat(newRefundItemTotalDiscountPrice, 'f', 2, 64))
	}
	if order.TotalPaymentPrice < number.ToFixed(newTotalRefundPrice, nil) {
		return errors.New("Refund Sum Price Greate then equal Order Price, order.TotalPaymentPrice=" + strconv.FormatFloat(order.TotalPaymentPrice, 'f', 2, 64) + ", newTotalRefundPrice=" + strconv.FormatFloat(newTotalRefundPrice, 'f', 2, 64))
	}
	if order.TotalDiscountPrice < number.ToFixed(newTotalDiscountPrice, nil) {
		return errors.New("Refund Sum Coupon Price Greate then equal Order Total Discount Price, order.TotalDiscountPrice=" + strconv.FormatFloat(order.TotalDiscountPrice, 'f', 2, 64) + ", newTotalDiscountPrice=" + strconv.FormatFloat(newTotalDiscountPrice, 'f', 2, 64))
	}
	return nil
}

func (o *Refund) Save(ctx context.Context) error {
	o.CreatedAt = time.Now().UTC()
	o.UpdatedAt = o.CreatedAt
	if err := o.insert(ctx); err != nil {
		return err
	}
	o.Extension.RefundId = o.Id
	if err := o.Extension.insert(ctx); err != nil {
		return err
	}
	for i := range o.Items {
		o.Items[i].RefundId = o.Id
		o.Items[i].CreatedAt = o.CreatedAt
		o.Items[i].UpdatedAt = o.CreatedAt
		if err := o.Items[i].insert(ctx); err != nil {
			return err
		}
	}

	for i := range o.Offers {
		o.Offers[i].OrderId = o.OrderId
		o.Offers[i].RefundId = o.Id
		o.Offers[i].CreatedAt = o.CreatedAt
		if err := o.Offers[i].Save(ctx); err != nil {
			return err
		}
	}

	userClaim := auth.UserClaim{}.FromCtx(ctx)
	refundHistory := o.NewRefundHistory(userClaim)
	if err := refundHistory.Save(ctx); err != nil {
		return err
	}
	if o.RefundType == "POS" && o.Status == enum.RefundOrderRegistered.String() {
		if _, err := o.ChangeStatus(ctx, enum.RefundOrderProcessing.String()); err != nil {
			return err
		}
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
	// go func() {
	o.addEventToOutBox(enum.RefundOrderRegistered.String())
	o.PublishEventMessagesInOutBox(ctx)
	// }()

	return nil
}

func (refund *Refund) insert(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(refund)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}

	return nil
}
func (extension *RefundExtension) insert(ctx context.Context) error {
	row, err := factory.DB(ctx).Insert(extension)
	if err != nil {
		return err
	}
	if int(row) == 0 {
		return InsertNotFoundError
	}
	return nil
}
func (refundItem *RefundItem) insert(ctx context.Context) error {
	if _, err := factory.DB(ctx).Insert(refundItem); err != nil {
		return err
	}
	for i := range refundItem.GroupOffers {
		refundItem.GroupOffers[i].OrderId = refundItem.OrderId
		refundItem.GroupOffers[i].OrderItemId = refundItem.OrderItemId
		refundItem.GroupOffers[i].RefundId = refundItem.RefundId
		refundItem.GroupOffers[i].RefundItemId = refundItem.Id
		if err := refundItem.GroupOffers[i].Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
