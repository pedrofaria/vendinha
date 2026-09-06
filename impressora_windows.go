//go:build windows

package main

import (
	"fmt"

	"github.com/alexbrainman/printer"
)

// listaImpressoras devolve os nomes das impressoras instaladas no sistema.
// A impressão térmica (MTP II / ESC-POS) é um recurso só do Windows.
func listaImpressoras() ([]string, error) {
	return printer.ReadNames()
}

// impressoraPadrao devolve o nome da impressora padrão do Windows.
func impressoraPadrao() (string, error) {
	return printer.Default()
}

// imprimirTexto envia `texto` como job RAW/ESC-POS para a impressora `nome`,
// pelo spooler do Windows (sem passar por diálogo de impressão). Espera-se que
// a térmica (MTP II) esteja com o driver ESC/POS do fabricante.
func imprimirTexto(nome, texto string) error {
	p, err := printer.Open(nome)
	if err != nil {
		return fmt.Errorf("abrir a impressora %q: %w", nome, err)
	}
	defer p.Close()

	// Datatype "RAW" EXPLÍCITO — replicando o teste que funcionou na HPRT MTP-II.
	// NÃO usar p.StartRawDocument(): ele lê DriverInfo e, se o driver for v4/XPS,
	// troca o datatype para "XPS_PASS", que não imprime bytes ESC/POS crus.
	if err := p.StartDocument("Vendinha", "RAW"); err != nil {
		return fmt.Errorf("iniciar o job de impressão: %w", err)
	}
	if err := p.StartPage(); err != nil {
		return fmt.Errorf("iniciar a página de impressão: %w", err)
	}
	if _, err := p.Write(montaEscpos(texto)); err != nil {
		return fmt.Errorf("enviar os dados à impressora: %w", err)
	}
	if err := p.EndPage(); err != nil {
		return fmt.Errorf("finalizar a página: %w", err)
	}
	if err := p.EndDocument(); err != nil {
		return fmt.Errorf("finalizar o job de impressão: %w", err)
	}
	return nil
}
