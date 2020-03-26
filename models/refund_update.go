package models

import (
	"context"

	"github.com/hublabs/common/auth"

	"github.com/hublabs/order-api/factory"
)

func (o *Refund) Update(ctx context.Context) error {
	err := o.update(ctx)
	if err != nil {
		return err
	}

	for i := range o.Items {
		if err := o.Items[i].update(ctx); err != nil {
			return err
		}
	}

	userClaim := UserClaim(auth.UserClaim{}.FromCtx(ctx))
	refundHistory := o.NewRefundHistory(userClaim)
	if err := refundHistory.Save(ctx); err != nil {
		return err
	}

	return nil
}

func (refund *Refund) update(ctx context.Context) (err error) {
	session := factory.DB(ctx).ID(refund.Id)
	session.Cols("mileage")

	_, err = session.Update(refund)
	if err != nil {
		return err
	}

	return nil
}

func (refundItem *RefundItem) update(ctx context.Context) (err error) {
	session := factory.DB(ctx).ID(refundItem.Id)
	session.Cols("mileage")

	_, err = session.Update(refundItem)
	if err != nil {
		return err
	}

	return nil
}
func (refundItem *RefundItem) UpdateFeeRate(ctx context.Context) (err error) {
	session := factory.DB(ctx).ID(refundItem.Id)
	session.Cols("fee_rate")
	_, err = session.Update(refundItem)
	if err != nil {
		return err
	}
	return nil
}
