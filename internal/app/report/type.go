package report

import "vendinha/internal/model"

type ReportPage struct {
	NrPurchases    uint                 `json:"NrPurchases"`
	TotalPurchases uint                 `json:"TotalPurchases"`
	PerProducts    []model.PurchaseItem `json:"PerProducts"`
}
