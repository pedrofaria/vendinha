package purchase

import (
	"bytes"
	"fmt"
	"io"
	"time"
	"vendinha/internal/adapter/db"
	"vendinha/internal/model"

	"gorm.io/gorm"
)

type Printer interface {
	Print(data []byte) error
}

type Service struct {
	printSrv Printer
}

func NewService(printSrv Printer) *Service {
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

func (s *Service) PlaceCartelaPurchase(purchase model.Purchase) error {
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
		return s.printCartelaRecipe(purchase)
	})
}

func (s *Service) printRecipe(purchase model.Purchase) error {
	p := bytes.NewBufferString("")

	s.printHeader(p, purchase.ID)

	for _, item := range purchase.Products {
		fmt.Fprintf(p, "%02d - %s\n", item.Quantity, item.ProductCode)
		fmt.Fprintf(p, "%s * %d = %s\n", frmt(item.ProductPrice), item.Quantity, frmt(item.Total))
		fmt.Fprint(p, "\n--------------------------------\n")
	}

	fmt.Fprintln(p, "")
	fmt.Fprintf(p, "TOTAL: %s\n\n\n", frmt(purchase.Total))

	return s.printSrv.Print(p.Bytes())
}

func (s *Service) printCartelaRecipe(purchase model.Purchase) error {
	p := bytes.NewBufferString("")

	s.printHeader(p, purchase.ID)

	for _, item := range purchase.Products {
		data, ok := cartelas[item.ProductCode]
		if !ok {
			return fmt.Errorf("cartela não encontrada: %s", item.ProductCode)
		}

		for i := uint(0); i < item.Quantity; i++ {
			p.Write(data)
			fmt.Fprintf(p, "\n\n\n")
		}
	}

	return s.printSrv.Print(p.Bytes())
}

func (s *Service) printHeader(p io.Writer, purchaseId uint) {
	ts := time.Now()

	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintln(p, "-   CEJV - Feira das Nacoes    -")
	fmt.Fprintf(p, "-     %s      -\n", ts.Format(time.DateTime))
	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintf(p, "** Pedido Nr.: %05d \n", purchaseId)
	fmt.Fprint(p, "\n\n")
}

func frmt(val uint) string {
	if val < 100 {
		return fmt.Sprintf("R$ 0,%02d", val)
	}

	return fmt.Sprintf("R$ %d,%02d", uint(val/100), val%100)
}
