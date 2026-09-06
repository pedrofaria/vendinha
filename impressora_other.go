//go:build !windows

package main

import "errors"

// Impressão térmica (e a enumeração de impressoras via alexbrainman/printer) é
// um recurso só do Windows. Noutros SO, devolve erro amigável para a UI saber
// que o seletor de impressora não está disponível.

func listaImpressoras() ([]string, error) {
	return []string{}, errors.New("listar impressoras só é suportado no Windows")
}

func impressoraPadrao() (string, error) {
	return "", errors.New("impressora padrão só é suportado no Windows")
}

func imprimirTexto(_ string, _ string) error {
	return errors.New("imprimir só é suportado no Windows")
}
