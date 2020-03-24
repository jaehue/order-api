package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hublabs/common/api"

	"github.com/hublabs/order-api/adapters"
	"github.com/hublabs/order-api/controllers"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/models"

	"github.com/pangpanglabs/goutils/test"
)

const (
	headerXRequestID = "Nylm1HI7eDAhvhrT6NKVNbQIVxwSEMPH"
)

func TestOrderCRUD(t *testing.T) {
	defer adapters.EventMessagePublisher.Close()
	inputs := []controllers.OrderInput{
		controllers.OrderInput{
			IsOutPaid:    true,
			OuterOrderNo: "1234-4321",
			CustomerId:   0,
			StoreId:      2366,
			SaleType:     "POS",
			Items: []controllers.ItemInput{
				controllers.ItemInput{
					ItemCode:       "1-2-3",
					ItemName:       "HB006103, 941, 165",
					SkuId:          1,
					ProductId:      2141,
					Quantity:       1,
					ListPrice:      123.50,
					SalePrice:      121,
					IsStockChecked: true,
					IsDelivery:     false,
				},
				controllers.ItemInput{
					ItemCode:       "1-2-2",
					ItemName:       "HB006103, 941, 165",
					SkuId:          2,
					ProductId:      2141,
					Quantity:       1,
					ListPrice:      123.50,
					SalePrice:      121,
					IsStockChecked: true,
					IsDelivery:     false,
				},
			},
		},
		controllers.OrderInput{
			IsOutPaid:    true,
			OuterOrderNo: "1234-4321",
			CustomerId:   0,
			StoreId:      2366,
			SaleType:     "WXSHOP",
			Items: []controllers.ItemInput{
				controllers.ItemInput{
					ItemCode:       "1-2-3",
					ItemName:       "HB006103, 941, 165",
					SkuId:          1,
					ProductId:      2141,
					Quantity:       1,
					ListPrice:      123.50,
					SalePrice:      121,
					IsStockChecked: true,
					IsDelivery:     false,
				},
				controllers.ItemInput{
					ItemCode:       "1-2-2",
					ItemName:       "HB006103, 941, 165",
					SkuId:          2,
					ProductId:      2141,
					Quantity:       1,
					ListPrice:      123.50,
					SalePrice:      121,
					IsStockChecked: true,
					IsDelivery:     false,
				},
			},
		},
	}
	orders := []models.Order{}
	for i, p := range inputs {
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			var v struct {
				Result  models.Order `json:"result"`
				Success bool         `json:"success"`
				Error   api.Error    `json:"error"`
			}
			rec := httptest.NewRecorder()
			req := setReq("/v1/order", p)
			test.Ok(t, handleWithFilter(controllers.OrderController{}.Create, echoApp.NewContext(req, rec)))
			test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))

			test.Equals(t, http.StatusOK, rec.Code)
			test.Equals(t, enum.SaleOrderProcessing.String(), v.Result.Status)
			orders = append(orders, v.Result)
		})
	}
	t.Run("SaleOrderCancel", func(t *testing.T) {
		var v struct {
			Result  models.Order `json:"result"`
			Success bool         `json:"success"`
			Error   api.Error    `json:"error"`
		}
		rec := httptest.NewRecorder()
		req := setReq("/v1/order/:id/cancel", nil)
		c := echoApp.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%v", orders[0].Id))
		test.Ok(t, handleWithFilter(controllers.OrderController{}.SaleOrderCancel, c))
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, http.StatusOK, rec.Code)
	})
	t.Run("GetOrder", func(t *testing.T) {
		var v struct {
			Result  models.Order `json:"result"`
			Success bool         `json:"success"`
			Error   api.Error    `json:"error"`
		}
		rec := httptest.NewRecorder()
		req := setReq("/v1/order", nil)
		c := echoApp.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%v", orders[0].Id))
		test.Ok(t, handleWithFilter(controllers.OrderController{}.GetOrder, c))
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, http.StatusOK, rec.Code)
		test.Equals(t, orders[0].OuterOrderNo, v.Result.OuterOrderNo)
		test.Equals(t, enum.SaleOrderCancel.String(), v.Result.Status)
		test.Equals(t, "pangpang", v.Result.TenantCode)
		test.Equals(t, "POS", v.Result.SaleType)
	})

	t.Run("GetOrderByItems", func(t *testing.T) {
		var v struct {
			Result struct {
				TotalCount int64          `json:"totalCount"`
				Items      []models.Order `json:"items"`
			} `json:"result"`
			Success bool      `json:"success"`
			Error   api.Error `json:"error"`
		}
		rec := httptest.NewRecorder()
		req := setReq("/v1/order/orders", nil)
		c := echoApp.NewContext(req, rec)
		ids := ""
		for _, order := range orders {
			ids += (strconv.FormatInt(order.Id, 10) + ",")
		}
		c.SetParamNames("orderIds")
		c.SetParamValues(ids)
		test.Ok(t, handleWithFilter(controllers.OrderController{}.GetOrderByItems, c))
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		test.Equals(t, http.StatusOK, rec.Code)
		test.Equals(t, int64(4), v.Result.TotalCount)
		test.Equals(t, enum.SaleOrderProcessing.String(), v.Result.Items[0].Status)
	})
	t.Run("OrderUpdate", func(t *testing.T) {
		var v struct {
			Result struct {
				TotalCount int64          `json:"totalCount"`
				Items      []models.Order `json:"items"`
			} `json:"result"`
			Success bool      `json:"success"`
			Error   api.Error `json:"error"`
		}
		rec := httptest.NewRecorder()
		ids := ""
		for _, order := range orders {
			ids += (strconv.FormatInt(order.Id, 10) + ",")
		}
		req := setReq("/v1/order/update?ids="+ids+"&updateType=feerate", nil)
		c := echoApp.NewContext(req, rec)
		test.Ok(t, handleWithFilter(controllers.OrderController{}.OrderUpdate, c))
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &v))
		// fmt.Println("v.Error.Details==", v.Error.Details)
		// fmt.Println("v.Error.Message==", v.Error.Message)
		// test.Equals(t, http.StatusOK, rec.Code)
		// test.Equals(t, int64(2), v.Result.TotalCount)
	})
}
