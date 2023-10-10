package crud

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

type (
	NestedModel[M Model] interface {
		Model
		ParentID() int
		SetParentID(int) M
	}

	// NestedResourceParam is used to bind nested uri like /parent/:parentID/child/:id
	// first param is expected to be named "parentID" and second as "id"
	NestedResourceParam struct {
		ID       int `uri:"id" binding:"required"`
		ParentID int `uri:"parentID" binding:"required"`
	}
	NestedParam struct {
		ParentID int `uri:"parentID"` // binding:"required"`
		ListParam
	}
	NestedResRequest[M NestedModel[M]] interface {
		ToModel() M
	}

	// NestedController represents crud controller taking model, request and response object
	// as type param. Response object dependency could be eliminated if model has a
	// method to convert to response
	// but that would move the contract definition to model, which is avoided here.
	// Path is expected to be in format /parent/:parentID/child/:id with exact param names
	NestedController[M NestedModel[M], S Response[M], R NestedResRequest[M]] struct {
		Svc CrudService[M]
	}
)

func (c *NestedController[M, S, R]) Create(ctx *gin.Context, p NestedParam, req R) (*S, error) {
	m := req.ToModel()
	m = m.SetParentID(p.ParentID)
	res, err := c.Svc.Create(ctx, m)
	fmt.Printf("res: %+v\n", res)

	var response S
	response, ok := response.FillFromModel(*res).(S)
	if !ok {
		panic("Invalid implementation of FillFromModel, it should return same type as implementor")
	}
	return &response, err
}

func (c *NestedController[M, S, R]) Retrieve(ctx *gin.Context, p NestedResourceParam) (*S, error) {
	res, err := c.Svc.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if (*res).ParentID() != p.ParentID {
		return nil, apperrors.NewPermissionError((*res).ResourceName())
	}
	var response S
	response, ok := response.FillFromModel(*res).(S)
	fmt.Println("response", p.ID, p.ParentID, res, response)
	if !ok {
		panic("Invalid implementation of FillFromModel, it should return same type as implementor")
	}
	return &response, nil
}

func (c *NestedController[M, S, R]) Update(ctx *gin.Context, p NestedResourceParam, req R) error {
	res, err := c.Svc.Get(ctx, p.ID)
	if err != nil {
		return err
	}
	if (*res).ParentID() != p.ParentID {
		return apperrors.NewPermissionError((*res).ResourceName())
	}

	_, err = c.Svc.Update(ctx, p.ID, req.ToModel())
	return err
}

func (c *NestedController[M, S, R]) Delete(ctx *gin.Context, p NestedResourceParam) error {
	res, err := c.Svc.Get(ctx, p.ID)
	if err != nil {
		return err
	}
	if (*res).ParentID() != p.ParentID {
		return apperrors.NewPermissionError((*res).ResourceName())
	}
	if err := c.Svc.Delete(ctx, p.ID); err != nil {
		return err
	}
	return nil
}

func (c NestedController[M, S, R]) List(ctx *gin.Context, p NestedParam) (*ListResponse[S], error) {
	items, total, err := c.Svc.List(ctx, p.After, p.Limit)
	if err != nil {
		return nil, err
	}
	var res []S
	for _, item := range items {
		var response S
		response, ok := response.FillFromModel(item).(S)
		if !ok {
			panic("Invalid implementation of FillFromModel, it should return same type as implementor")
		}
		res = append(res, response)
	}
	return &ListResponse[S]{Items: res, Total: total}, nil
}
