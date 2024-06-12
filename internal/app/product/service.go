package product

import (
	"time"
	"vendinha/internal/adapter/db"
	"vendinha/internal/model"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (*Service) List() []model.Product {
	var data []model.Product
	db.GetDB().Order("`order`, code ASC").Find(&data)

	return data
}

func (*Service) AddProduct(code string, price uint, order uint, sc string) error {
	product := model.Product{
		Code:  code,
		Price: price,
		Order: order,
		Key:   sc,
	}

	return db.GetDB().Save(&product).Error
}

func (*Service) DelProduct(id uint) error {
	return db.GetDB().Delete(&model.Product{BaseModel: model.BaseModel{ID: id}}).Error
}

func (*Service) UpdateProduct(id uint, code string, price uint, order uint, sc string) error {
	product := model.Product{
		BaseModel: model.BaseModel{
			ID:        id,
			UpdatedAt: time.Now(),
		},
		Code:  code,
		Price: price,
		Order: order,
		Key:   sc,
	}

	return db.GetDB().Save(&product).Error
}
