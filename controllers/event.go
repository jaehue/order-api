package controllers

import (
	"net/http"
	"strconv"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type EventController struct{}

func (c EventController) Init(g *echo.Group) {
	g.POST("", c.HandleEvent)
	g.GET("/republish/order/:id", c.RePublishOrderMessage)
	g.GET("/republish/refund/:id", c.RePublishRefundMessage)
}

func (this EventController) HandleEvent(c echo.Context) error {
	var event models.Event

	if err := c.Bind(&event); err != nil {
		return renderFail(c, http.StatusBadRequest, err)
	}

	if err := (models.EventHandler{}).HandleEvent(c.Request().Context(), event); err != nil {
		return renderFail(c, http.StatusInternalServerError, err)
	}
	return renderSuccess(c, nil)
}

func (this EventController) RePublishOrderMessage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	logrus.WithField("id", id).Info("Param id")
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "order id"))
	}
	orderResult, err := models.Order{}.GetOrder(c.Request().Context(), "", 0, id, nil, "", true)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	if orderResult.Id == 0 {
		return renderFail(c, http.StatusBadRequest, api.ErrorNotFound.New(nil))
	}
	msg := orderResult.RePublishEventMessages(c.Request().Context())
	return renderSuccess(c, msg)
}
func (this EventController) RePublishRefundMessage(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	logrus.WithField("id", id).Info("Param id")
	if id == 0 {
		return renderFail(c, http.StatusBadRequest, factory.NewError(api.ErrorMissParameter, "refund id"))
	}
	refundResult, err := models.Refund{}.GetRefund(c.Request().Context(), "", 0, id, 0, nil, "", true)
	if err != nil {
		return renderFail(c, http.StatusBadRequest, api.ErrorParameterParsingFailed.New(err))
	}
	if refundResult.Id == 0 {
		return renderFail(c, http.StatusBadRequest, api.ErrorNotFound.New(nil))
	}
	msg := refundResult.RePublishEventMessages(c.Request().Context())
	return renderSuccess(c, msg)
}
