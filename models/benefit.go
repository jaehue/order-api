package models

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hublabs/common/api"

	"github.com/hublabs/order-api/adapters"
)

type BenefitHandler struct {
	BenefitApiUrl string
}
type Mileage struct {
	TenantCode  string        `json:"tenantCode,omitempty"`     /*租户代码*/
	StoreId     int64         `json:"storeId,omitempty"`        /*店铺Id*/
	MemberId    int64         `json:"memberId" validate:"gt=0"` /*会员id*/
	ChannelId   int64         `json:"channelId"`                /*渠道*/
	TradeNo     string        `json:"tradeNo,omitempty"`        /*交易单号*/
	ObtainPoint float64       `json:"obtainPoint,omitempty"`    /*累积积分数量*/
	Point       float64       `json:"point,omitempty"`          /*使用积分数量*/
	PointPrice  float64       `json:"pointPrice,omitempty"`     /*积分抵扣金额*/
	Items       []MileageItem `json:"items,omitempty" `
}

type MileageItem struct {
	ItemId          int64   `json:"itemId,omitempty"`          /*详情Id*/
	TotalListAmount float64 `json:"totalListAmount,omitempty"` /*吊牌价*/
	ObtainPoint     float64 `json:"obtainPoint,omitempty"`     /*累积积分数量*/
	Point           float64 `json:"point,omitempty"`           /*积分数量*/
	PointPrice      float64 `json:"pointPrice,omitempty"`      /*积分抵扣金额*/
}

func (c BenefitHandler) GetRefundMileage(ctx context.Context, param Mileage) (Mileage, error) {
	var resp struct {
		Result  Mileage   `json:"result"`
		Success bool      `json:"success"`
		Error   api.Error `json:"error"`
	}
	url := c.BenefitApiUrl + "/v1/mileage/refund"

	err := adapters.RetryRestApi(ctx, &resp, http.MethodPost, url, param)
	if err != nil {
		return Mileage{}, err
	}
	if !resp.Success {
		LogParamErrorMsg("GetRefundMileage", param, resp.Error)
		return Mileage{}, fmt.Errorf("[%d]%s", resp.Error.Code, resp.Error.Details)
	}

	return resp.Result, nil
}
