# go-rest-kit
Ready made packages for quickly setting up REST APIs in Go. 
Specially avoiding to write handlers and request parsing for every new API you want.
Limitation here is most packages are helful if you're using Gin for your APIs.

## Motivation
Frameworks like FastAPI for Python are able to provide automatic request parsing and validation based on type hints
and Python insn't even a typed language. I could not find any framework providing similar feature in Go, i.e. not 
having to repeat the same parsing and error handling in every API handler. 

I should be able to write following handler / controller similar to FastAPI by just defining `SpecificRequest` and `SpecificResponse` i.e. get request body type as argument and return response type along with error:
```go
func controller(ctx context.Context, request SpecificRequest) (*SpecificResponse, error) {}
```

## Setup
This module is written with generic usecases in mind, which serves well in most usecase e.g. CRUD APIs. 
If you have a custom requirement, it's recommended to clone the repo in your development machine and use the local
version of the module by replacing module path. This allows you extend and customize 
the types and methods provided as per your requirement.

```bash
go get github.com/krsoninikhil/go-rest-kit
git clone github.com/krsoninikhil/go-rest-kit ../
go mod edit -replace github.com/krsoninikhil/go-rest-kit=../go-rest-kit
```

## Example

### 1. Exposing CRUD APIs
Exposing simple CRUD APIs is as simple as following
- Ensure your model implements `crud.Model`, you can embed `pgdb.BaseModel` to get default implementation for few methods.
- Define your request and response types and have them implement `crud.Request` and `crud.Response` interfaces.
Best part here is -- you can replace any component (controller, usecase or dao) with your own implementation.

> This example also shows use of request binding methods, and how you can avoid writing repeated code for parsing request body.

```go
// models.go
type BusinessType struct {
    ID int
    Name int
    Icon string
    pgdb.BaseModel
}

func (b BusinessType) ResourceName() string { return fmt.Sprintf("%T", b) }

// entities.go
type (
    BusinessTypeRequest {
        Name string `json:"name" binding:"required"`
        Icon string `json:"icon"`
    }
    BusinessTypeResponse {
        BusinessTypeRequest
        ID int `json:"id"`
    }
)

// implement `crud.Request`
func (b BusinessTypeRequest) ToModel(_ *gin.Context) BusinessType {
    return BusinessType{Name: b.Name, Icon: b.Icon}
}

// implement `crud.Response`
func (b BusinessTypeResponse) FillFromModel(m models.BusinessType) crud.Response[models.BusinessType] {
    return BusinessTypeResponse{
        ID: m.ID, 
        BusinessTypeRequest: BusinessTypeRequest{Name: m.Name, Icon: m.Icon},
    }
}
func (b BusinessTypeResponse) ItemID() int { return b.ID }

// setup routes
func main() {
    businessTypeDao        = crud.Dao[models.BusinessType]{PGDB: s.pgdb} // use your own doa if you need custom implementation
	businessTypeCtlr = crud.Controller[models.BusinessType, types.BusinessTypeResponse, types.BusinessTypeRequest]{
        Svc: &businessTypeDao, // using dao for service as no business logic is required here
    } // prewritten controller struct with CRUD methods
    
    r := gin.Default()
	r.GET("/business-types", request.BindGet(businessTypeCtlr.Retrieve))
    r.GET("/business-types", request.BindGet(businessTypeCtlr.List))
	r.POST("/business-types", request.BindCreate(businessTypeCtlr.Create))
	r.PATCH("/business-types", request.BindUpdate(businessTypeCtlr.Create))
	r.DELET("/business-types", request.BindDelete(businessTypeCtlr.Delete))
    // start your server
}
```

### 2. Load Your Application Config
Configs are loaded from yaml files where empty values are overriden from environment, which is set using `.env` file.
e.g. if `redis.password` in your yaml is empty, it will be set by `REDIS_PASSWORD` env value. Neat, Hmm?

