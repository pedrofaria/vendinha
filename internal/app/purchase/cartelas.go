package purchase

import _ "embed"

//go:embed cartelas/100.txt
var cartela100 []byte

//go:embed cartelas/50.txt
var cartela50 []byte

//go:embed cartelas/20.txt
var cartela20 []byte

//go:embed cartelas/10.txt
var cartela10 []byte

var cartelas = map[string][]byte{
	"CARTELA 100": cartela100,
	"CARTELA 50":  cartela50,
	"CARTELA 20":  cartela20,
	"CARTELA 10":  cartela10,
}
