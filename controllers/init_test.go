package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"nomni/utils/auth"

	"github.com/hublabs/order-api/adapters"
	"github.com/hublabs/order-api/config"
	"github.com/hublabs/order-api/factory"
	"github.com/hublabs/order-api/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/behaviorlog"
	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/jwtutil"
)

var (
	echoApp          *echo.Echo
	handleWithFilter func(handlerFunc echo.HandlerFunc, c echo.Context) error
	ctx              context.Context
	xormEngine       *xorm.Engine
)

func TestMain(m *testing.M) {
	db := enterTest()
	xormEngine = db
	code := m.Run()
	exitTest(db)
	os.Exit(code)
}

func enterTest() *xorm.Engine {
	configutil.SetConfigPath("../")
	c := config.Init(os.Getenv("APP_ENV"), "")

	xormEngine, err := xorm.NewEngine(c.Database.Driver, c.Database.Connection)
	if err != nil {
		panic(err)
	}
	xormEngine.SetMaxIdleConns(10)
	xormEngine.SetMaxOpenConns(30)
	xormEngine.SetConnMaxLifetime(time.Minute * 10)
	xormEngine.ShowSQL(true)

	if err = models.DropTables(xormEngine); err != nil {
		panic(err)
	}

	if err = models.Init(xormEngine); err != nil {
		panic(err)
	}
	echoApp = echo.New()
	echoApp.Validator = factory.NewValidator()
	db := echomiddleware.ContextDB("test", xormEngine, echomiddleware.KafkaConfig{})
	kafkaConfig := c.EventMessageBroker.Kafka
	orderEventMessagePublisher, _ := adapters.NewOrderEventMessagePublisher(kafkaConfig)
	adapters.EventMessagePublisher = orderEventMessagePublisher
	behaviorlogger := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			c.SetRequest(req.WithContext(context.WithValue(req.Context(),
				behaviorlog.LogContextName, behaviorlog.New("test", req),
			)))
			return next(c)
		}
	}
	header := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c.SetRequest(req)
			return next(c)
		}
	}
	handleWithFilter = func(handlerFunc echo.HandlerFunc, c echo.Context) error {
		return behaviorlogger(header(db(handlerFunc)))(c)
	}
	ctx = context.WithValue(context.Background(), echomiddleware.ContextDBName, xormEngine)
	return xormEngine
}

func exitTest(db *xorm.Engine) {
	// if err := models.DropTables(db); err != nil {
	// 	panic(err)
	// }
}
func getToken() string {
	JWT_SECRET := "Z2tza3NsYWRtbDAxMjM0ZWhkbnRsYTk4"
	token, _ := jwtutil.NewTokenWithSecret(map[string]interface{}{
		"aud": "colleague", "tenantCode": "pangpang", "iss": "colleague", "id": 1001,
		"nbf": time.Now().Add(-5 * time.Minute).Unix(),
	}, JWT_SECRET)
	// os.Getenv("JWT_SECRET")
	// token, _ := jwtutil.NewToken(map[string]interface{}{"aud": "colleague", "tenantCode": "pangpang", "iss": "colleague"})
	return token
}

func setReq(url string, body interface{}) *http.Request {
	req := httptest.NewRequest(echo.PUT, url, nil)
	if body != nil {
		pb, _ := json.Marshal(body)
		req = httptest.NewRequest(echo.POST, url, bytes.NewReader(pb))
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderXRequestID, headerXRequestID)
	req.Header.Set(echo.HeaderAuthorization, getToken())
	userClaim := auth.UserClaim{
		Id:         1001,
		Iss:        "colleague",
		TenantCode: "pangpang",
	}
	req = req.WithContext(context.WithValue(context.Background(), "userClaim", userClaim))
	return req
}
