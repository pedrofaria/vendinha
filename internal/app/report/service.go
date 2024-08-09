package report

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"vendinha/internal/adapter/db"
	"vendinha/internal/service/print"
)

type Service struct {
	printSrv *print.Service
}

func NewService(printSrv *print.Service) *Service {
	return &Service{printSrv: printSrv}
}

func (*Service) hidrateReport() (*ReportPage, error) {
	rp := ReportPage{}

	if err := db.GetDB().Raw("SELECT count(*) from purchases").Scan(&rp.NrPurchases).Error; err != nil {
		return nil, err
	}

	if err := db.GetDB().Raw("SELECT sum(total) from purchases").Scan(&rp.TotalPurchases).Error; err != nil {
		return nil, err
	}

	if err := db.GetDB().Raw(`
		SELECT product_code, sum(total) total, sum(quantity) quantity
		FROM purchase_items
		GROUP BY product_code
		HAVING sum(quantity) > 0
		ORDER BY product_code
	`).Scan(&rp.PerProducts).Error; err != nil {
		return nil, err
	}

	return &rp, nil
}

func (s *Service) GetReportData() (*ReportPage, error) {
	return s.hidrateReport()
}

func (s *Service) PrintReport() error {
	rp, err := s.hidrateReport()
	if err != nil {
		return err
	}

	p := bytes.NewBufferString("")

	ts := time.Now()

	fmt.Fprintln(p, "--------------------------------")
	fmt.Fprintln(p, "-     RELATORIO DE VENDAS      -")
	fmt.Fprintf(p, "-     %s      -\n", ts.Format(time.DateTime))
	fmt.Fprint(p, "--------------------------------\n\n")

	fmt.Fprintf(p, "Nr. de Pedidos: %d\n", rp.NrPurchases)
	fmt.Fprintf(p, "Total  Entrada: %s\n\n", frmt(rp.TotalPurchases))

	fmt.Fprintln(p, "--- Totalizacao por Produtos ---")
	fmt.Fprintln(p, "--------------------------------")
	for _, item := range rp.PerProducts {
		fmt.Fprintln(p, item.ProductCode)
		qtd := fmt.Sprintf("%d", item.Quantity)
		total := frmt(item.Total)
		spaces := 30 - len(qtd) - len(total)

		fmt.Fprintf(p, "%s %s %s\n", qtd, strings.Repeat("_", spaces), total)

		fmt.Fprintln(p, "")
	}

	fmt.Fprint(p, "\n--------------------------------\n\n\n")

	return s.printSrv.Print(p.Bytes())
}

func frmt(val uint) string {
	if val < 100 {
		return fmt.Sprintf("R$ 0,%02d", val)
	}

	return fmt.Sprintf("R$ %d,%02d", uint(val/100), val%100)
}
