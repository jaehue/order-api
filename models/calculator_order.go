package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hublabs/common/api"

	"github.com/hublabs/order-api/adapters"

	"github.com/pangpanglabs/goutils/number"
)

type CalculatorEventHandler struct {
	CalculatorApiUrl string
}

type CalculateItemRequest struct {
	Code     string
	Quantity int
	OfferNo  string
}
type CalculateOffer struct {
	No       string
	CouponNo string
}
type CalculateRequest struct {
	Items        []CalculateItemRequest
	Offers       []CalculateOffer
	CustomerId   int64
	StoreId      int64
	Mileage      float64 `json:"mileage"`
	MileagePrice float64 `json:"mileagePrice"`
	CouponNos    []string
	SaleType     string
	IsManually   bool
}

type OrderCalculateResult struct {
	TotalListPrice        float64 `json:"totalListPrice"`
	TotalSalePrice        float64 `json:"totalSalePrice"`     //单品促销后的总售价
	TotalDiscountPrice    float64 `json:"totalDiscountPrice"` //单品促销的总优惠金额
	DiscountRate          float64 `json:"discountRate"`
	Quantity              int     `json:"quantity"`
	SubTotalSalePrice     float64 `json:"subTotalSalePrice"`     //总售价（单品和购物车促销优惠后）
	SubTotalDiscountPrice float64 `json:"subTotalDiscountPrice"` //总优惠金额（单品和购物车促销）
	ObtainMileage         float64 `json:"obtainMileage"`
	Mileage               float64 `json:"mileage"`
	MileagePrice          float64 `json:"mileagePrice"`
	FreightPrice          float64 `json:"freightPrice"`
	Groups                []struct {
		TotalListPrice float64 `json:"totalListPrice"`
		TotalSalePrice float64 `json:"totalSalePrice"`
		DiscountRate   float64 `json:"discountRate"`
		Quantity       int     `json:"quantity"`
		Items          []struct {
			Code    string  `json:"code"`
			Seq     int     `json:"seq"`
			Name    string  `json:"name"`
			OfferNo string  `json:"offerNo"`
			ItemFee float64 `json:"itemFee"`
			FeeRate float64 `json:"feeRate"`
			Sku     struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				Image   string `json:"image"`
				Options []struct {
					ID    int64  `json:"id"`
					SkuID int64  `json:"skuId"`
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"options"`
				Product struct {
					ID    int64  `json:"id"`
					Name  string `json:"name"`
					Brand struct {
						ID   int64  `json:"id"`
						Code string `json:"code"`
						Name string `json:"name"`
					} `json:"brand"`
					TitleImage string    `json:"titleImage"`
					ListPrice  float64   `json:"listPrice"`
					HasDigital bool      `json:"hasDigital"`
					Enable     bool      `json:"enable"`
					CreatedAt  time.Time `json:"createdAt"`
					UpdatedAt  time.Time `json:"updatedAt"`
				} `json:"product"`
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"sku"`
			SalePrice             float64 `json:"salePrice"`
			TotalListPrice        float64 `json:"totalListPrice"`
			TotalSalePrice        float64 `json:"totalSalePrice"`
			TotalDiscountPrice    float64 `json:"totalDiscountPrice"`
			DiscountRate          float64 `json:"discountRate"`
			ObtainMileage         float64 `json:"obtainMileage"`
			Mileage               float64 `json:"mileage"`
			MileagePrice          float64 `json:"mileagePrice"`
			Quantity              int     `json:"quantity"`
			SubTotalDiscountPrice float64 `json:"subTotalDiscountPrice"` //总优惠金额（单品和购物车促销）
			UsedOffers            []struct {
				No                 string  `json:"no"`
				TotalDiscountPrice float64 `json:"totalDiscountPrice"` //单品促销的总优惠金额
				IsTarget           bool    `json:"isTarget"`
			} `json:"UsedOffers"`
		} `json:"items"`
		Offer struct {
			No                 string   `json:"no"`
			OriginalNo         string   `json:"originalNo"`
			Name               string   `json:"name"`
			TotalDiscountPrice float64  `json:"totalDiscountPrice"`
			ItemCodes          []string `json:"itemCodes"`
			TargetItemCodes    []string `json:"targetItemCodes"`
			Valid              bool     `json:"valid"`
			RulesetGroup       struct {
				ID      int64  `json:"id"`
				Name    string `json:"name"`
				Actions []struct {
					StandardType  string  `json:"standardType"`
					StandardValue float64 `json:"standardValue"`
					DiscountType  string  `json:"discountType"`
					DiscountValue float64 `json:"discountValue"`
				} `json:"actions"`
			} `json:"rulesetGroup"`
			CouponNo string `json:"couponNo"`
		} `json:"offer"`
		TotalDiscountPrice    float64 `json:"totalDiscountPrice"`    //单品促销的总优惠金额
		SubTotalDiscountPrice float64 `json:"subTotalDiscountPrice"` //总优惠金额（单品和购物车促销）
	} `json:"groups"`
	Offers []struct {
		No                 string   `json:"no"`
		OriginalNo         string   `json:"originalNo"`
		Name               string   `json:"name"`
		TotalDiscountPrice float64  `json:"totalDiscountPrice"`
		ItemCodes          []string `json:"itemCodes"`
		Valid              bool     `json:"valid"`
		RulesetGroup       struct {
			ID      int64  `json:"id"`
			Name    string `json:"name"`
			Actions []struct {
				StandardType  string  `json:"standardType"`
				StandardValue float64 `json:"standardValue"`
				DiscountType  string  `json:"discountType"`
				DiscountValue float64 `json:"discountValue"`
			} `json:"actions"`
		} `json:"rulesetGroup"`
		CouponNo string `json:"couponNo"`
	} `json:"offers"`
}

func (calculateResult *OrderCalculateResult) offerCountAsResultOfCalculator() int {
	// applyOfferCount := len(calculateResult.Offers)
	var applyOfferCount int
	for _, group := range calculateResult.Groups {
		if group.Offer.No != "" && group.Offer.Valid { //&& group.Offer.TotalDiscountPrice != 0
			applyOfferCount++
		}
	}
	for _, offer := range calculateResult.Offers {
		if offer.No != "" && offer.Valid {
			applyOfferCount++
		}
	}
	return applyOfferCount
}

func (c CalculatorEventHandler) OrderCalculate(ctx context.Context, order *Order) error {
	param, calculateResult, err := c.GetOrderCalculate(ctx, order)
	if err != nil {
		return err
	}
	applyOfferCount := calculateResult.offerCountAsResultOfCalculator()

	if len(order.Offers) != applyOfferCount {
		LogParamErrorMsg("GetOrderCalculate", param, api.Error{})
		return errors.New("Offer Count Miss Match, Calculator Offer Count=" + strconv.Itoa(applyOfferCount) + ", Order Offer Count=" + strconv.Itoa(len(order.Offers)))
	}
	if order.FreightPrice != calculateResult.FreightPrice {
		LogParamErrorMsg("GetOrderCalculate", param, api.Error{})
		return errors.New("FreightPrice Miss Match, Calculator FreightPrice=" + strconv.FormatFloat(calculateResult.FreightPrice, 'f', 2, 64) + ", Order FreightPrice=" + strconv.FormatFloat(order.FreightPrice, 'f', 2, 64))
	}
	if order.ObtainMileage != calculateResult.ObtainMileage {
		LogParamErrorMsg("GetOrderCalculate", param, api.Error{})
		return errors.New("ObtainMileage Miss Match, Calculator ObtainMileage=" + strconv.FormatFloat(calculateResult.ObtainMileage, 'f', 2, 64) + ", Order ObtainMileage=" + strconv.FormatFloat(order.ObtainMileage, 'f', 2, 64))
	}
	if err := orderItemValidateAsResultOfCalculator(order, calculateResult); err != nil {
		LogParamErrorMsg("GetOrderCalculate", param, api.Error{})
		return err
	}

	if err := orderOfferValidateAsResultOfCalculator(order, calculateResult); err != nil {
		LogParamErrorMsg("GetOrderCalculate", param, api.Error{})
		return err
	}
	orderPrice := OrderCalculatePrice(calculateResult.TotalListPrice, calculateResult.TotalListPrice-calculateResult.SubTotalDiscountPrice, int(1), calculateResult.FreightPrice, order.MileagePrice)
	order.TotalListPrice = orderPrice.TotalListPrice
	order.TotalSalePrice = orderPrice.TotalSalePrice
	order.TotalDiscountPrice = orderPrice.TotalDiscountPrice
	order.TotalPaymentPrice = orderPrice.TotalPaymentPrice
	order.CashPrice = orderPrice.CashPrice

	return nil
}

func orderItemValidateAsResultOfCalculator(order *Order, calculateResult OrderCalculateResult) error {
	for i, item := range order.Items {
		for _, group := range calculateResult.Groups {
			for _, groupItem := range group.Items {
				if groupItem.Sku.Product.ListPrice < 0 {
					return errors.New("Item List Price Error, item.ListPrice=" + strconv.FormatFloat(item.ListPrice, 'f', 2, 64) + ", groupItem.Sku.Product.ListPrice=" + strconv.FormatFloat(groupItem.Sku.Product.ListPrice, 'f', 2, 64))
				}
				if groupItem.SalePrice < 0 {
					return errors.New("Item Sale Price Error, item.SalePrice=" + strconv.FormatFloat(item.SalePrice, 'f', 2, 64) + ", groupItem.SalePrice=" + strconv.FormatFloat(groupItem.SalePrice, 'f', 2, 64))
				}
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
					if item.Quantity != groupItem.Quantity {
						return errors.New("Item Quantity Error, item.Quantity=" + strconv.Itoa(item.Quantity) + ", groupItem.Quantity=" + strconv.Itoa(groupItem.Quantity))
					}
					if item.TotalListPrice != groupItem.TotalListPrice {
						return errors.New("Item TotalListPrice Error, item.TotalListPrice=" + strconv.FormatFloat(item.TotalListPrice, 'f', 2, 64) + ", groupItem.TotalListPrice=" + strconv.FormatFloat(groupItem.TotalListPrice, 'f', 2, 64))
					}
					if item.TotalSalePrice != groupItem.TotalSalePrice {
						return errors.New("Item TotalSalePrice Error, item.TotalSalePrice=" + strconv.FormatFloat(item.TotalSalePrice, 'f', 2, 64) + ", groupItem.TotalSalePrice=" + strconv.FormatFloat(groupItem.TotalSalePrice, 'f', 2, 64))
					}
					if item.TotalDiscountPrice != groupItem.TotalDiscountPrice {
						return errors.New("Item TotalDiscountPrice Error, item.TotalDiscountPrice=" + strconv.FormatFloat(item.TotalDiscountPrice, 'f', 2, 64) + ", groupItem.TotalDiscountPrice=" + strconv.FormatFloat(groupItem.TotalDiscountPrice, 'f', 2, 64))
					}
					if item.TotalDistributedCartOfferPrice != number.ToFixed(groupItem.SubTotalDiscountPrice-groupItem.TotalDiscountPrice, nil) {
						return errors.New("Item TotalDistributedCartOfferPrice Error, item.TotalDistributedCartOfferPrice=" +
							strconv.FormatFloat(item.TotalDistributedCartOfferPrice, 'f', 2, 64) + ", groupItem.TotalDistributedCartOfferPrice=" +
							strconv.FormatFloat(number.ToFixed(groupItem.SubTotalDiscountPrice-groupItem.TotalDiscountPrice, nil), 'f', 2, 64))
					}
					for _, itemOffer := range item.AppliedCartOffers {
						for _, groupItemOffer := range groupItem.UsedOffers {
							if itemOffer.OfferNo == groupItemOffer.No {
								if itemOffer.Price != groupItemOffer.TotalDiscountPrice {
									return errors.New("Item AppliedCartOfferPrice Error, itemOffer.Price=" +
										strconv.FormatFloat(itemOffer.Price, 'f', 2, 64) + ",  groupItemOffer.TotalDiscountPric=" +
										strconv.FormatFloat(groupItemOffer.TotalDiscountPrice, 'f', 2, 64))

								}
							}
						}
					}
					order.Items[i].ObtainMileage = groupItem.ObtainMileage
					order.Items[i].Mileage = groupItem.Mileage
					order.Items[i].MileagePrice = groupItem.MileagePrice
					order.Items[i].ItemFee = groupItem.ItemFee
					order.Items[i].FeeRate = groupItem.FeeRate
				}
			}
		}
	}
	return nil
}

