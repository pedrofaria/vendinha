package main

import "testing"

func newTestRepo(t *testing.T) *repo {
	t.Helper()
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepo(db)
}

func TestEventoProdutoPedido(t *testing.T) {
	r := newTestRepo(t)

	// Evento
	ev, err := r.CreateEvento("Retiro 2026")
	if err != nil {
		t.Fatal(err)
	}
	list, _ := r.ListEventos()
	if len(list) != 1 {
		t.Fatalf("esperava 1 evento, tem %d", len(list))
	}

	// Produto ilimitado e limitado
	ilim, err := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0)
	if err != nil {
		t.Fatal(err)
	}
	lim, err := r.CreateProduto(ev.ID, "Caneca", 2000, EstoqueLimitado, 3)
	if err != nil {
		t.Fatal(err)
	}

	// Visão de venda: caneca não esgotada
	venda, _ := r.ListProdutosVenda(ev.ID)
	if len(venda) != 2 {
		t.Fatalf("esperava 2 produtos na venda, tem %d", len(venda))
	}

	// Fechar pedido: 2 canecas + 1 pastel
	pedido, err := r.CriarPedido(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if pedido.Numero != 1 {
		t.Fatalf("numero esperado 1, tem %d", pedido.Numero)
	}
	itens := []PedidoItem{
		{ProdutoID: ilim.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: 1},
		{ProdutoID: lim.ID, Nome: "Caneca", PrecoUnit: 2000, Qtd: 2},
	}
	fechado, err := r.FecharPedido(pedido.ID, itens)
	if err != nil {
		t.Fatal(err)
	}
	if fechado.Total != 1500+2*2000 {
		t.Fatalf("total esperado 5500, tem %d", fechado.Total)
	}
	if len(fechado.Itens) != 2 {
		t.Fatalf("esperava 2 itens, tem %d", len(fechado.Itens))
	}

	// estoque da caneca baixou p/ 1
	lim2, _ := r.GetProduto(lim.ID)
	if lim2.Quantidade != 1 {
		t.Fatalf("estoque caneca esperado 1, tem %d", lim2.Quantidade)
	}

	// próximo pedido vende as 2 restantes -> deve falhar (só resta 1)
	p2, _ := r.CriarPedido(ev.ID)
	itens2 := []PedidoItem{{ProdutoID: lim.ID, Nome: "Caneca", PrecoUnit: 2000, Qtd: 2}}
	if _, err := r.FecharPedido(p2.ID, itens2); err == nil {
		t.Fatal("esperava erro de estoque insuficiente")
	}
}

func TestListEventosVazio(t *testing.T) {
	r := newTestRepo(t)
	list, err := r.ListEventos()
	if err != nil {
		t.Fatal(err)
	}
	if list == nil {
		t.Fatal("ListEventos vazio deve retornar slice vazio, não nil")
	}
	venda, err := r.ListProdutosVenda(999)
	if err != nil {
		t.Fatal(err)
	}
	if venda == nil {
		t.Fatal("ListProdutosVenda vazio deve retornar [] não nil")
	}
}
