package model

type Purchase struct {
	BaseModel

	Products []PurchaseItem `json:"Products"`
	Total    uint           `json:"Total"`
}