```go
// define application config
type Config {
    DB pgdb.Config
    Twilio twilio.Config
    Auth auth.Config
    Env string
}
// implement `config.AppConfig`
func (c *Config) EnvPath() string    { return "./.env" }
func (c *Config) SourcePath() string { return fmt.Sprintf("./api/%s.yml", c.Env) }
func (c *Config) SetEnv(env string) { c.Env = env } // embed `config.BaseConfig` to avoid defining SetEnv

func main() {
    var (
        ctx = context.Background()
	    conf Config
    )
	config.Load(ctx, &conf)
    // that's about it, you use the `conf` now, e.g. see below usage for creating postgres connection
    db := pgdb.NewPGConnection(ctx, conf.DB)
}
```

### 3. Exposing Auth APIs -- Signup, OTP Verification and Token Refresh
> This also shows an example use of `pgdb` and `twillio` packages.

Assuming you have a `models.User` already defined, exposing auth APIs are just about defining your routes. Easy peazy!

```go
func main() {
    // reusing gin engine `r`, configuration `conf` and postgres connection `db` from above examples
    var (
        cache          = cache.NewInMemory() // im memory cache impelementation is provided for the purpose of examples
        // injecting dependencies, use your own implementation for any object if default isn't enough for your usecase
		userDao        = auth.NewUserDao[models.User](db)
		authSvc        = auth.NewService(conf.Auth, userDao)
		smsProvider    = twilio.NewClient(conf.Auth.Twilio)
		otpSvc         = auth.NewOTPSvc(conf.Auth.OTP, smsProvider, cache) // 
		authController = auth.NewController(authSvc, otpSvc, cache)
	)

	r.POST("/auth/otp/send", request.BindCreate(authController.SendOTP))
	r.POST("/auth/otp/verify", request.BindCreate(authController.VerifyOTP))
	r.POST("/auth/token/refresh", request.BindCreate(authController.RefreshToken))
	r.GET("/countries/:alpha2Code", request.BindGet(authController.CountryInfo))
    // start your server
}
```

### _Getting Rid Of Repeated Code For Request Parsing In Every Handler
If you usecase is not a typical CRUD, don't worry, you can still use the generic binding from `request` package.
While `request.BindGet` would parse only the query string for GET request, `request.BindAll` would parse values
from request uri, body and query based on the tags defined in request struct.

```go
func main() {
    r := gin.Default()
    r.GET("/business-types/insights", request.BindGet(businessTypeInsights))
    r.GET("/business-types/insights-2", request.BindAll(businessTypeInsights))
}

type (
    RequestType struct {}
    ResponseType struct {}
)

func businessTypeInsights(c *gin.Context, req RequestType) (*ResponseType, error) {
    // your customer controller
    return nil, apperrors.NewServerError(errors.New("not implemented"))
}

```

## Packages Implementation Details

- `request`: Provides parameter binding based on defined request type, this allows controllers to receive the request parameters and body as a argument and not have to parse and unmarshal the request in every controller.
    - `request.WithBinding` takes a fast-api like controller as argument and converts it to a `gin` controller.
    - `request.BindGet`, `request.BindCreate`, `request.BindUpdate` and `request.BindDelete` all takes a method and converts it go `gin` controller while providing parsed and validated request body to the function argument. Since these binding methods require the function signature to be defined, it assumes that the `Get` and `Delete` binding expects the argument struct to be parsed from URI and query params while `Update` and `Delete` exepects a struct for parsing URI and another for request body. See example to a better idea of usage.

- `crud`: Provides controllers for any resource like which request typical CRUD apis. These controller methods follow the signature that can be used directly with above explained `request` package binding methods. CRUD apis for any new model become just about registering these controllers with router. See example.
    - `crud.Controller`: Controller for a resource like `GET /resource`
    - `crud.NestedController`: Controller for a nested resources like `GET /parent/:parentID/resource`

- `apperrors`: Provides error that any typical API exposing application will require. Idea is to add more as per your requirement.

- `config`: Provides quick method to load and parse your config files to the provide struct. See example.

- `pgdb`: Provides config and constructor to create a new connection. See example.
  
- `integrations`: Provides frequently used third party client like Twilio for sending OTPs.
  
- `auth`: Almost all backend apps will require API to signup by a mobile no. and respond with JWT token on OTP verification. This also comes with controller for refreshing the tokens.

