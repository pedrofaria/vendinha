package model

import (
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"ID"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type Product struct {
	BaseModel

	Code  string `json:"Code"`
	Price uint   `json:"Price"`
	Order uint   `json:"Order"`
	Key   string `json:"Key"`
}
