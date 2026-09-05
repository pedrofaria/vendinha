package main

// Modelos de domínio (pt-BR) da Vendinha.

// Evento é um evento (festa, retiro etc.) onde acontecem vendas.
type Evento struct {
	ID      int64  `json:"id"`
	Nome    string `json:"nome"`
	Ativo   bool   `json:"ativo"`
	CriadoEm string `json:"criadoEm"`
}

// EstoqueTipo descreve o tipo de controle de estoque de um produto.
const (
	EstoqueIlimitado = "ilimitado"
	EstoqueLimitado  = "limitado"
)

// Produto é um item à venda dentro de um evento. Estoque pode ser
// ilimitado (estoque_tipo='ilimitado', quantidade irrelevante) ou
// limitado (estoque_tipo='limitado', quantidade = unidades restantes).
type Produto struct {
	ID          int64  `json:"id"`
	EventoID    int64  `json:"eventoId"`
	Nome        string `json:"nome"`
	Preco       int64  `json:"preco"` // centavos
	EstoqueTipo string `json:"estoqueTipo"`
	Quantidade  int64  `json:"quantidade"`
	Ativo       bool   `json:"ativo"`
}

// PedidoItem é uma linha de um pedido. Nome/preço são snapshot no momento
// da venda para preservar histórico mesmo se o produto mudar depois.
type PedidoItem struct {
	ID        int64  `json:"id"`
	ProdutoID int64  `json:"produtoId"`
	Nome      string `json:"nome"`
	PrecoUnit int64  `json:"precoUnit"`
	Qtd       int64  `json:"qtd"`
	Subtotal  int64  `json:"subtotal"`
}

// Pedido é um pedido/venda fechado de um evento.
type Pedido struct {
	ID       int64        `json:"id"`
	EventoID int64        `json:"eventoId"`
	Numero   int64        `json:"numero"`
	Total    int64        `json:"total"` // centavos
	CriadoEm string       `json:"criadoEm"`
	Itens    []PedidoItem `json:"itens"`
}

// ProdutoVenda é a visão de um produto na tela de venda (PDV), já incluindo
// o estoque restante que a UI usa para barrar/avisar quantidades.
type ProdutoVenda struct {
	ID           int64  `json:"id"`
	Nome         string `json:"nome"`
	Preco        int64  `json:"preco"`
	EstoqueTipo  string `json:"estoqueTipo"`
	Quantidade   int64  `json:"quantidade"`
	Esgotado     bool   `json:"esgotado"`
}
