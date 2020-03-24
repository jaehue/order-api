package models

import (
	"context"
	"nomni/utils/auth"

	"github.com/hublabs/common/api"
	"github.com/hublabs/order-api/enum"
	"github.com/hublabs/order-api/factory"
)

func (refund *Refund) ChangeStatus(ctx context.Context, status string) (*Refund, error) {
	isChange := false
	if status == enum.RefundOrderProcessing.String() {
		if refund.Status != enum.RefundOrderRegistered.String() {
			return nil, factory.NewError(api.ErrorInvalidStatus, "refund is confirmed :"+refund.Status)
		}
	}
	if refund.Status != status {
		refund.Status = status
		isChange = true
		if err := refund.changeStatus(ctx); err != nil {
			return nil, err
		}
	}
	for i := range refund.Items {
		if refund.Items[i].Status != status {
			refund.Items[i].Status = status
			isChange = true
			if err := refund.Items[i].changeStatus(ctx); err != nil {
				return nil, err
			}
		}
	}
	if isChange == true {
		refund.addEventToOutBox(status)
		userClaim := auth.UserClaim{}.FromCtx(ctx)
		refundHistory := refund.NewRefundHistory(userClaim)
		if err := refundHistory.Save(ctx); err != nil {
			return nil, err
		}
	}
	refundNotDeliveryItems := []int64{}
	for _, item := range refund.Items {
		if item.IsDelivery == false {
			refundNotDeliveryItems = append(refundNotDeliveryItems, item.Id)
		}
	}
	if status == enum.RefundOrderProcessing.String() && len(refundNotDeliveryItems) > 0 {
		if _, err := refund.ChangeStatus(ctx, enum.RefundRequisiteApprovals.String()); err != nil {
			return nil, err
		}
	}
	return refund, nil
}

func (refund *Refund) changeStatus(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(refund.Id).
		Cols("status", "refuse_reason").
		Update(refund); err != nil {
		return err
	}

	return nil
}

func (refundItem *RefundItem) changeStatus(ctx context.Context) error {
	if _, err := factory.DB(ctx).
		ID(refundItem.Id).
		And("refund_id = ?", refundItem.RefundId).
		Cols("status").
		Update(refundItem); err != nil {
		return err
	}
	return nil
}
