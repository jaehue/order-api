package models

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hublabs/order-api/adapters"

	"github.com/sirupsen/logrus"
)

type StockEventHandler struct {
	StockApiUrl string
}

func (c StockEventHandler) StockValidate(ctx context.Context, order *Order) error {
	for _, item := range order.Items {
		if item.IsStockChecked == true {
			continue
		}
		stock, err := c.GetStock(ctx, item.ProductId, item.SkuId)
		if err != nil {
			return err
		}

		if item.Quantity > stock.AvailableQuantity {
			return fmt.Errorf("Sale Quantity better then Sale Available Quantity")
		}
	}

	return nil
}

func (c StockEventHandler) GetStock(ctx context.Context, productId int64, skuId int64) (Stock, error) {
	var resp struct {
		Result  Stock `json:"result"`
		Success bool  `json:"success"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Details string `json:"details"`
		} `json:"error"`
	}
	ProductId := strconv.FormatInt(productId, 10)
	SkuId := strconv.FormatInt(skuId, 10)
	url := c.StockApiUrl + "/v1/stocks/" + ProductId + "/" + SkuId

	err := adapters.RetryRestApi(ctx, &resp, http.MethodGet, url, nil)
	if err != nil {
		return Stock{}, err
	}
	if !resp.Success {
		logrus.WithFields(logrus.Fields{
			"productId":    productId,
			"skuId":        skuId,
			"errorCode":    resp.Error.Code,
			"errorMessage": resp.Error.Message,
		}).Error("Fail to get stock")
		return Stock{}, fmt.Errorf("[%d]%s", resp.Error.Code, resp.Error.Details)
	}

	return resp.Result, nil
}

type Stock struct {
	Id                int64
	SkuId             int64 `json:"skuId" xorm:"index"`
	ProductId         int64 `json:"productId" xorm:"index"`
	AvailableQuantity int   `json:"availableQuantity"`
}
