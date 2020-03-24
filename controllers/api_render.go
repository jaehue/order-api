package controllers

import (
	"net/http"
	"strings"

	"github.com/hublabs/common/api"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/behaviorlog"
)

var (
	// Business Error
	ApiErrorCanNotBeCanceled = api.Error{Code: 20001, Message: "This order can not be canceled."}
	ApiErrorAlreadyDeleted   = api.Error{Code: 20002, Message: "This order has already been deleted."}
	ApiErrorGiftNotRefund    = func(detail string) api.Error {
		return api.Error{Code: 20003, Message: "This gift should be refund.", Details: detail}
	}
	ApiErrorUpdate = func(detail string) api.Error {
		return api.Error{Code: 20004, Message: "Order Update Error.", Details: detail}
	}
)

func renderFail(c echo.Context, status int, err error) error {
	if err != nil {
		behaviorlog.FromCtx(c.Request().Context()).WithError(err)
	}
	if apiError, ok := err.(api.Error); ok {
		return c.JSON(status, api.Result{
			Error: apiError,
		})
	}

	return c.JSON(status, api.Result{
		Success: false,
		Error:   api.ErrorUnknown.New(err),
	})
}

func renderSuccess(c echo.Context, result interface{}) error {
	return c.JSON(http.StatusOK, api.Result{
		Success: true,
		Result:  result,
	})
}

func getErrorInfo(err error) (code, msg, detail string) {
	if err == nil {
		return
	}
	strArr := strings.Split(err.Error(), "#")
	for i, content := range strArr {
		switch i {
		case 0:
			code = content
		case 1:
			msg = content
		case 2:
			detail = content
		}
	}
	return
}
