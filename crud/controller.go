package crud

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type (
	Request[M Model] interface {
		ToModel(*gin.Context) M
	}
	Response[M Model] interface {
		FillFromModel(m M) Response[M]
		ItemID() int
	}

	Controller[M Model, S Response[M], R Request[M]] struct {
		Svc Service[M]
	}
)

func (c *Controller[M, S, R]) Create(ctx *gin.Context, req R) (*S, error) {
	m := req.ToModel(ctx)
	res, err := c.Svc.Create(ctx, m)
	if err != nil {
		return nil, err
	}

	var response S
	response, ok := response.FillFromModel(*res).(S)
	if !ok {
		panic("Invalid implementation of FillFromModel, it should return same type as implementor")
	}
	return &response, err
}

func (c *Controller[M, S, R]) Retrieve(ctx *gin.Context, p ResourceParam) (*S, error) {
	res, err := c.Svc.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	var response S
	response, ok := response.FillFromModel(*res).(S)
	if !ok {
		panic("Invalid implementation of FillFromModel, it should return same type as implementor")
	}
	return &response, err
}

func (c *Controller[M, S, R]) Update(ctx *gin.Context, p ResourceParam, req R) error {
	m := req.ToModel(ctx)
	_, err := c.Svc.Update(ctx, p.ID, m)
	return err
}

func (c *Controller[M, S, R]) Delete(ctx *gin.Context, id int) error {
	return c.Svc.Delete(ctx, id)
}

func (c *Controller[M, S, R]) List(ctx *gin.Context, p ListParam) (*ListResponse[S], error) {
	var pageItem S
	if _, ok := any(pageItem).(PageItem); !ok {
		return nil, errors.New("list response type must implement PageItem interface")
	}

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
	return &ListResponse[S]{Items: res, Total: total, NextAfter: GetLastItemID(res)}, nil
}

func GetLastItemID[T PageItem](items []T) int {
	var res int
	for _, item := range items {
		if item.ItemID() > res {
			res = item.ItemID()
		}
	}
	return res
}
