package main

// Modelos de domínio (pt-BR) da Vendinha.

// Evento é um evento (festa, retiro etc.) onde acontecem vendas.
// VendeCartela habilita a venda de cartelas no PDV desse evento.
type Evento struct {
	ID           int64  `json:"id"`
	Nome         string `json:"nome"`
	Ativo        bool   `json:"ativo"`
	VendeCartela bool   `json:"vendeCartela"`
	CriadoEm     string `json:"criadoEm"`
}

// Cartela é um produto tipo "raspadinha"/cartela de prêmios vendida pelo
// valor (R$ 10, 20, 50 ou 100). É ilimitada (imprime-se sob demanda). Preco é
// o valor em centavos (= reais*100); Conteudo é o layout ASCII a imprimir.
type Cartela struct {
	Reais    int64  `json:"reais"`
	Preco    int64  `json:"preco"` // centavos = reais * 100
	Nome     string `json:"nome"`  // ex.: "Cartela R$ 10"
	Conteudo string `json:"conteudo"`
}

// EstoqueTipo descreve o tipo de controle de estoque de um produto.
const (
	EstoqueIlimitado = "ilimitado"
	EstoqueLimitado  = "limitado"
)

// Grupo agrupa produtos de um evento (ex.: Comida, Bebida, Sobremesa),
// com uma cor de identificação para o PDV e uma ordem de exibição.
type Grupo struct {
	ID       int64  `json:"id"`
	EventoID int64  `json:"eventoId"`
	Nome     string `json:"nome"`
	Cor      string `json:"cor"`
	Ordem    int64  `json:"ordem"`
}

// Produto é um item à venda dentro de um evento. Estoque pode ser
// ilimitado (estoque_tipo='ilimitado', quantidade irrelevante) ou
// limitado (estoque_tipo='limitado', quantidade = unidades restantes).
// GrupoID é o id do grupo (0 = sem grupo); Ordem ordena dentro do grupo;
// Atalho é a tecla (1-9) que adiciona o produto no PDV (0 = sem atalho).
// O atalho é único por evento: dois produtos do mesmo evento nunca compartilham
// a mesma tecla.
type Produto struct {
	ID          int64  `json:"id"`
	EventoID    int64  `json:"eventoId"`
	Nome        string `json:"nome"`
	Preco       int64  `json:"preco"` // centavos
	EstoqueTipo string `json:"estoqueTipo"`
	Quantidade  int64  `json:"quantidade"`
	Ativo       bool   `json:"ativo"`
	GrupoID     int64  `json:"grupoId"` // 0 = sem grupo
	Atalho      int64  `json:"atalho"`  // 0 = sem atalho; tecla 1-9
	Ordem       int64  `json:"ordem"`
}

// Formas de pagamento de um pedido fechado. 'anotaai' marca uma venda anotada
// (fiado) vinculada a uma Conta do mesmo evento — a pendência da conta é a soma
// dos totais dos pedidos anotaai ainda não quitados.
const (
	FormaDinheiro = "dinheiro"
	FormaCartao   = "cartao"
	FormaAnotaAi  = "anotaai"
)

// Conta é o "dono da conta" para vendas "Anota aí" (fiado) dentro de um evento.
// Uma conta pertence a um único evento; o nome é único por evento
// (comparação case-insensitive) — nomes diferentes ("Maria", "Maria Souza") são
// contas distintas, e "maria" digitar é tratado como a mesma conta de "Maria".
type Conta struct {
	ID            int64  `json:"id"`
	EventoID      int64  `json:"eventoId"`
	Nome          string `json:"nome"`
	TotalPendente int64  `json:"totalPendente"` // soma dos totais dos pedidos anotaai da conta (centavos)
	CriadoEm      string `json:"criadoEm"`
}

// PedidoItem é uma linha de um pedido. Nome/preço são snapshot no momento
// da venda para preservar histórico mesmo se o produto mudar depois.
// Uma linha é OU de produto (ProdutoID > 0, CartelaReais == 0) OU de cartela
// (CartelaReais > 0, ProdutoID == 0). CartelaReais guarda a denominação
// vendida (10/20/50/100) para regerar a impressão da cartela.
type PedidoItem struct {
	ID           int64  `json:"id"`
	ProdutoID    int64  `json:"produtoId"`
	CartelaReais int64  `json:"cartelaReais"`
	Nome         string `json:"nome"`
	PrecoUnit    int64  `json:"precoUnit"`
	Qtd          int64  `json:"qtd"`
	Subtotal     int64  `json:"subtotal"`
}

