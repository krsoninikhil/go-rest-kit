# go-rest-kit
Ready made packages for quickly setting up REST APIs in Go.

## Setup
This module, in most cases serves as a prewritten boilerplate code for your Go app.
So it's recommended to clone the repo in your development machine and use the local
version of the module by replacing module path. This allows you extend and customize 
the struct and methods provided as per your requirement.

```bash
go get github.com/krsoninikhil/go-rest-kit
git clone github.com/krsoninikhil/go-rest-kit ../
go mod edit -replace github.com/krsoninikhil/go-rest-kit=../go-rest-kit
```

## Packages

- `request`: Provides parameter binding based on defined request type, this allows controllers to receive the request parameters and body as a argument and not have to parse and unmarshal the request in every controller.
    - `request.WithBinding` takes a fast-api like controller as argument and converts it to a `gin` controller.
    - `request.BindGet`, `request.BindCreate`, `request.BindUpdate` and `request.BindDelete` all takes a method and converts it go `gin` controller while providing parsed and validated request body to the function argument. Since these binding methods require the function signature to be defined, it assumes that the `Get` and `Delete` binding expects the argument struct to be parsed from URI and query params while `Update` and `Delete` exepects a struct for parsing URI and another for request body. See example to a better idea of usage.

- `crud`: Provides controllers for any resource like which request typical CRUD apis. These controller methods follow the signature that can be used directly with above explained `request` package binding methods. CRUD apis for any new model become just about registering these controllers with router. See example.
    - `crud.Controller`: Controller for a resource like `GET /resource`
    - `crud.NestedController`: Controller for a nested resources like `GET /parent/:parentID/resource`

- `apperrors`: Provides error that any typical API exposing application will require. Idea is to add more as per your requirement.

- `config`: Provides quick method to load and parse your config files to the provide struct. See example.

- `pgdb`: Provides config and constructor to create a new connection. See example.