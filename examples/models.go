package main

import (
	"database/sql"
	"fmt"

	"github.com/krsoninikhil/go-rest-kit/auth"
	"github.com/krsoninikhil/go-rest-kit/pgdb"
)

type User struct {
	Name     string
	Email    sql.NullString `gorm:"uniqueIndex"`
	Phone    string         `gorm:"uniqueIndex"`
	DialCode string
	Country  string
	Locale   string
	pgdb.BaseModel
}

func (u User) ResourceName() string { return "user" }
func (u User) SetPhone(phone string) auth.UserModel {
	u.Phone = phone
	return u
}

// SetSignupInfo enables to set any extra fields that you might have as signup step
func (u User) SetSignupInfo(info auth.SigupInfo) auth.UserModel {
	u.DialCode = info.DialCode
	u.Country = info.Country
	u.Locale = info.Locale
	return u
}

// BusinessType is an example model without any user context
type BusinessType struct {
	Name string
	Icon string
	pgdb.BaseModel
}

func (b BusinessType) ResourceName() string { return fmt.Sprintf("%T", b) }

// Business is an example model with user context
type Business struct {
	Name           string
	BusinessTypeID int
	OwnerID        int
	pgdb.BaseModel

	Type  *BusinessType
	Owner *User
}

func (b Business) ResourceName() string { return "business" }
func (b Business) SetOwnerID(id int) Business {
	b.OwnerID = id
	return b
}

// Product is an example model for nested resources where Business is the parent model
type Product struct {
	Name       string
	BusinessID int
	pgdb.BaseModel

	Business *Business
}

func (p Product) ResourceName() string { return "product" }
func (p Product) GetName() string      { return p.Name }
func (p Product) ParentID() int        { return p.BusinessID }
func (p Product) SetParentID(id int) Product {
	p.BusinessID = id
	return p
}
