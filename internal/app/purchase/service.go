package purchase

import (
	"fmt"
	"time"
	"vendinha/internal/model"

	"github.com/alexbrainman/printer"
)

const printerName = "HPRT MPT-II"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (*Service) PlacePurchase(purchase model.Purchase) error {
	fmt.Println("Opaaaaa")
	fmt.Println(purchase)

	p, err := printer.Open(printerName) // Opens the named printer and returns a *Printer
	if err != nil {
		return err
	}

	if err := p.StartDocument("vendinha", "RAW"); err != nil {
		return err
	}

	if err := p.StartPage(); err != nil {
		return err
	}

	ts := time.Now()

	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintln(p, "CEJV - Feira Missionaria")
	fmt.Fprintln(p, ts.Format(time.DateTime))
	// fmt.Fprintln(p, "CHURRASCO COMPLETO          - 02")
	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintln(p, "")
	fmt.Fprintln(p, "")

	total := uint(0)

	for _, i := range purchase.Products {
		fmt.Fprintf(p, "%02d - %s\n", i.Quantity, i.ProductCode)
		fmt.Fprintf(p, "%s * %d = %s\n", frmt(i.ProductPrice), i.Quantity, frmt(i.ProductPrice*i.Quantity))
		fmt.Fprintln(p, "")
		fmt.Fprintln(p, "--------------------------------")
		fmt.Fprintln(p, "")

		total += i.ProductPrice * i.Quantity
	}

	fmt.Fprintln(p, "")
	fmt.Fprintf(p, "TOTAL: %s\n\n\n", frmt(total))

	if err := p.EndPage(); err != nil {
		return err
	}

	if err := p.EndDocument(); err != nil {
		return err
	}

	// close the resource
	if err := p.Close(); err != nil {
		return err
	}

	return nil
}

func frmt(val uint) string {
	if val < 100 {
		return fmt.Sprintf("R$ 0,%02d", val)
	}

	return fmt.Sprintf("R$ %d,%02d", uint(val/100), val%100)
}
