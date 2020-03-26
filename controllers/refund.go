package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/hublabs/common/auth"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type RefundController struct{}

func (c RefundController) Init(g *echo.Group) {
	g.GET("", c.GetRefunds)
	g.GET("/:id", c.GetRefund)
	g.POST("", c.Create)
	g.GET("/expectedRefund", c.ExpectedRefund)
	g.POST("/expectedRefund", c.ExpectedRefund)
	g.PUT("/confirm", c.RefundOrderConfirm)
	g.PUT("/cancel", c.RefundOrderCancel)
	g.PUT("/approvals", c.RefundOrderApprovals)
}

func (RefundController) Create(c echo.Context) error {
	var refundInput RefundInput
	if err := c.Bind(&refundInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(refundInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	id := refundInput.OrderId
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}
	totalCount, orderResults, err := models.Order{}.GetOrdersByItem(c.Request().Context(), "", refundInput.CustomerId, 0, "", strconv.FormatInt(id, 10), "", "", "", 0, 0, true)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if totalCount == 0 {
		return renderFail(c, http.StatusBadRequest, api.ErrorNotFound.New(nil))
	}

	var oldOrderItemSeparates, orderItemSeparates []models.OrderItemSeparate

	defer func() {
		if err != nil {
			OrderItemSeparatesDelete(c.Request().Context(), oldOrderItemSeparates, orderItemSeparates)
		}
	}()

	oldOrderItemSeparates, orderItemSeparates, err = refundInput.PartialRefund(c.Request().Context(), orderResults[0])
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	userClaim := UserClaim(auth.UserClaim{}.FromCtx(c.Request().Context()))
	refund, err := refundInput.NewRefundEntity(userClaim, orderResults[0])
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	_, err = refund.RefundValidate(c.Request().Context(), "", orderResults[0])
	if err != nil {
		code, _, detail := getErrorInfo(err)
		if code == "22001" {
			return renderFail(c, http.StatusBadRequest, ApiErrorGiftNotRefund(detail))
		}
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := refund.Save(c.Request().Context()); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, &refund)
}

func (RefundController) ExpectedRefund(c echo.Context) error {
	var refundInput RefundInput
	if err := c.Bind(&refundInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(refundInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	id := refundInput.OrderId
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}
	// itemIds := []int64{}
	// for _, item := range refundInput.RefundOrderItems {
	// 	itemIds = append(itemIds, item.OrderItemId)
	// }
	totalCount, orderResults, err := models.Order{}.GetOrdersByItem(c.Request().Context(), "", refundInput.CustomerId, 0, "", strconv.FormatInt(id, 10), "", "", "", 0, 0, true)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if totalCount == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}

	var oldOrderItemSeparates, orderItemSeparates []models.OrderItemSeparate

	defer func() {
		OrderItemSeparatesDelete(c.Request().Context(), oldOrderItemSeparates, orderItemSeparates)
	}()

	oldOrderItemSeparates, orderItemSeparates, err = refundInput.PartialRefund(c.Request().Context(), orderResults[0])
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	userClaim := UserClaim(auth.UserClaim{}.FromCtx(c.Request().Context()))
	refund, err := refundInput.NewRefundEntity(userClaim, orderResults[0])
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	_, err = refund.RefundValidate(c.Request().Context(), "", orderResults[0])
	if err != nil {
		code, _, detail := getErrorInfo(err)
		if code == "22001" || code == "22002" {
			return c.JSON(http.StatusOK, api.Result{
				Result:  refund,
				Success: true,
				Error:   ApiErrorGiftNotRefund(detail),
			})
		}
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, &refund)
}

func (RefundController) GetRefunds(c echo.Context) error {
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	orderId, _ := strconv.ParseInt(c.QueryParam("orderId"), 10, 64)
	tenantCode := c.QueryParam("tenantCode")
	refundId, _ := strconv.ParseInt(c.QueryParam("refundId"), 10, 64)
	storeId, _ := strconv.ParseInt(c.QueryParam("storeId"), 10, 64)
	refundStatus := c.QueryParam("refundStatus")
	outerOrderNo := c.QueryParam("outerOrderNo")
	skipCount, _ := strconv.ParseInt(c.QueryParam("skipCount"), 10, 64)
	maxResultCount, _ := strconv.ParseInt(c.QueryParam("maxResultCount"), 10, 64)
	isCancelInclude, _ := strconv.ParseBool(c.QueryParam("isCancelInclude"))
	if maxResultCount == 0 || maxResultCount > 100 {
		maxResultCount = defaultMaxResultCount
	}
	logrus.WithField("refundStatus", refundStatus).Info("refundStatus")
	startAt := c.QueryParam("startAt")
	endAt := c.QueryParam("endAt")
	logrus.WithField("startAt", startAt).Info("startAt")

	if valid, err := DateTermMaxValidate(startAt, endAt, 31); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	} else if !valid {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorInvalidFields, fmt.Sprintf("Invalid fields [ %s, %s ]. Date term should ress then 1 month. [%s~%s] is invalid term.", "startAt", "endAt", startAt, endAt)))
	}

	totalCount, refundResult, err := models.Refund{}.GetRefunds(c.Request().Context(), tenantCode, customerId, refundId, orderId, storeId, outerOrderNo, refundStatus, startAt, endAt, int(skipCount), int(maxResultCount), isCancelInclude)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, api.ArrayResult{
		TotalCount: totalCount,
		Items:      refundResult,
	})
}

func (RefundController) GetRefund(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	customerId, _ := strconv.ParseInt(c.QueryParam("customerId"), 10, 64)
	tenantCode := c.QueryParam("tenantCode")

	logrus.WithField("id", id).Info("Param id")
	logrus.WithField("customerId", customerId).Info("Param customerId")

	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "refund id"))
	}

	refundResult, err := models.Refund{}.GetRefund(c.Request().Context(), tenantCode, customerId, id, 0, nil, "", true)

	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	return renderSuccess(c, refundResult)
}

func (RefundController) RefundOrderConfirm(c echo.Context) error {
	var refundStatusInput RefundStatusInput
	if err := c.Bind(&refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	id := refundStatusInput.RefundId
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "refund id"))
	}

	event, err := refundStatusInput.RefundOrderStatusEntity(c.Request().Context(), enum.RefundOrderProcessing)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if err := (models.EventHandler{}).HandleEvent(c.Request().Context(), event); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}

	return renderSuccess(c, nil)
}

func (RefundController) RefundOrderCancel(c echo.Context) error {
	var refundStatusInput RefundStatusInput
	if err := c.Bind(&refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	id := refundStatusInput.RefundId
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "refund id"))
	}

	event, err := refundStatusInput.RefundOrderStatusEntity(c.Request().Context(), enum.RefundOrderCancel)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if err := (models.EventHandler{}).HandleEvent(c.Request().Context(), event); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}

	return renderSuccess(c, nil)
}

func (RefundController) RefundOrderApprovals(c echo.Context) error {
	var refundStatusInput RefundStatusInput
	if err := c.Bind(&refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}

	if err := c.Validate(refundStatusInput); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	id := refundStatusInput.RefundId
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "refund id"))
	}

	event, err := refundStatusInput.RefundOrderStatusEntity(c.Request().Context(), enum.RefundRequisiteApprovals)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if err := (models.EventHandler{}).HandleEvent(c.Request().Context(), event); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}

	return renderSuccess(c, nil)
}
