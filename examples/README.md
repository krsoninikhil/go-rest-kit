# go-rest-kit example

This example package adds the rest API for CRUD actions for following types of resources
- Auth: API to sigup using phone and verify OTP and refresh the access tokens
- Independent resource (`BusinessType`): CRUD APIs for a resource which doesn't require any context information
- User context dependent resource (`Business`): CRUD for a resource where authenticated user id is required from context 
- Child resource (`Product`): CRUD for a resouce which is nested under another resource and 
  parent id is extracted from the api path