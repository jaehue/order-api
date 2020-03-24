package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/common/api"

	"github.com/hublabs/order-api/adapters"
	"github.com/hublabs/order-api/enum"

	"github.com/pangpanglabs/goutils/number"
)

type CalculateRefundRequest struct {
	Original struct {
		Items      []CalculateRefundItemRequest
		SaleType   string
		Offers     []CalculateOffer
		CustomerId int64
	}
	RemainItems []CalculateRefundItemRequest
	SellingAt   time.Time `json:"sellingAt"`
	StoreId     int64     `json:"storeId"`
}

type CalculateRefundItemRequest struct {
	Code     string
	Quantity int
	OfferNo  string
}

func GenerateRefundCalculateRequest(order Order, newOrder Order) CalculateRefundRequest {
	var calculateRefundRequest CalculateRefundRequest
	calculateRefundRequest.Original.CustomerId = order.CustomerId
	calculateRefundRequest.Original.SaleType = order.SaleType
	calculateRefundRequest.SellingAt = order.CreatedAt
	calculateRefundRequest.StoreId = order.StoreId
	for _, item := range order.Items {
		var calculateItemRequest CalculateRefundItemRequest
		calculateItemRequest.Code = item.ItemCode
		calculateItemRequest.Quantity = item.Quantity
		calculateRefundRequest.Original.Items = append(calculateRefundRequest.Original.Items, calculateItemRequest)
	}
	for _, offer := range order.Offers {
		calculateOffer := CalculateOffer{}
		calculateOffer.No = offer.OfferNo
		calculateOffer.CouponNo = offer.CouponNo
		calculateRefundRequest.Original.Offers = append(calculateRefundRequest.Original.Offers, calculateOffer)
	}
	for _, item := range newOrder.Items {
		var calculateItemRequest CalculateRefundItemRequest
		calculateItemRequest.Code = item.ItemCode
		calculateItemRequest.Quantity = item.Quantity
		for _, offer := range order.Offers {
			if offer.AdditionalType == "" && offer.ItemIds != "" {
				if strings.Index(offer.ItemIds+","+offer.TargetItemIds+",", strconv.FormatInt(item.Id, 10)+",") > -1 {
					calculateItemRequest.OfferNo = offer.OfferNo
					break
				}
			}
		}
		calculateRefundRequest.RemainItems = append(calculateRefundRequest.RemainItems, calculateItemRequest)
	}

	return calculateRefundRequest
}

func (refund *Refund) RefundCalculate(ctx context.Context, order Order, newOrder Order) error {
	if len(newOrder.Items) == 0 {
		//未发货的情况全部退款的时候退运费
		if order.Status == enum.SaleOrderFinished.String() || order.Status == enum.StockDistributed.String() {
			refund.FreightPrice = number.ToFixed(newOrder.FreightPrice, nil)
		} else {
			refund.TotalRefundPrice = number.ToFixed(refund.TotalRefundPrice-newOrder.FreightPrice, nil)
		}
		return nil
	}
	calculateRemainResult, err := calculatorEventHandler.GetRefundCalculate(ctx, order, newOrder)
	if err != nil {
		return err
	}
	//运费为之前的剩余额度和最后剩余部分支付运费的差额
	//发货之前直接计算运费多退少补;发货后部分退货后需补运费，不退运费
	if order.Status == enum.SaleOrderFinished.String() || order.Status == enum.StockDistributed.String() || newOrder.FreightPrice < calculateRemainResult.FreightPrice {
		refund.FreightPrice = number.ToFixed(newOrder.FreightPrice-calculateRemainResult.FreightPrice, nil)
		refund.TotalRefundPrice = number.ToFixed(refund.TotalRefundPrice+refund.FreightPrice, nil)
	}
	if err := remainedOrderItemValidateAsResultOfCalculator(newOrder, calculateRemainResult); err != nil {
		return err
	}
	if err := refund.refundPriceAsResultOfCalculator(newOrder, calculateRemainResult); err != nil {
		return err
	}

	return nil
}