func orderOfferValidateAsResultOfCalculator(order *Order, calculateResult OrderCalculateResult) error {
	for _, offer := range order.Offers {
		offerApply := false
		for _, group := range calculateResult.Groups {
			if offer.OfferNo == group.Offer.No && offer.CouponNo == group.Offer.CouponNo && group.Offer.Valid {
				if group.Offer.TotalDiscountPrice < 0 {
					return errors.New("Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(group.Offer.TotalDiscountPrice, 'f', 2, 64))
				}
				if offer.Price == group.Offer.TotalDiscountPrice {
					offerApply = true
					break
				} else {
					return errors.New("Offer Error, Order Offer Price=" + strconv.FormatFloat(offer.Price, 'f', 2, 64) + ",Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(group.Offer.TotalDiscountPrice, 'f', 2, 64))
				}
			}
		}
		if offerApply {
			continue
		}
		for _, calculateOffer := range calculateResult.Offers {
			if offer.OfferNo == calculateOffer.No && offer.CouponNo == calculateOffer.CouponNo && calculateOffer.Valid {
				if calculateOffer.TotalDiscountPrice < 0 {
					return errors.New("Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(calculateOffer.TotalDiscountPrice, 'f', 2, 64))
				}
				if offer.Price == calculateOffer.TotalDiscountPrice {
					offerApply = true
					break
				} else {
					return errors.New("Offer Error, Order Offer Price=" + strconv.FormatFloat(offer.Price, 'f', 2, 64) +
						",Calculate Offer TotalDiscountPrice=" + strconv.FormatFloat(calculateOffer.TotalDiscountPrice, 'f', 2, 64))
				}
			}
		}
		if offerApply == false {
			return errors.New("Offer Error, Order Offer No=" + offer.OfferNo + ",Calculate Offer No Nothing Match")
		}
	}
	return nil
}

func GenerateCalculateRequest(order *Order) CalculateRequest {
	var calculateRequest CalculateRequest
	calculateRequest.CustomerId = order.CustomerId
	calculateRequest.StoreId = order.StoreId
	calculateRequest.SaleType = order.SaleType
	calculateRequest.Mileage = order.Mileage
	calculateRequest.MileagePrice = order.MileagePrice
	calculateRequest.SaleType = order.SaleType
	calculateRequest.IsManually = true
	for _, orderOffer := range order.Offers {
		calculateOffer := CalculateOffer{}
		calculateOffer.No = orderOffer.OfferNo
		calculateOffer.CouponNo = orderOffer.CouponNo
		calculateRequest.Offers = append(calculateRequest.Offers, calculateOffer)
	}
	for _, item := range order.Items {
		var calculateItemRequest CalculateItemRequest
		calculateItemRequest.Code = item.ItemCode
		calculateItemRequest.Quantity = item.Quantity
		for _, itemOffer := range item.AppliedCartOffers {
			//only add not additional offerNo
			if itemOffer.AdditionalType == "" {
				calculateItemRequest.OfferNo = itemOffer.OfferNo
				break
			}
		}
		calculateRequest.Items = append(calculateRequest.Items, calculateItemRequest)
	}
	for _, offer := range order.Offers {
		if offer.CouponNo != "" {
			calculateRequest.CouponNos = append(calculateRequest.CouponNos, offer.CouponNo)
		}
	}
	return calculateRequest
}

func (c CalculatorEventHandler) GetOrderCalculate(ctx context.Context, order *Order) (CalculateRequest, OrderCalculateResult, error) {
	var resp struct {
		Success bool
		Result  OrderCalculateResult
		Error   api.Error
	}
	var param = GenerateCalculateRequest(order)
	url := c.CalculatorApiUrl + "/v1/calculate/"
	err := adapters.RetryRestApi(ctx, &resp, http.MethodPost, url, param)
	if err != nil {
		return param, OrderCalculateResult{}, err
	}
	if !resp.Success || resp.Error.Code != 0 {
		LogParamErrorMsg("GetOrderCalculate", param, resp.Error)
		return param, OrderCalculateResult{}, fmt.Errorf("%d#%s#%s", resp.Error.Code, resp.Error.Message, resp.Error.Details)
	}
	return param, resp.Result, nil
}
