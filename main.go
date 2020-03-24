package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hublabs/common/auth"
	"github.com/hublabs/order-api/adapters"
	"github.com/hublabs/order-api/factory"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/hublabs/order-api/config"
	"github.com/hublabs/order-api/controllers"
	"github.com/hublabs/order-api/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pangpanglabs/goutils/echomiddleware"
)

func main() {
	c := config.Init(os.Getenv("APP_ENV"), "")
	db := initDB(c.Database.Driver, c.Database.Connection)

	orderEventMessagePublisher, err := adapters.NewOrderEventMessagePublisher(c.EventMessageBroker.Kafka)
	if err != nil {
		log.Fatal("error set up OrderEventMessagePublisher", err)
	} else {
		adapters.EventMessagePublisher = orderEventMessagePublisher
	}
	defer orderEventMessagePublisher.Close()

	e := echo.New()
	e.Validator = factory.NewValidator()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	e.GET("/swagger", func(c echo.Context) error {
		return c.File("./swagger.yml")
	})
	e.Static("/docs", "./swagger-ui")

	v1 := e.Group("/v1")
	controllers.OrderController{}.Init(v1.Group("/order"))
	controllers.RefundController{}.Init(v1.Group("/refund"))
	controllers.EventController{}.Init(v1.Group("/events"))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(echomiddleware.BehaviorLogger(c.ServiceName, c.BehaviorLog.Kafka))
	e.Use(echomiddleware.ContextDB(c.ServiceName, db, c.Database.Logger.Kafka))
	e.Use(auth.UserClaimMiddleware("/ping", "/swagger", "/docs"))

	if err := e.Start(":8000"); err != nil {
		log.Println(err)
	}
}

func initDB(driver, connection string) *xorm.Engine {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		panic(err)
	}
	db.ShowSQL(true)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(30)
	db.SetConnMaxLifetime(time.Minute * 10)

	if err := models.Init(db); err != nil {
		log.Fatal(err)
	}
	return db
}
