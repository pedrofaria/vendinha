package main

import (
	"bytes"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// Comandos ESC/POS usados para imprimir o recibo numa térmica (MTP II).
const (
	escInit     = "\x1b\x40"     // ESC @  — inicializa a impressora
	escCp850    = "\x1b\x74\x02" // ESC t 2 — code page PC850 (acentos do pt-BR)
	escCorte    = "\x1d\x56\x41" // GS V 65 — avança o papel e corta
	linhasCorte = 3              // nº de linhas em branco antes do corte
)

// montaEscpos transforma o texto do recibo (linhas separadas por \n) numa
// sequência ESC/POS pronta para enviar à impressora térmica: inicializa o
// dispositivo, fixa a code page PC850 (para os acentos do português), imprime
// cada linha e, no fim, alimenta o papel e corta. As linhas já chegam
// alinhadas/formatadas na largura certa (montadas no frontend).
func montaEscpos(texto string) []byte {
	var b bytes.Buffer
	b.WriteString(escInit)
	b.WriteString(escCp850)
	for _, linha := range strings.Split(texto, "\n") {
		b.Write(paraCP850(linha))
		b.WriteByte('\n')
	}
	for i := 0; i < linhasCorte; i++ {
		b.WriteByte('\n')
	}
	b.WriteString(escCorte)
	return b.Bytes()
}

// paraCP850 codifica uma linha (UTF-8 vinda do frontend) na code page 850, que
// é a usada pela impressora. Caracteres sem representação viram '?'.
func paraCP850(s string) []byte {
	b, _ := charmap.CodePage850.NewEncoder().Bytes([]byte(s))
	return b
}
