package main

import (
	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/crud"
)

// request and response for business type CRUD
type (
	BusinessTypeRequest struct {
		Name string `json:"name" binding:"required"`
		Icon string `json:"icon"`
	}
	BusinessTypeResponse struct {
		BusinessTypeRequest
		ID int `json:"id"`
	}
)

// implement `crud.Request`
func (b BusinessTypeRequest) ToModel(_ *gin.Context) BusinessType {
	return BusinessType{Name: b.Name, Icon: b.Icon}
}

// implement `crud.Response`
func (b BusinessTypeResponse) FillFromModel(m BusinessType) crud.Response[BusinessType] {
	return BusinessTypeResponse{
		ID:                  m.ID,
		BusinessTypeRequest: BusinessTypeRequest{Name: m.Name, Icon: m.Icon},
	}
}

func (b BusinessTypeResponse) ItemID() int { return b.ID }

// request and response for product CRUD
type (
	ProductRequest struct {
		Name string `json:"name" binding:"max=100,required"`
	}
	ProductResponse struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		BusinessID int    `json:"business_id"`
	}
)

func (p ProductResponse) ItemID() int { return p.ID }

func (p ProductRequest) ToModel(_ *gin.Context) Product {
	return Product{Name: p.Name}
}

func (p ProductResponse) FillFromModel(m Product) crud.Response[Product] {
	return ProductResponse{
		ID:         m.ID,
		Name:       m.Name,
		BusinessID: m.BusinessID,
	}
}

// request and response for Business CRUD
type (
	BusinessRequest struct {
		Name   string `json:"name" binding:"required"`
		TypeID int    `json:"type_id" binding:"required"`
	}
	BusinessResponse struct {
		BusinessRequest
		ID int `json:"id"`
	}
)

func (b BusinessRequest) ToModel(_ *gin.Context) Business {
	return Business{Name: b.Name, BusinessTypeID: b.TypeID}
}

func (b BusinessResponse) FillFromModel(m Business) crud.Response[Business] {
	return BusinessResponse{
		ID:              m.ID,
		BusinessRequest: BusinessRequest{Name: m.Name, TypeID: m.BusinessTypeID},
	}
}

func (b BusinessResponse) ItemID() int { return b.ID }