func remainedOrderItemValidateAsResultOfCalculator(newOrder Order, calculateRemainResult OrderCalculateResult) error {
	for i, item := range newOrder.Items {
		var quantity int
		for _, group := range calculateRemainResult.Groups {
			for _, groupItem := range group.Items {
				if i == groupItem.Seq {
					if item.ItemCode != groupItem.Code {
						return errors.New("Item ItemCode Error, item.ItemCode=" + item.ItemCode +
							", groupItem.Code=" + groupItem.Code)
					}
					if item.ListPrice != groupItem.Sku.Product.ListPrice {
						return errors.New("Item ListPrice Error, item.ListPrice=" + strconv.FormatFloat(item.ListPrice, 'f', 2, 64) + ", groupItem.Sku.Product.ListPrice=" + strconv.FormatFloat(groupItem.Sku.Product.ListPrice, 'f', 2, 64))
					}
					if item.SalePrice != groupItem.SalePrice {
						return errors.New("Item SalePrice Error, item.SalePrice=" + strconv.FormatFloat(item.SalePrice, 'f', 2, 64) + ", groupItem.SalePrice=" + strconv.FormatFloat(groupItem.SalePrice, 'f', 2, 64))
					}
					quantity += groupItem.Quantity
				}
			}
		}
		if item.Quantity != quantity {
			return errors.New("Item Quantity Error, item.Quantity=" + strconv.Itoa(item.Quantity) +
				", sum(groupItem.Quantity)=" + strconv.Itoa(quantity))
		}
	}
	return nil
}

func (refund *Refund) refundPriceAsResultOfCalculator(newOrder Order, calculateRemainResult OrderCalculateResult) error {
	for _, offer := range newOrder.Offers {
		for _, group := range calculateRemainResult.Groups {
			if offer.CouponNo == group.Offer.CouponNo && offer.OfferNo == group.Offer.OriginalNo && group.Offer.Valid {
				if offer.Price < group.Offer.TotalDiscountPrice {
					return errors.New("Offer Error, Order Offer Price=" + strconv.FormatFloat(offer.Price, 'f', 2, 64) + ",Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(group.Offer.TotalDiscountPrice, 'f', 2, 64))
				}
				offer.Price = number.ToFixed(offer.Price-group.Offer.TotalDiscountPrice, nil)
				break
			}
		}
		for _, calculateOffer := range calculateRemainResult.Offers {
			if offer.CouponNo == calculateOffer.CouponNo && offer.OfferNo == calculateOffer.OriginalNo && calculateOffer.Valid {
				if offer.Price < calculateOffer.TotalDiscountPrice {
					return errors.New("Offer Error, Order Offer Price=" + strconv.FormatFloat(offer.Price, 'f', 2, 64) +
						",Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(calculateOffer.TotalDiscountPrice, 'f', 2, 64))
				}
				offer.Price = number.ToFixed(offer.Price-calculateOffer.TotalDiscountPrice, nil)
				break
			}
		}
		offerApply := false
		for _, item := range refund.Items {
			if IsItemIdInOffer(item.OrderItemId, offer) {
				offerApply = true
			}
		}
		if offerApply {
			refund.TotalSalePrice = number.ToFixed(refund.TotalSalePrice-offer.Price, nil)
			refund.TotalDiscountPrice = number.ToFixed(refund.TotalDiscountPrice+offer.Price, nil)
			refund.TotalRefundPrice = number.ToFixed(refund.TotalRefundPrice-offer.Price, nil)
			offer.Price = number.ToFixed(offer.Price, nil)
			refund.Offers = append(refund.Offers, offer)
		}
	}

	if refund.TotalRefundPrice < 0 {
		return errors.New("Refund Price Less than 0, refund.TotalRefundPrice=" + strconv.FormatFloat(refund.TotalRefundPrice, 'f', 2, 64))
	}

	return nil
}

func (c CalculatorEventHandler) GetRefundCalculate(ctx context.Context, order Order, newOrder Order) (OrderCalculateResult, error) {
	var resp struct {
		Success bool
		Result  struct {
			// Original RefundCalculateResult
			Remain OrderCalculateResult
		}
		Error api.Error
	}
	var param = GenerateRefundCalculateRequest(order, newOrder)
	LogParamErrorMsg("GetRefundCalculate", param, api.Error{})
	url := c.CalculatorApiUrl + "/v1/refund/"
	err := adapters.RetryRestApi(ctx, &resp, http.MethodPost, url, param)
	if err != nil {
		return resp.Result.Remain, err
	}
	if !resp.Success || resp.Error.Code != 0 {
		LogParamErrorMsg("GetRefundCalculate", param, resp.Error)
		return resp.Result.Remain, fmt.Errorf("%d#%s#%s", resp.Error.Code, resp.Error.Message, resp.Error.Details)
	}
	return resp.Result.Remain, nil
}
