package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hublabs/common/auth"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type OrderController struct{}

func (c OrderController) Init(g *echo.Group) {
	g.GET("", c.GetOrders)
	g.GET("/:id", c.GetOrder)
	g.GET("/statusCount", c.GetOrderStatusCount)
	g.GET("/:id/orderValidate", c.GetOrderValidate)
	g.GET("/orders", c.GetOrderByItems)
	g.GET("/orderItems", c.GetOrderItems)
	g.POST("", c.Create)
	g.DELETE("/:id/delete", c.Delete)
	g.PUT("/:id/cancel", c.SaleOrderCancel)
	g.PUT("/update", c.OrderUpdate)
}

//Create Create
func (OrderController) Create(c echo.Context) error {
	var orderInput OrderInput
	if err := c.Bind(&orderInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(orderInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}
	userClaim := UserClaim(auth.UserClaim{}.FromCtx(c.Request().Context()))
	order, err := orderInput.NewOrderEntity(userClaim)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	if err := order.Validate(c.Request().Context()); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := order.Save(c.Request().Context()); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}

	return renderSuccess(c, &order)
}

//Get Orders by id & orderStatus
func (OrderController) GetOrders(c echo.Context) error {
	ids := c.QueryParam("ids")
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	storeId, _ := strconv.ParseInt(c.QueryParam("storeId"), 10, 64)
	salesmanId, _ := strconv.ParseInt(c.QueryParam("salesmanId"), 10, 64)
	orderStatus := c.QueryParam("orderStatus")
	saleType := c.QueryParam("saleType")
	startAt := c.QueryParam("startAt")
	endAt := c.QueryParam("endAt")
	tenantCode := c.QueryParam("tenantCode")
	skipCount, _ := strconv.ParseInt(c.QueryParam("skipCount"), 10, 64)
	maxResultCount, _ := strconv.ParseInt(c.QueryParam("maxResultCount"), 10, 64)
	isGetAll, _ := strconv.ParseBool(c.QueryParam("isGetAll"))
	outerOrderNo := c.QueryParam("outerOrderNo")
	if !isGetAll && (maxResultCount == 0 || maxResultCount > 200) {
		maxResultCount = defaultMaxResultCount
	}
	if valid, err := DateTermMaxValidate(startAt, endAt, 31); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	} else if !valid {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorInvalidFields, fmt.Sprintf("Invalid fields [ %s, %s ]. Date term should ress then 1 month. [%s~%s] is invalid term.", "startAt", "endAt", startAt, endAt)))
	}

	totalCount, orderResult, err := models.Order{}.GetOrders(c.Request().Context(), tenantCode, customerId, storeId, salesmanId, orderStatus, saleType, startAt, endAt, ids, outerOrderNo, int(skipCount), int(maxResultCount))

	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, api.ArrayResult{
		TotalCount: totalCount,
		Items:      orderResult,
	})
}

func (OrderController) GetOrder(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	tenantCode := c.QueryParam("tenantCode")

	logrus.WithField("id", id).Info("Param id")
	logrus.WithField("customerId", customerId).Info("Param customerId")

	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}

	orderResult, err := models.Order{}.GetOrder(c.Request().Context(), tenantCode, customerId, id, nil, "", true)

	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, orderResult)
}

func (OrderController) GetOrderByItems(c echo.Context) error {
	ids := c.QueryParam("orderIds")
	tenantCode := c.QueryParam("tenantCode")
	orderItemIds := c.QueryParam("orderItemIds")
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	storeId, _ := strconv.ParseInt(c.QueryParam("storeId"), 10, 64)
	orderStatus := c.QueryParam("orderStatus")
	startAt := c.QueryParam("startAt")
	endAt := c.QueryParam("endAt")
	isRemoveRefund, _ := strconv.ParseBool(c.QueryParam("isRemoveRefund"))
	skipCount, _ := strconv.ParseInt(c.QueryParam("skipCount"), 10, 64)
	maxResultCount, _ := strconv.ParseInt(c.QueryParam("maxResultCount"), 10, 64)
	isGetAll, _ := strconv.ParseBool(c.QueryParam("isGetAll"))
	if !isGetAll && (maxResultCount == 0 || maxResultCount > 200) {
		maxResultCount = defaultMaxResultCount
	}

	if valid, err := DateTermMaxValidate(startAt, endAt, 31); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	} else if !valid {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorInvalidFields, fmt.Sprintf("Invalid fields [ %s, %s ]. Date term should ress then 1 month. [%s~%s] is invalid term.", "startAt", "endAt", startAt, endAt)))
	}

	totalCount, orderResults, err := models.Order{}.GetOrdersByItem(c.Request().Context(), tenantCode, customerId, storeId, orderStatus, ids, orderItemIds, startAt, endAt, int(skipCount), int(maxResultCount), isRemoveRefund)

	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, api.ArrayResult{
		TotalCount: totalCount,
		Items:      orderResults,
	})
}

