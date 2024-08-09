package print

import (
	"github.com/alexbrainman/printer"
)

const printerName = "HPRT MPT-II"

type Service struct {
	printerName string
}

func NewService() *Service {
	return &Service{
		printerName: printerName,
	}
}

func (s *Service) Print(data []byte) error {
	p, err := printer.Open(s.printerName) // Opens the named printer and returns a *Printer
	if err != nil {
		return err
	}

	if err := p.StartDocument("vendinha", "RAW"); err != nil {
		return err
	}

	if err := p.StartPage(); err != nil {
		return err
	}

	p.Write(data)

	if err := p.EndPage(); err != nil {
		return err
	}

	if err := p.EndDocument(); err != nil {
		return err
	}

	// close the resource
	return p.Close()
}
