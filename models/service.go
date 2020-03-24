package models

import (
	"errors"
	"fmt"

	"github.com/hublabs/order-api/factory"
)

var (
	InsertNotFoundError      = errors.New("Insert Data Not Found.")
	NotFoundDataError        = errors.New("Not found Data.")
	NotYetPaidError          = errors.New("NotYetPaidError")
	ListPriceError           = errors.New("ListPriceError")
	SalePriceError           = errors.New("SalePriceError")
	TotalListPriceError      = errors.New("TotalListPriceError")
	TotalSalePriceError      = errors.New("TotalSalePriceError")
	TotalDiscountPriceError  = errors.New("TotalDiscountPriceError")
	TotalPaymentPriceError   = errors.New("TotalPaymentPriceError")
	TotalOfferPriceError     = errors.New("TotalOfferPriceError")
	ProductError             = errors.New("ProductError")
	SkuError                 = errors.New("SkuError")
	RefundPriceError         = errors.New("RefundPriceError")
	RefundRequestError       = errors.New("RefundRequestError")
	RemainAmountInvalidError = fmt.Errorf("RemainAmountInvalidError")
	AddressInvalidError      = fmt.Errorf("AddressInvalidError")
	NotAuthorizedError       = fmt.Errorf("Not authorized.")
	NotAuthorizedActionError = fmt.Errorf("Not authorized action.")

	validate = factory.NewValidator()
)
