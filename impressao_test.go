package main

import (
	"bytes"
	"testing"
)

func TestMontaEscposIniciaEEncerraComandos(t *testing.T) {
	b := montaEscpos("Linha 1\nLinha 2")
	if !bytes.HasPrefix(b, []byte("\x1b\x40")) { // ESC @ (inicializa)
		t.Fatalf("montaEscpos não começa com ESC @: % x", b[:min(4, len(b))])
	}
	// logo depois do init vem ESC t 2 (code page PC850)
	if !bytes.Contains(b, []byte("\x1b\x74\x02")) {
		t.Fatalf("montaEscpos não inclui a seleção de code page PC850 (ESC t 2): % x", b)
	}
	if !bytes.HasSuffix(b, []byte("\x1d\x56\x41")) { // GS V 65 (avança e corta)
		t.Fatalf("montaEscpos não termina com o comando de corte (GS V): % x", b[40:])
	}
}

func TestMontaEscposMantemLinhasEQuebras(t *testing.T) {
	b := montaEscpos("ABC")
	if !bytes.Contains(b, []byte("ABC")) {
		t.Fatalf("texto ASCII puro foi alterado: % x", b)
	}
	if !bytes.Contains(b, []byte("ABC\n")) {
		t.Fatalf("linha não terminou com \n: % x", b)
	}
}

func TestMontaEscposCodificaAcentosPC850(t *testing.T) {
	b := montaEscpos("Açúcar")
	// 'ç' em PC850 é 0x87; se viesse UTF-8 cru apareceria o byte 0xC3.
	if bytes.IndexByte(b, 0xC3) >= 0 {
		t.Fatalf("encontrou byte UTF-8 (0xC3) — acento não foi convertido p/ PC850: % x", b)
	}
	if !bytes.Contains(b, []byte{0x87}) {
		t.Fatalf("'ç' não virou 0x87 em PC850: % x", b)
	}
}
