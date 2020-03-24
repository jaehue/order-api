package factory

import (
	"context"

	"github.com/go-xorm/xorm"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/number"
)

const (
	ContextExportLogger = "ExportLogger"
	ContextTenant       = "Tenant"
	ContextChannel      = "Channel"
)

func DB(ctx context.Context) xorm.Interface {
	v := ctx.Value(echomiddleware.ContextDBName)
	if v == nil {
		panic("DB is not exist")
	}
	if db, ok := v.(*xorm.Session); ok {
		return db
	}
	if db, ok := v.(*xorm.Engine); ok {
		return db
	}
	panic("DB is not exist")
}

func CommitDB(ctx context.Context) error {
	v := ctx.Value(echomiddleware.ContextDBName)
	if v == nil {
		panic("DB is not exist")
	}
	session, ok := v.(*xorm.Session)
	if ok {
		return session.Commit()
	}
	return nil
}

func RollBackDB(ctx context.Context) error {
	v := ctx.Value(echomiddleware.ContextDBName)
	if v == nil {
		panic("DB is not exist")
	}
	session, ok := v.(*xorm.Session)
	if ok {
		return session.Rollback()
	}
	return nil
}

func PriceSetting(ctx context.Context) *number.Setting {
	return &number.Setting{}
}
