package model

type PurchaseItem struct {
	BaseModel

	Quantity     uint   `json:"Quantity"`
	ProductID    uint   `json:"ProductID"`
	ProductCode  string `json:"ProductCode"`
	ProductPrice uint   `json:"ProductPrice"`
	PurchaseID   uint   `json:"PurchaseID"`
	Total        uint   `json:"Total"`
}
