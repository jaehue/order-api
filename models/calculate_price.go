package models

import (
	"github.com/pangpanglabs/goutils/number"
)

type OrderPrice struct {
	TotalListPrice     float64 `json:"totalListPrice"`
	TotalSalePrice     float64 `json:"totalSalePrice"`
	TotalDiscountPrice float64 `json:"totalDiscountPrice"`
	FreightPrice       float64 `json:"freightPrice"`
	TotalPaymentPrice  float64 `json:"totalPaymentPrice"`
	MileagePrice       float64 `json:"mileagePrice"`
	CashPrice          float64 `json:"cashPrice"`
	Quantity           int     `json:"quantity"`
}

func OrderCalculatePrice(listPrice, salePrice float64, quantity int, freightPrice float64, mileagePrice float64) OrderPrice {
	totalListPrice := number.ToFixed(listPrice*float64(quantity), nil)
	totalSalePrice := number.ToFixed(salePrice*float64(quantity), nil)
	totalDiscountPrice := number.ToFixed(totalListPrice-totalSalePrice, nil)
	freightPrice = number.ToFixed(freightPrice, nil)
	totalPaymentPrice := number.ToFixed(totalSalePrice+freightPrice, nil)
	totalCashPrice := number.ToFixed(totalPaymentPrice-mileagePrice, nil)

	return OrderPrice{
		Quantity:           quantity,
		TotalListPrice:     totalListPrice,
		TotalSalePrice:     totalSalePrice,
		TotalDiscountPrice: totalDiscountPrice,
		FreightPrice:       freightPrice,
		TotalPaymentPrice:  totalPaymentPrice,
		MileagePrice:       mileagePrice,
		CashPrice:          totalCashPrice,
	}
}

type RefundPrice struct {
	TotalListPrice     float64 `json:"totalListPrice"`
	TotalSalePrice     float64 `json:"totalSalePrice"`
	TotalDiscountPrice float64 `json:"totalDiscountPrice"`
	FreightPrice       float64 `json:"freightPrice"`
	TotalRefundPrice   float64 `json:"totalRefundPrice"`
	MileagePrice       float64 `json:"mileagePrice"`
	CashPrice          float64 `json:"cashPrice"`
	Quantity           int     `json:"quantity"`
}

func RefundCalculatePrice(listPrice, salePrice float64, quantity int, freightPrice float64, mileagePrice float64) RefundPrice {
	totalListPrice := number.ToFixed(listPrice*float64(quantity), nil)
	totalSalePrice := number.ToFixed(salePrice*float64(quantity), nil)
	totalDiscountPrice := number.ToFixed(totalListPrice-totalSalePrice, nil)
	freightPrice = number.ToFixed(freightPrice, nil)
	totalRefundPrice := number.ToFixed(totalSalePrice, nil)
	totalCashPrice := number.ToFixed(totalRefundPrice-mileagePrice, nil)

	return RefundPrice{
		Quantity:           quantity,
		TotalListPrice:     totalListPrice,
		TotalSalePrice:     totalSalePrice,
		TotalDiscountPrice: totalDiscountPrice,
		FreightPrice:       freightPrice,
		TotalRefundPrice:   totalRefundPrice,
		MileagePrice:       mileagePrice,
		CashPrice:          totalCashPrice,
	}
}
