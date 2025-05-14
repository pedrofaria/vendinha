package purchase

import (
	"bytes"
	"fmt"
	"time"
	"vendinha/internal/adapter/db"
	"vendinha/internal/model"
	"vendinha/internal/service/print"

	"gorm.io/gorm"
)

type Service struct {
	printSrv *print.Service
}

func NewService(printSrv *print.Service) *Service {
	return &Service{printSrv: printSrv}
}

func (s *Service) PlacePurchase(purchase model.Purchase) error {
	for i, item := range purchase.Products {
		item.Total = item.ProductPrice * item.Quantity
		purchase.Total += item.Total
		purchase.Products[i] = item
	}

	return db.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := db.GetDB().Create(&purchase).Error; err != nil {
			return err
		}

		// return nil will commit the whole transaction
		return s.printRecipe(purchase)
	})
}

func (s *Service) printRecipe(purchase model.Purchase) error {
	p := bytes.NewBufferString("")

	ts := time.Now()

	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintln(p, "-   CEJV - Feira das Nacoes    -")
	fmt.Fprintf(p, "-     %s      -\n", ts.Format(time.DateTime))
	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintf(p, "** Pedido Nr.: %05d \n", purchase.ID)
	fmt.Fprint(p, "\n\n")

	for _, item := range purchase.Products {
		fmt.Fprintf(p, "%02d - %s\n", item.Quantity, item.ProductCode)
		fmt.Fprintf(p, "%s * %d = %s\n", frmt(item.ProductPrice), item.Quantity, frmt(item.Total))
		fmt.Fprint(p, "\n--------------------------------\n")
	}

	fmt.Fprintln(p, "")
	fmt.Fprintf(p, "TOTAL: %s\n\n\n", frmt(purchase.Total))

	return s.printSrv.Print(p.Bytes())
}

func frmt(val uint) string {
	if val < 100 {
		return fmt.Sprintf("R$ 0,%02d", val)
	}

	return fmt.Sprintf("R$ %d,%02d", uint(val/100), val%100)
}
