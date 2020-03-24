package models

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hublabs/order-api/config"

	"github.com/hublabs/order-api/adapters"
)

type ProductHandler struct {
}

type ItemInfo struct {
	Code    string  `json:"code" `
	FeeRate float64 `json:"feeRate"`
}

// func (c ProductHandler) UpdateFeeRate(ctx context.Context, order *Order) error {
// 	for i, item := range order.Items {
// 		itemInfo, err := getItemByCode(ctx, item.ItemCode)
// 		if err != nil {
// 			return err
// 		}
// 		if itemInfo.FeeRate == 0 {
// 			return fmt.Errorf("[%s]FeeRate 0", item.ItemCode)
// 		}
// 		order.Items[i].FeeRate = itemInfo.FeeRate
// 	}
// 	return nil
// }

func (c ProductHandler) GetItemByCode(ctx context.Context, itemCode string) (ItemInfo, error) {
	var resp struct {
		Result  ItemInfo `json:"result"`
		Success bool     `json:"success"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Details string `json:"details"`
		} `json:"error"`
	}

	url := config.Config().Services.ProductApiUrl + "/v1/items/" + itemCode
	err := adapters.RetryRestApi(ctx, &resp, http.MethodGet, url, nil)
	if err != nil {
		return ItemInfo{}, err
	}
	if !resp.Success {
		return ItemInfo{}, fmt.Errorf("[%d]%s", resp.Error.Code, resp.Error.Details)
	}
	return resp.Result, nil
}