// Pedido é um pedido/venda fechado de um evento. Forma é a forma de pagamento
// ('dinheiro' | 'cartao' | 'anotaai'); ContaID é a conta vinculada quando a forma
// é 'anotaai' (0 = sem conta, p.ex. dinheiro/cartão).
// CanceladoEm marca um pedido cancelado ("" = ativo). Pedidos cancelados saem
// de toda soma/relatório (receita, pendência da conta) e o estoque limitado
// deles é devolvido; seguem na listagem de gerenciamento como histórico.
type Pedido struct {
	ID          int64        `json:"id"`
	EventoID    int64        `json:"eventoId"`
	Numero      int64        `json:"numero"`
	Total       int64        `json:"total"` // centavos
	Forma       string       `json:"forma"`
	ContaID     int64        `json:"contaId"`   // > 0 quando venda anotada (anotaai)
	QuitadoEm   string       `json:"quitadoEm"` // "" = em aberto; preenchido quando a venda anotada foi paga
	CanceladoEm string       `json:"canceladoEm"`
	CriadoEm    string       `json:"criadoEm"`
	Itens       []PedidoItem `json:"itens"`
}

// PedidoResumo é uma linha leve da listagem gerenciada de pedidos de um evento.
// Ao contrário de Pedido, traz o nome da conta do "Anota aí" (ContaNome) quando
// houver, sem carregar os itens (use GetPedido para o detalhe de um pedido).
type PedidoResumo struct {
	ID          int64  `json:"id"`
	EventoID    int64  `json:"eventoId"`
	Numero      int64  `json:"numero"`
	Total       int64  `json:"total"` // centavos
	Forma       string `json:"forma"`
	ContaID     int64  `json:"contaId"`
	ContaNome   string `json:"contaNome"` // preenchido quando a forma é 'anotaai'
	QuitadoEm   string `json:"quitadoEm"`
	CanceladoEm string `json:"canceladoEm"`
	CriadoEm    string `json:"criadoEm"`
}

// ListaPedidos é a resposta paginada da listagem de pedidos de um evento
// (mais recentes primeiro). Total é o nº de pedidos que casam com o filtro
// (independente da página), usado para desenhar a paginação.
type ListaPedidos struct {
	Pedidos      []PedidoResumo `json:"pedidos"`
	Total        int64          `json:"total"`
	Pagina       int64          `json:"pagina"`
	TotalPaginas int64          `json:"totalPaginas"`
}

// ProdutoVenda é a visão de um produto na tela de venda (PDV), já incluindo
// o estoque restante que a UI usa para barrar/avisar quantidades.
type ProdutoVenda struct {
	ID          int64  `json:"id"`
	Nome        string `json:"nome"`
	Preco       int64  `json:"preco"`
	EstoqueTipo string `json:"estoqueTipo"`
	Quantidade  int64  `json:"quantidade"`
	Esgotado    bool   `json:"esgotado"`
	Atalho      int64  `json:"atalho"` // 0 = sem atalho; tecla 1-9
}

// SemGrupoID marca o grupo virtual "Sem grupo" (produtos sem grupo_id).
const SemGrupoID = 0

// Cores auxiliares: CorPadraoGrupo é usada quando um grupo é criado sem cor;
// CorSemGrupo colore o bucket virtual "Sem grupo".
const (
	CorPadraoGrupo = "#10b981"
	CorSemGrupo    = "#9ca3af"
)

// GrupoProdutos é um grupo com seus produtos ordenados (tela de gerenciamento).
// Representa também o grupo virtual "Sem grupo" (Grupo.ID == SemGrupoID).
type GrupoProdutos struct {
	Grupo
	Produtos []Produto `json:"produtos"`
}

// GrupoVenda é um grupo com seus produtos ativos ordenados, como o PDV consome.
type GrupoVenda struct {
	Grupo
	Produtos []ProdutoVenda `json:"produtos"`
}

// ContaSaldo é a visão da conta na tela "Anota aí": quanto deve em aberto
// (vendas anotadas não quitadas), quantas vendas em aberto e quanto já quitou.
// TotalPendente = soma dos totais dos pedidos anotaai com quitado_em NULL.
type ContaSaldo struct {
	ID            int64  `json:"id"`
	EventoID      int64  `json:"eventoId"`
	Nome          string `json:"nome"`
	TotalPendente int64  `json:"totalPendente"` // centavos ainda devidos
	NumAberto     int64  `json:"numAberto"`     // nº de vendas anotadas em aberto
	TotalQuitado  int64  `json:"totalQuitado"`  // centavos já quitados
	CriadoEm      string `json:"criadoEm"`
}

// HoraVendas é uma fatia do histograma de vendas por hora do evento.
type HoraVendas struct {
	Hora   int64 `json:"hora"`
	Vendas int64 `json:"vendas"`
}

// ResumoEvento é o relatório exibido no Dashboard de um evento (só considera
// pedidos fechados com total > 0).
type ResumoEvento struct {
	ReceitaTotal        int64        `json:"receitaTotal"`        // soma dos totais dos pedidos fechados (centavos)
	NumPedidos          int64        `json:"numPedidos"`          // nº de vendas fechadas
	NumProdutosVendidos int64        `json:"numProdutosVendidos"` // unidades de produto vendidas (sem cartelas)
	VendasPorHora       []HoraVendas `json:"vendasPorHora"`       // histograma por hora do dia (0–23)
}