func (OrderController) GetOrderItems(c echo.Context) error {
	orderIds := c.QueryParam("orderIds")
	tenantCode := c.QueryParam("tenantCode")
	orderItemIds := c.QueryParam("orderItemIds")
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	storeId, _ := strconv.ParseInt(c.QueryParam("storeId"), 10, 64)
	orderStatus := c.QueryParam("orderStatus")
	startAt := c.QueryParam("startAt")
	endAt := c.QueryParam("endAt")
	isRemoveRefund, _ := strconv.ParseBool(c.QueryParam("isRemoveRefund"))
	skipCount, _ := strconv.ParseInt(c.QueryParam("skipCount"), 10, 64)
	maxResultCount, _ := strconv.ParseInt(c.QueryParam("maxResultCount"), 10, 64)
	isGetAll, _ := strconv.ParseBool(c.QueryParam("isGetAll"))
	if !isGetAll && (maxResultCount == 0 || maxResultCount > 200) {
		maxResultCount = defaultMaxResultCount
	}
	if valid, err := DateTermMaxValidate(startAt, endAt, 31); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	} else if !valid {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorInvalidFields, fmt.Sprintf("Invalid fields [ %s, %s ]. Date term should ress then 1 month. [%s~%s] is invalid term.", "startAt", "endAt", startAt, endAt)))
	}

	totalCount, orderItemResults, err := models.OrderItem{}.GetOrderItemsWithRefundItem(c.Request().Context(), tenantCode, customerId, storeId, orderStatus, orderIds, orderItemIds, startAt, endAt, int(skipCount), int(maxResultCount), isRemoveRefund)

	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, api.ArrayResult{
		TotalCount: totalCount,
		Items:      orderItemResults,
	})
}

func (OrderController) GetOrderValidate(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantCode := c.QueryParam("tenantCode")
	logrus.WithField("id", id).Info("Param id")
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}
	orderResult, err := models.Order{}.GetOrder(c.Request().Context(), tenantCode, 0, id, nil, "", false)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	if err := orderResult.Validate(c.Request().Context()); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	return renderSuccess(c, nil)
}

func (OrderController) GetOrderStatusCount(c echo.Context) error {
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	isOrder, _ := strconv.ParseBool(c.QueryParam("isOrder"))
	tenantCode := c.QueryParam("tenantCode")
	result, err := models.Order{}.GetOrderStatusCount(c.Request().Context(), tenantCode, customerId, isOrder)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorUnknown.New(err))
	}
	return renderSuccess(c, result)
}

func (OrderController) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	logrus.WithField("id", id).Info("Param id")
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}

	var orderDelete OrderDelete
	if err := c.Bind(&orderDelete); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(orderDelete); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	isChange, order, err := NewOrderEntity(c.Request().Context(), orderDelete.CustomerId, id, orderDelete)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if isChange == true {
		if err := order.Update(c.Request().Context()); err != nil {
			return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
		}
	}

	return renderSuccess(c, &order)
}

func (OrderController) SaleOrderCancel(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	logrus.WithField("id", id).Info("Param id")

	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}

	event, err := SaleOrderStatusEntity(c.Request().Context(), enum.SaleOrderCancel, id)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if err := (models.EventHandler{}).HandleEvent(c.Request().Context(), *event); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}

	return renderSuccess(c, nil)
}
func (OrderController) OrderUpdate(c echo.Context) error {
	ids := c.QueryParam("ids")
	if ids == "" {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "ids"))
	}
	updateType := c.QueryParam("updateType")
	if updateType != "feerate" {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "updateType"))
	}
	isResendEvent, _ := strconv.ParseBool(c.QueryParam("isResendEvent"))
	totalCount, orderResult, err := models.Order{}.GetOrders(c.Request().Context(), "", 0, 0, 0, "", "", "", "", ids, "", 0, 0)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	switch strings.ToLower(updateType) {
	case "feerate":
		for _, order := range orderResult {
			_, refundResult, _ := models.Refund{}.GetRefunds(c.Request().Context(), "", 0, 0, order.Id, 0, "", "", "", "", 0, 0, true)
			for i, item := range order.Items {
				itemInfo, err := models.ProductHandler{}.GetItemByCode(c.Request().Context(), item.ItemCode)
				if err != nil {
					return renderFail(c, http.StatusInternalServerError, ApiErrorUpdate("GetItemByCode Err %s"+err.Error()))
				}
				if itemInfo.FeeRate == 0 {
					return renderFail(c, http.StatusInternalServerError, fmt.Errorf("orderId:%d itemcode:%s feeRate 0", order.Id, item.ItemCode))
				}
				if item.FeeRate != itemInfo.FeeRate && item.FeeRate == 0 {
					order.Items[i].FeeRate = itemInfo.FeeRate
					if err := order.Items[i].UpdateFeeRate(c.Request().Context()); err != nil {
						return renderFail(c, http.StatusInternalServerError, err)
					}

				}
				for i, refund := range refundResult {
					for j, refundItem := range refund.Items {
						if refundItem.ItemCode == item.ItemCode && refundItem.FeeRate != itemInfo.FeeRate && refundItem.FeeRate == 0 {
							refundResult[i].Items[j].FeeRate = itemInfo.FeeRate
							if err := refundResult[i].Items[j].UpdateFeeRate(c.Request().Context()); err != nil {
								return renderFail(c, http.StatusInternalServerError, err)
							}
						}
					}

				}
			}
			if isResendEvent {
				order.RePublishEventMessages(c.Request().Context())
				for _, refund := range refundResult {
					refund.RePublishEventMessages(c.Request().Context())
				}
			}
		}
	}
	return renderSuccess(c, api.ArrayResult{
		TotalCount: totalCount,
		Items:      orderResult,
	})
}
