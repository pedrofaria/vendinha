package main

import (
	"embed"
	"fmt"
	"strings"
)

// Os layouts ASCII das cartelas ficam em cartelas/<valor>.txt e são embutidos
// no binário (go:embed) — fonte única do conteúdo a imprimir em cada cartela.
//
//go:embed cartelas/*.txt
var cartelasFS embed.FS

// CartelaReaisValidos são as denominações de cartela existentes, na ordem de
// exibição no PDV. Os arquivos cartelas/<reais>.txt devem existir para todos.
var CartelaReaisValidos = []int64{10, 20, 50, 100}

// cartelaPorReais monta a Cartela da denominação `reais` lendo o layout do
// arquivo embutido cartelas/<reais>.txt.
func cartelaPorReais(reais int64) (Cartela, error) {
	data, err := cartelasFS.ReadFile(fmt.Sprintf("cartelas/%d.txt", reais))
	if err != nil {
		return Cartela{}, err
	}
	return Cartela{
		Reais:    reais,
		Preco:    reais * 100,
		Nome:     fmt.Sprintf("Cartela R$ %d", reais),
		Conteudo: strings.TrimRight(string(data), "\r\n"),
	}, nil
}

// ehDenominacaoCartela diz se `reais` é uma denominação de cartela conhecida.
func ehDenominacaoCartela(reais int64) bool {
	for _, v := range CartelaReaisValidos {
		if v == reais {
			return true
		}
	}
	return false
}
