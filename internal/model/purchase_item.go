package model

type PurchaseItem struct {
	Quantity     uint   `json:"Quantity"`
	ProductID    uint   `json:"ProductID"`
	ProductCode  string `json:"ProductCode"`
	ProductPrice uint   `json:"ProductPrice"`
}
