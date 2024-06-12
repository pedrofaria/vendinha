package model

type Purchase struct {
	BaseModel

	Products []PurchaseItem `json:"Products" gorm:"many2many:purchase_products;"`
}
