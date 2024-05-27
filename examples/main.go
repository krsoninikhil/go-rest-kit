package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/auth"
	"github.com/krsoninikhil/go-rest-kit/cache"
	"github.com/krsoninikhil/go-rest-kit/config"
	"github.com/krsoninikhil/go-rest-kit/crud"
	"github.com/krsoninikhil/go-rest-kit/integrations/twilio"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
	"github.com/krsoninikhil/go-rest-kit/request"
)

func main() {
	// configuration
	var (
		ctx  = context.Background()
		conf Config
	)
	config.Load(ctx, &conf)

	// connnections
	var (
		db    = pgdb.NewPGConnection(ctx, conf.DB)
		cache = cache.NewInMemory()
	)

	// inject dependencies for auth service
	var (
		userDao        = auth.NewUserDao[User](db)
		authSvc        = auth.NewService(conf.Auth, userDao)
		smsProvider    = twilio.NewClient(conf.Auth.Twilio)
		otpSvc         = auth.NewOTPSvc(conf.Auth.OTP, smsProvider, cache) //
		authController = auth.NewController(authSvc, otpSvc, cache)
	)

	// inject dependencies
	var (
		businessTypeDao  = crud.Dao[BusinessType]{PGDB: db} // use your own doa if you need custom implementation
		businessTypeCtlr = crud.Controller[BusinessType, BusinessTypeResponse, BusinessTypeRequest]{
			Svc: &businessTypeDao, // using dao for service as no business logic is required here
		} // prewritten controller struct with CRUD methods

		businessDao  = crud.Dao[Business]{PGDB: db}
		businessCtrl = crud.Controller[Business, BusinessResponse, BusinessRequest]{
			Svc: &businessDao,
		}

		// nested resources
		productDao        = crud.Dao[Product]{PGDB: db}
		productController = crud.NestedController[Product, ProductResponse, ProductRequest]{
			Svc: &productDao,
		}
	)

	r := gin.Default()

	r.POST("/auth/otp/send", request.BindCreate(authController.SendOTP))
	r.POST("/auth/otp/verify", request.BindCreate(authController.VerifyOTP))
	r.POST("/auth/token/refresh", request.BindCreate(authController.RefreshToken))
	r.GET("/countries/:alpha2Code", request.BindGet(authController.CountryInfo))

	r.GET("/business-types", request.BindGet(businessTypeCtlr.List))
	r.POST("/business-types", request.BindCreate(businessTypeCtlr.Create))
	r.GET("/business-types/:id", request.BindGet(businessTypeCtlr.Retrieve))
	r.PATCH("/business-types/:id", request.BindUpdate(businessTypeCtlr.Update))
	r.DELETE("/business-types/:id", request.BindDelete(businessTypeCtlr.Delete))

	// use auth middleware
	r.Use(auth.GinStdMiddleware(conf.Auth))

	r.POST("/business", request.BindCreate(businessCtrl.Create))
	r.GET("/business", request.BindGet(businessCtrl.List))
	r.PATCH("/business/:id", request.BindUpdate(businessCtrl.Update))

	// nested resources crud, note the `request.BindNestedCreate` for create action
	r.POST("/business/:parentID/products", request.BindNestedCreate(productController.Create))
	r.GET("/business/:parentID/products", request.BindGet(productController.List))
	r.PATCH("/business/:parentID/products/:id", request.BindUpdate(productController.Update))

	// start server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("could not start server", err)
	}
}
