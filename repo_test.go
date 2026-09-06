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
	ev, err := r.CreateEvento("Retiro 2026", false)
	if err != nil {
		t.Fatal(err)
	}
	list, _ := r.ListEventos()
	if len(list) != 1 {
		t.Fatalf("esperava 1 evento, tem %d", len(list))
	}

	// Produto ilimitado e limitado (sem grupo)
	ilim, err := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	lim, err := r.CreateProduto(ev.ID, "Caneca", 2000, EstoqueLimitado, 3, 0, 2)
	if err != nil {
		t.Fatal(err)
	}

	// Visão de venda: caneca não esgotada. Sem grupos, tudo cai no bucket
	// virtual "Sem grupo" (1 grupo com 2 produtos).
	venda, _ := r.ListProdutosVenda(ev.ID)
	if len(venda) != 1 {
		t.Fatalf("esperava 1 grupo (sem grupo) na venda, tem %d", len(venda))
	}
	if len(venda[0].Produtos) != 2 {
		t.Fatalf("esperava 2 produtos no grupo, tem %d", len(venda[0].Produtos))
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
	fechado, err := r.FecharPedido(pedido.ID, itens, FormaDinheiro, "")
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
	if _, err := r.FecharPedido(p2.ID, itens2, FormaDinheiro, ""); err == nil {
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

func TestGruposOrdenacao(t *testing.T) {
	r := newTestRepo(t)
	ev, err := r.CreateEvento("Feira", false)
	if err != nil {
		t.Fatal(err)
	}

	comida, _ := r.CreateGrupo(ev.ID, "Comida", "#ef4444")
	bebida, _ := r.CreateGrupo(ev.ID, "Bebida", "#3b82f6")
	if comida.Ordem != 1 || bebida.Ordem != 2 {
		t.Fatalf("ordem dos grupos esperada 1,2, tem %d,%d", comida.Ordem, bebida.Ordem)
	}

	pastel, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, comida.ID, 0)
	caneca, _ := r.CreateProduto(ev.ID, "Caneca", 2000, EstoqueLimitado, 3, comida.ID, 0)
	suco, _ := r.CreateProduto(ev.ID, "Suco", 800, EstoqueIlimitado, 0, bebida.ID, 0)
	solto, _ := r.CreateProduto(ev.ID, "Chaveiro", 500, EstoqueLimitado, 5, 0, 0)

	// Listagem: grupos na ordem, produtos dentro ordenados, "Sem grupo" por último.
	grupos, _ := r.ListProdutos(ev.ID)
	if len(grupos) != 3 {
		t.Fatalf("esperava 3 grupos (comida, bebida, sem grupo), tem %d", len(grupos))
	}
	if grupos[0].Nome != "Comida" || grupos[1].Nome != "Bebida" || grupos[2].ID != SemGrupoID {
		t.Fatalf("ordem/agrupamento inesperado: %s, %s, %d", grupos[0].Nome, grupos[1].Nome, grupos[2].ID)
	}
	if eq := idsDe(grupos[0].Produtos); !equalInts(eq, []int64{pastel.ID, caneca.ID}) {
		t.Fatalf("produtos de Comida esperado [%d %d], tem %v", pastel.ID, caneca.ID, eq)
	}
	if eq := idsDe(grupos[2].Produtos); !equalInts(eq, []int64{solto.ID}) {
		t.Fatalf("Sem grupo esperado [%d], tem %v", solto.ID, eq)
	}

	// Sobe Bebida -> fica antes de Comida.
	if err := r.MoverGrupo(bebida.ID, -1); err != nil {
		t.Fatal(err)
	}
	grupos, _ = r.ListProdutos(ev.ID)
	if grupos[0].Nome != "Bebida" || grupos[1].Nome != "Comida" {
		t.Fatalf("após mover, esperado Bebida,Comida; tem %s,%s", grupos[0].Nome, grupos[1].Nome)
	}

	// Dentro de Comida: sobe Caneca -> fica antes de Pastel.
	if err := r.MoverProduto(caneca.ID, -1); err != nil {
		t.Fatal(err)
	}
	grupos, _ = r.ListProdutos(ev.ID)
	comidaG := grupoPorID(grupos, comida.ID)
	if eq := idsDe(comidaG.Produtos); !equalInts(eq, []int64{caneca.ID, pastel.ID}) {
		t.Fatalf("Comida após subir caneca esperado [%d %d], tem %v", caneca.ID, pastel.ID, eq)
	}

	// Muda Caneca p/ Bebida -> vai para o fim de Bebida.
	if err := r.UpdateProduto(caneca.ID, "Caneca", 2000, EstoqueLimitado, 3, bebida.ID, 0); err != nil {
		t.Fatal(err)
	}
	grupos, _ = r.ListProdutos(ev.ID)
	comidaG = grupoPorID(grupos, comida.ID)
	bebidaG := grupoPorID(grupos, bebida.ID)
	if eq := idsDe(comidaG.Produtos); !equalInts(eq, []int64{pastel.ID}) {
		t.Fatalf("Comida após mover caneca esperado [%d], tem %v", pastel.ID, eq)
	}
	if eq := idsDe(bebidaG.Produtos); !equalInts(eq, []int64{suco.ID, caneca.ID}) {
		t.Fatalf("Bebida esperado [%d %d], tem %v", suco.ID, caneca.ID, eq)
	}

	// Apaga Comida -> Pastel cai para "Sem grupo" (que já tinha o Chaveiro).
	if err := r.DeleteGrupo(comida.ID); err != nil {
		t.Fatal(err)
	}
	grupos, _ = r.ListProdutos(ev.ID)
	ultimo := grupos[len(grupos)-1]
	if ultimo.ID != SemGrupoID {
		t.Fatalf("último grupo deveria ser 'Sem grupo', é %s", ultimo.Nome)
	}
	if eq := idsDe(ultimo.Produtos); !equalInts(eq, []int64{solto.ID, pastel.ID}) {
		t.Fatalf("Sem grupo esperado [%d %d], tem %v", solto.ID, pastel.ID, eq)
	}
}

func TestReorderFuncoes(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Barraquinha", false)

	comida, _ := r.CreateGrupo(ev.ID, "Comida", "#ef4444")
	bebida, _ := r.CreateGrupo(ev.ID, "Bebida", "#3b82f6")
	sobremesa, _ := r.CreateGrupo(ev.ID, "Sobremesa", "#f59e0b")

	// produtos em Comida, na ordem de criação
	a, _ := r.CreateProduto(ev.ID, "A", 100, EstoqueIlimitado, 0, comida.ID, 0)
	b, _ := r.CreateProduto(ev.ID, "B", 200, EstoqueIlimitado, 0, comida.ID, 0)
	c, _ := r.CreateProduto(ev.ID, "C", 300, EstoqueIlimitado, 0, comida.ID, 0)

	// reordena produtos: C,B,A (ordem final)
	if err := r.ReorderProdutos([]int64{c.ID, b.ID, a.ID}); err != nil {
		t.Fatal(err)
	}
	grupos, _ := r.ListProdutos(ev.ID)
	comidaG := grupoPorID(grupos, comida.ID)
	if eq := idsDe(comidaG.Produtos); !equalInts(eq, []int64{c.ID, b.ID, a.ID}) {
		t.Fatalf("Comida após reorder esperado [%d %d %d], tem %v", c.ID, b.ID, a.ID, eq)
	}

	// reordena grupos: Sobremesa, Comida, Bebida
	if err := r.ReorderGrupos(ev.ID, []int64{sobremesa.ID, comida.ID, bebida.ID}); err != nil {
		t.Fatal(err)
	}
	grupos, _ = r.ListProdutos(ev.ID)
	names := []string{grupos[0].Nome, grupos[1].Nome, grupos[2].Nome}
	want := []string{"Sobremesa", "Comida", "Bebida"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("ordem grupos esperado %v, tem %v", want, names)
		}
	}
}

func TestAtalhoProduto(t *testing.T) {
	r := newTestRepo(t)
	ev, err := r.CreateEvento("PDV Teste", false)
	if err != nil {
		t.Fatal(err)
	}
	ev2, _ := r.CreateEvento("Outro Evento", false)

	// cria produto com tecla 1
	p1, err := r.CreateProduto(ev.ID, "Coxinha", 700, EstoqueIlimitado, 0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p1.Atalho != 1 {
		t.Fatalf("atalho esperado 1, tem %d", p1.Atalho)
	}

	// 0 = sem atalho, guarda NULL; lê de volta como 0
	sem, _ := r.CreateProduto(ev.ID, "Refri", 500, EstoqueIlimitado, 0, 0, 0)
	if sem.Atalho != 0 {
		t.Fatalf("atalho sem tecla deveria ser 0, tem %d", sem.Atalho)
	}

	// mesmo evento, mesma tecla -> erro
	if _, err := r.CreateProduto(ev.ID, "Pastel", 900, EstoqueIlimitado, 0, 0, 1); err == nil {
		t.Fatal("esperava erro ao reusar tecla 1 no mesmo evento")
	}

	// mesmo evento, tecla DIFERENTE -> permitido (regressão: índice errado em
	// produtos(evento_id) barrava qualquer segundo produto com atalho no evento)
	if _, err := r.CreateProduto(ev.ID, "Pastel", 900, EstoqueIlimitado, 0, 0, 2); err != nil {
		t.Fatalf("tecla 2 no mesmo evento deveria passar: %v", err)
	}

	// evento diferente pode reusar a tecla 1
	if _, err := r.CreateProduto(ev2.ID, "Pastel", 900, EstoqueIlimitado, 0, 0, 1); err != nil {
		t.Fatalf("evento diferente não deveria conflitar: %v", err)
	}

	// atualizar p/ uma tecla já usada no mesmo evento -> erro
	if err := r.UpdateProduto(sem.ID, "Refri", 500, EstoqueIlimitado, 0, 0, 1); err == nil {
		t.Fatal("esperava erro ao dar a um produto a tecla 1 já usada")
	}
	// manter a própria tecla 1 ao editar p1 é permitido
	if err := r.UpdateProduto(p1.ID, "Coxinha", 800, EstoqueIlimitado, 0, 0, 1); err != nil {
		t.Fatalf("manter a própria tecla deveria passar: %v", err)
	}
	// range inválido
	if _, err := r.CreateProduto(ev.ID, "X", 100, EstoqueIlimitado, 0, 0, 10); err == nil {
		t.Fatal("esperava erro para atalho fora de 1-9")
	}

	// ListProdutosVenda expõe a tecla
	venda, _ := r.ListProdutosVenda(ev.ID)
	if len(venda) == 0 || venda[0].Produtos[0].Atalho != 1 {
		t.Fatalf("ListProdutosVenda deveria expor atalho 1")
	}
}

func idsDe(ps []Produto) []int64 {
	out := make([]int64, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}

func equalInts(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func grupoPorID(gs []GrupoProdutos, id int64) *GrupoProdutos {
	for i := range gs {
		if gs[i].ID == id {
			return &gs[i]
		}
	}
	return nil
}

// TestCartelaVenda cobre o recurso de cartelas: toggle no evento, catálogo
// (ListCartelas) e venda mista (produto + cartelas) num pedido.
func TestCartelaVenda(t *testing.T) {
	r := newTestRepo(t)

	// Evento com cartelas habilitadas.
	ev, err := r.CreateEvento("Festa Junina", true)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := r.GetEvento(ev.ID)
	if !got.VendeCartela {
		t.Fatal("CreateEvento deveria guardar vende_cartela=true")
	}
	// update desabilita
	if err := r.UpdateEvento(ev.ID, got.Nome, got.Ativo, false); err != nil {
		t.Fatal(err)
	}
	got, _ = r.GetEvento(ev.ID)
	if got.VendeCartela {
		t.Fatal("UpdateEvento deveria desligar vende_cartela")
	}
	if err := r.UpdateEvento(ev.ID, got.Nome, got.Ativo, true); err != nil {
		t.Fatal(err)
	}

	// Catálogo: as 4 denominações com conteúdo não-vazio.
	cartelas, err := cartelaPorReaisList(t)
	if err != nil {
		t.Fatal(err)
	}
	if len(cartelas) != 4 {
		t.Fatalf("esperava 4 cartelas, tem %d", len(cartelas))
	}
	if cartelas[0].Reais != 10 || cartelas[3].Reais != 100 {
		t.Fatalf("ordem das cartelas inesperada: %d..%d", cartelas[0].Reais, cartelas[3].Reais)
	}
	for _, c := range cartelas {
		if c.Preco != c.Reais*100 {
			t.Fatalf("cartela R$%d preco=%d (esperava %d)", c.Reais, c.Preco, c.Reais*100)
		}
		if c.Conteudo == "" || c.Nome == "" {
			t.Fatalf("cartela R$%d sem conteúdo/nome", c.Reais)
		}
	}

	// Produto limitado p/ garantir que a baixa de estoque continua funcionando
	// num pedido misto.
	pastel, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 1)

	pedido, err := r.CriarPedido(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	itens := []PedidoItem{
		{ProdutoID: pastel.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: 2},
		{CartelaReais: 10, Nome: "Cartela R$ 10", PrecoUnit: 1000, Qtd: 3},
		{CartelaReais: 50, Nome: "Cartela R$ 50", PrecoUnit: 5000, Qtd: 1},
	}
	fechado, err := r.FecharPedido(pedido.ID, itens, FormaDinheiro, "")
	if err != nil {
		t.Fatal(err)
	}
	if fechado.Total != 2*1500+3*1000+1*5000 {
		t.Fatalf("total esperado 11000, tem %d", fechado.Total)
	}
	if len(fechado.Itens) != 3 {
		t.Fatalf("esperava 3 itens, tem %d", len(fechado.Itens))
	}
	// as cartelas voltam identificadas pelo CartelaReais e sem produto.
	achouCartela := false
	for _, it := range fechado.Itens {
		if it.CartelaReais == 10 {
			achouCartela = true
			if it.ProdutoID != 0 || it.Qtd != 3 {
				t.Fatalf("cartela R$10 esperada qtd 3 sem produto; veio id=%d qtd=%d", it.ProdutoID, it.Qtd)
			}
		}
	}
	if !achouCartela {
		t.Fatal("cartela R$10 não retornada no pedido fechado")
	}

	// Denominação desconhecida é rejeitada.
	p2, _ := r.CriarPedido(ev.ID)
	ruim := []PedidoItem{{CartelaReais: 7, Nome: "Cartela R$ 7", PrecoUnit: 700, Qtd: 1}}
	if _, err := r.FecharPedido(p2.ID, ruim, FormaDinheiro, ""); err == nil {
		t.Fatal("esperava erro para cartela de valor desconhecido")
	}
}

func cartelaPorReaisList(t *testing.T) ([]Cartela, error) {
	t.Helper()
	out := []Cartela{}
	for _, reais := range CartelaReaisValidos {
		c, err := cartelaPorReais(reais)
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Anota aí (fiado): uma venda cria a conta; outra do mesmo dono (mesmo evento)
// reusa a mesma conta; eventos diferentes têm contas separadas.
func TestAnotaAi(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Festa 2026", false)
	ev2, _ := r.CreateEvento("Retiro 2026", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 0)

	// Nova conta "Maria" (cria) com forma anotaai.
	itens := []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: 1}}
	p1, _ := r.CriarPedido(ev.ID)
	f1, err := r.FecharPedido(p1.ID, itens, FormaAnotaAi, "Maria")
	if err != nil {
		t.Fatal(err)
	}
	if f1.Forma != FormaAnotaAi || f1.ContaID == 0 {
		t.Fatalf("esperava forma anotaai com conta, veio forma=%q conta=%d", f1.Forma, f1.ContaID)
	}

	// Segundo pedido do mesmo dono reusa a MESMA conta (não duplica).
	p2, _ := r.CriarPedido(ev.ID)
	f2, err := r.FecharPedido(p2.ID, itens, FormaAnotaAi, "maria") // case-insensitive
	if err != nil {
		t.Fatal(err)
	}
	if f2.ContaID != f1.ContaID {
		t.Fatalf("conta deveria ser reutilizada (mesmo evento/nome): %d vs %d", f2.ContaID, f1.ContaID)
	}

	// Nome diferente vira conta distinta.
	p3, _ := r.CriarPedido(ev.ID)
	f3, _ := r.FecharPedido(p3.ID, itens, FormaAnotaAi, "Maria Souza")
	if f3.ContaID == f1.ContaID {
		t.Fatal("conta 'Maria Souza' não deveria ser a mesma de 'Maria'")
	}

	// Mesmo nome em OUTRO evento cria outra conta (contas são por evento).
	prod2, _ := r.CreateProduto(ev2.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 0)
	p4, _ := r.CriarPedido(ev2.ID)
	f4, _ := r.FecharPedido(p4.ID, []PedidoItem{{ProdutoID: prod2.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: 1}}, FormaAnotaAi, "Maria")
	if f4.ContaID == f1.ContaID {
		t.Fatal("conta em outro evento não deveria reusar a do evento 1")
	}

	// Só o dono do evento vê suas contas (autocomplete).
	contas, _ := r.ListContas(ev.ID)
	if len(contas) != 2 { // Maria + Maria Souza
		t.Fatalf("evento 1 esperava 2 contas, tem %d", len(contas))
	}
	// TotalPendente de cada conta = soma dos totais dos pedidos anotaai.
	for _, c := range contas {
		switch c.Nome {
		case "Maria":
			if c.TotalPendente != 1500+1500 { // f1 + f2 (Maria Souza e dinheiro não entram)
				t.Fatalf("Maria deveria ter pendência 3000, tem %d", c.TotalPendente)
			}
		case "Maria Souza":
			if c.TotalPendente != 1500 {
				t.Fatalf("Maria Souza deveria ter pendência 1500, tem %d", c.TotalPendente)
			}
		}
	}

	// Dinheiro/cartão não geram conta e rejeitam nome vazio no anotaai.
	p5, _ := r.CriarPedido(ev.ID)
	f5, err := r.FecharPedido(p5.ID, itens, FormaDinheiro, "")
	if err != nil {
		t.Fatal(err)
	}
	if f5.ContaID != 0 || f5.Forma != FormaDinheiro {
		t.Fatalf("dinheiro não deveria ter conta (conta=%d forma=%q)", f5.ContaID, f5.Forma)
	}
	p6, _ := r.CriarPedido(ev.ID)
	if _, err := r.FecharPedido(p6.ID, itens, FormaAnotaAi, "   "); err == nil {
		t.Fatal("esperava erro ao anotar sem nome do dono")
	}
	// forma desconhecida é rejeitada
	p7, _ := r.CriarPedido(ev.ID)
	if _, err := r.FecharPedido(p7.ID, itens, "pix", ""); err == nil {
		t.Fatal("esperava erro para forma de pagamento desconhecida")
	}
}

// TestQuitarContaEresumo cobre a quitação de uma conta (vendas anotadas em
// aberto -> pagas) e o relatório do Dashboard (ResumoEvento). Só vendas
// fechadas com total > 0 entram no resumo.
func TestQuitarContaEresumo(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Bazar 2026", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1000, EstoqueIlimitado, 0, 0, 0)
	it1 := []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1000, Qtd: 1}}

	// 2 vendas anotadas da Maria + 1 venda em dinheiro (qtd 2) = receita 4000.
	p1, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p1.ID, it1, FormaAnotaAi, "Maria")
	p2, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p2.ID, it1, FormaAnotaAi, "Maria")
	p3, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p3.ID, []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1000, Qtd: 2}}, FormaDinheiro, "")

	// Um pedido aberto (sem FecharPedido) não deve contar no resumo.
	r.CriarPedido(ev.ID)

	// Resumo do evento.
	res, err := r.ResumoEvento(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.ReceitaTotal != 1000+1000+2000 {
		t.Fatalf("receita esperada 4000, tem %d", res.ReceitaTotal)
	}
	if res.NumPedidos != 3 {
		t.Fatalf("esperava 3 pedidos fechados, tem %d", res.NumPedidos)
	}
	if res.NumProdutosVendidos != 1+1+2 {
		t.Fatalf("esperava 4 unidades de produto, tem %d", res.NumProdutosVendidos)
	}
	if len(res.VendasPorHora) != 24 {
		t.Fatalf("histograma esperava 24 horas, tem %d", len(res.VendasPorHora))
	}
	var totalHora int64
	for _, h := range res.VendasPorHora {
		totalHora += h.Vendas
	}
	if totalHora != 3 {
		t.Fatalf("histograma deveria somar 3 vendas, somou %d", totalHora)
	}

	// Autocomplete do PDV: só conta vendas em aberto da conta.
	contas, _ := r.ListContas(ev.ID)
	if len(contas) != 1 || contas[0].Nome != "Maria" {
		t.Fatalf("esperava só a conta Maria, tem %d", len(contas))
	}
	if contas[0].TotalPendente != 2000 {
		t.Fatalf("Maria deveria dever 2000 (aberto), tem %d", contas[0].TotalPendente)
	}
	maria := contas[0].ID

	// Saldo na tela "Anota aí": pendente 2000, 2 abertas, nada quitado.
	saldo, _ := r.ListContasSaldo(ev.ID)
	if len(saldo) != 1 {
		t.Fatalf("esperava 1 conta no saldo, tem %d", len(saldo))
	}
	s := saldo[0]
	if s.TotalPendente != 2000 || s.NumAberto != 2 || s.TotalQuitado != 0 {
		t.Fatalf("saldo inicial inesperado: pendente=%d aberto=%d quitado=%d", s.TotalPendente, s.NumAberto, s.TotalQuitado)
	}

	// Quita a conta: aberto zera, quitado vira 2000, autocomplete some o valor.
	if err := r.QuitarConta(maria); err != nil {
		t.Fatal(err)
	}
	saldo, _ = r.ListContasSaldo(ev.ID)
	s = saldo[0]
	if s.TotalPendente != 0 || s.NumAberto != 0 || s.TotalQuitado != 2000 {
		t.Fatalf("saldo pós-quitação inesperado: pendente=%d aberto=%d quitado=%d", s.TotalPendente, s.NumAberto, s.TotalQuitado)
	}
	contas, _ = r.ListContas(ev.ID)
	if contas[0].TotalPendente != 0 {
		t.Fatalf("autocomplete deveria mostrar pendência 0 pós-quitação, tem %d", contas[0].TotalPendente)
	}

	// Nova venda anotada reabre a conta como pendente (sem perder o quitado).
	p4, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p4.ID, it1, FormaAnotaAi, "Maria")
	saldo, _ = r.ListContasSaldo(ev.ID)
	s = saldo[0]
	if s.TotalPendente != 1000 || s.NumAberto != 1 || s.TotalQuitado != 2000 {
		t.Fatalf("saldo após nova anotação inesperado: pendente=%d aberto=%d quitado=%d", s.TotalPendente, s.NumAberto, s.TotalQuitado)
	}
}

// Conta em outro evento não é quitada por engano: QuitarConta só age nos
// pedidos da conta alvo (que já é por evento), então cria contas de eventos
// diferentes e quita só uma delas.
func TestQuitarContaIsoladaPorEvento(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("A", false)
	ev2, _ := r.CreateEvento("B", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1000, EstoqueIlimitado, 0, 0, 0)
	prod2, _ := r.CreateProduto(ev2.ID, "Pastel", 1000, EstoqueIlimitado, 0, 0, 0)
	it := func(pid int64) []PedidoItem {
		return []PedidoItem{{ProdutoID: pid, Nome: "Pastel", PrecoUnit: 1000, Qtd: 1}}
	}

	p1, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p1.ID, it(prod.ID), FormaAnotaAi, "Maria")
	p2, _ := r.CriarPedido(ev2.ID)
	r.FecharPedido(p2.ID, it(prod2.ID), FormaAnotaAi, "Maria")

	contasEv1, _ := r.ListContas(ev.ID)
	if len(contasEv1) != 1 {
		t.Fatal("evento A deveria ter a conta Maria")
	}
	if err := r.QuitarConta(contasEv1[0].ID); err != nil {
		t.Fatal(err)
	}
	// Evento A quitado; evento B segue pendente.
	s1, _ := r.ListContasSaldo(ev.ID)
	if s1[0].TotalPendente != 0 {
		t.Fatalf("evento A deveria estar quitado, pendente=%d", s1[0].TotalPendente)
	}
	s2, _ := r.ListContasSaldo(ev2.ID)
	if s2[0].TotalPendente != 1000 {
		t.Fatalf("evento B deveria seguir pendente 1000, tem %d", s2[0].TotalPendente)
	}
}

// TestListPedidosConta garante que ListPedidosConta retorna os pedidos anotaai
// da conta com os itens de cada um (mais recentes primeiro) e sem deadlock — o
// banco usa MaxOpenConns(1), então não se pode abrir a query de itens enquanto
// o rows do SELECT pedidos ainda está aberto.
func TestListPedidosConta(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Festa 2026", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 0)
	it := func(qtd int64) []PedidoItem {
		return []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: qtd}}
	}

	// 1º pedido (1 pastel) e depois 2º (2 pastéis) para a mesma conta.
	p1, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p1.ID, it(1), FormaAnotaAi, "Maria")
	p2, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p2.ID, it(2), FormaAnotaAi, "Maria")

	contas, _ := r.ListContas(ev.ID)
	if len(contas) != 1 {
		t.Fatal("esperava 1 conta")
	}
	lista, err := r.ListPedidosConta(contas[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) != 2 {
		t.Fatalf("esperava 2 pedidos na conta, tem %d", len(lista))
	}
	// Mais recente primeiro: o pedido 2 (2 pastéis, total 3000) deve vir antes.
	if lista[0].Numero != 2 || lista[0].Total != 3000 {
		t.Fatalf("pedido mais recente deveria ser o #2 total 3000, veio #%d total %d", lista[0].Numero, lista[0].Total)
	}
	if len(lista[0].Itens) != 1 || lista[0].Itens[0].Qtd != 2 {
		t.Fatalf("pedido #2 deveria ter 1 item com qtd 2")
	}
	if lista[1].Numero != 1 || len(lista[1].Itens) != 1 {
		t.Fatalf("pedido #1 deveria vir em segundo com 1 item")
	}
	for _, p := range lista {
		if p.QuitadoEm != "" {
			t.Fatalf("pedido #%d não deveria estar quitado antes da quitação", p.Numero)
		}
	}

	// Após quitar, ambos ficam marcados como pagos.
	if err := r.QuitarConta(contas[0].ID); err != nil {
		t.Fatal(err)
	}
	lista, _ = r.ListPedidosConta(contas[0].ID)
	for _, p := range lista {
		if p.QuitadoEm == "" {
			t.Fatalf("pedido #%d deveria estar quitado após QuitarConta", p.Numero)
		}
	}
}

// TestListPedidos cobre a listagem paginada de pedidos de um evento: do mais
// recente para o mais antigo, com o nome da conta do "Anota aí" quando houver,
// filtro por número, e ignorando pedidos ainda em aberto (total 0).
func TestListPedidos(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Festa 2026", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 0)
	it := func(qtd int64) []PedidoItem {
		return []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: qtd}}
	}

	// 3 vendas fechadas: dinheiro (#1), anotaai Maria (#2) e dinheiro (#3).
	p1, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p1.ID, it(1), FormaDinheiro, "")
	p2, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p2.ID, it(1), FormaAnotaAi, "Maria")
	p3, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p3.ID, it(2), FormaDinheiro, "")
	// Um pedido aberto (nunca fechado) NÃO deve entrar na listagem.
	r.CriarPedido(ev.ID)

	// Página com tudo: total 3, mais recente primeiro (#3, #2, #1).
	all, err := r.ListPedidos(ev.ID, 1, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if all.Total != 3 || len(all.Pedidos) != 3 {
		t.Fatalf("esperava 3 pedidos no total e na página, tem total=%d len=%d", all.Total, len(all.Pedidos))
	}
	wantNum := []int64{3, 2, 1}
	for i, pr := range all.Pedidos {
		if pr.Numero != wantNum[i] {
			t.Fatalf("posição %d deveria ser o pedido #%d, veio #%d", i, wantNum[i], pr.Numero)
		}
	}
	// O pedido anotaai (#2) traz o nome da conta; os em dinheiro, não.
	if all.Pedidos[1].Forma != FormaAnotaAi || all.Pedidos[1].ContaNome != "Maria" || all.Pedidos[1].ContaID == 0 {
		t.Fatalf("pedido anotaai deveria trazer conta Maria, veio forma=%q nome=%q conta=%d",
			all.Pedidos[1].Forma, all.Pedidos[1].ContaNome, all.Pedidos[1].ContaID)
	}
	if all.Pedidos[0].ContaNome != "" || all.Pedidos[0].ContaID != 0 {
		t.Fatalf("pedido em dinheiro não deveria ter conta")
	}

	// Paginação: página de 2 -> página 1 tem os 2 mais recentes, página 2 o 3º.
	pg1, _ := r.ListPedidos(ev.ID, 1, 2, 0)
	if len(pg1.Pedidos) != 2 || pg1.TotalPaginas != 2 {
		t.Fatalf("página 1 esperava 2 pedidos e 2 páginas, tem len=%d paginas=%d", len(pg1.Pedidos), pg1.TotalPaginas)
	}
	if pg1.Pedidos[0].Numero != 3 || pg1.Pedidos[1].Numero != 2 {
		t.Fatalf("página 1 deveria ser #3 e #2, veio #%d e #%d", pg1.Pedidos[0].Numero, pg1.Pedidos[1].Numero)
	}
	pg2, _ := r.ListPedidos(ev.ID, 2, 2, 0)
	if len(pg2.Pedidos) != 1 || pg2.Pedidos[0].Numero != 1 || pg2.Pagina != 2 {
		t.Fatalf("página 2 deveria ter só o #1, veio len=%d primeiro=%d pagina=%d",
			len(pg2.Pedidos), pg2.Pedidos[0].Numero, pg2.Pagina)
	}

	// Busca por número: só o #2 volta, com total 1 (independente do tamanho da página).
	busca, _ := r.ListPedidos(ev.ID, 1, 50, 2)
	if busca.Total != 1 || len(busca.Pedidos) != 1 || busca.Pedidos[0].Numero != 2 {
		t.Fatalf("busca pelo #2 deveria trazer só ele, tem total=%d len=%d", busca.Total, len(busca.Pedidos))
	}
	// Número que não existe: lista vazia, total 0.
	nenhum, _ := r.ListPedidos(ev.ID, 1, 50, 99)
	if nenhum.Total != 0 || len(nenhum.Pedidos) != 0 {
		t.Fatalf("busca por número inexistente deveria vir vazia, tem total=%d", nenhum.Total)
	}
}

// TestCancelarPedido garante que cancelar um pedido devolve o estoque dos
// produtos limitados, tira o pedido do resumo/dashboard, marca cancelado_em e
// rejeita cancelar duas vezes ou cancelar pedido ainda em aberto.
func TestCancelarPedido(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Bazar 2026", false)
	// Ilimitado (não tem estoque p/ devolver) + limitado com 5 unidades.
	ilim, _ := r.CreateProduto(ev.ID, "Pastel", 1000, EstoqueIlimitado, 0, 0, 0)
	lim, _ := r.CreateProduto(ev.ID, "Caneca", 2000, EstoqueLimitado, 5, 0, 0)

	p1, _ := r.CriarPedido(ev.ID)
	f1, err := r.FecharPedido(p1.ID, []PedidoItem{
		{ProdutoID: ilim.ID, Nome: "Pastel", PrecoUnit: 1000, Qtd: 2},
		{ProdutoID: lim.ID, Nome: "Caneca", PrecoUnit: 2000, Qtd: 3},
	}, FormaDinheiro, "")
	if err != nil {
		t.Fatal(err)
	}
	if f1.Total != 2*1000+3*2000 {
		t.Fatalf("total esperado 8000, tem %d", f1.Total)
	}
	limDepois, _ := r.GetProduto(lim.ID)
	if limDepois.Quantidade != 2 {
		t.Fatalf("estoque da caneca deveria baixar p/ 2, tem %d", limDepois.Quantidade)
	}

	// Resumo antes: receita 8000, 1 pedido, 5 unidades.
	res0, _ := r.ResumoEvento(ev.ID)
	if res0.ReceitaTotal != 8000 || res0.NumPedidos != 1 || res0.NumProdutosVendidos != 5 {
		t.Fatalf("resumo inicial inesperado: receita=%d pedidos=%d unidades=%d",
			res0.ReceitaTotal, res0.NumPedidos, res0.NumProdutosVendidos)
	}

	// Cancela: estoque devolvido, pedido fora do resumo, marcado cancelado.
	if err := r.CancelarPedido(p1.ID); err != nil {
		t.Fatal(err)
	}
	limPos, _ := r.GetProduto(lim.ID)
	if limPos.Quantidade != 5 {
		t.Fatalf("estoque da caneca deveria voltar a 5, tem %d", limPos.Quantidade)
	}
	ilimPos, _ := r.GetProduto(ilim.ID)
	if ilimPos.Quantidade != 0 {
		t.Fatalf("produto ilimitado não deveria mudar (estoque %d)", ilimPos.Quantidade)
	}
	res1, _ := r.ResumoEvento(ev.ID)
	if res1.ReceitaTotal != 0 || res1.NumPedidos != 0 || res1.NumProdutosVendidos != 0 {
		t.Fatalf("resumo pós-cancelamento deveria zerar, veio receita=%d pedidos=%d unidades=%d",
			res1.ReceitaTotal, res1.NumPedidos, res1.NumProdutosVendidos)
	}
	ped, _ := r.GetPedido(p1.ID)
	if ped.CanceladoEm == "" {
		t.Fatal("pedido deveria estar marcado como cancelado")
	}

	// O cancelado segue na listagem (histórico), com cancelado_em preenchido.
	lista, _ := r.ListPedidos(ev.ID, 1, 50, 0)
	if lista.Total != 1 || lista.Pedidos[0].CanceladoEm == "" {
		t.Fatalf("cancelado deveria seguir listado como histórico, total=%d", lista.Total)
	}

	// Cancelar duas vezes e cancelar pedido em aberto (total 0) dão erro.
	if err := r.CancelarPedido(p1.ID); err == nil {
		t.Fatal("esperava erro ao cancelar pedido já cancelado")
	}
	aberto, _ := r.CriarPedido(ev.ID)
	if err := r.CancelarPedido(aberto.ID); err == nil {
		t.Fatal("esperava erro ao cancelar pedido ainda em aberto")
	}
	// Cancelar pedido inexistente também é erro.
	if err := r.CancelarPedido(999999); err == nil {
		t.Fatal("esperava erro ao cancelar pedido inexistente")
	}
}

// TestCancelarPedidoAnotaAi garante que cancelar uma venda "Anota aí" remove o
// débito da conta (quando em aberto) e deixa de contar no que já foi quitado.
func TestCancelarPedidoAnotaAi(t *testing.T) {
	r := newTestRepo(t)
	ev, _ := r.CreateEvento("Retiro 2026", false)
	prod, _ := r.CreateProduto(ev.ID, "Pastel", 1500, EstoqueIlimitado, 0, 0, 0)
	it := []PedidoItem{{ProdutoID: prod.ID, Nome: "Pastel", PrecoUnit: 1500, Qtd: 1}}

	// Uma venda anotada da Maria, em aberto.
	p1, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p1.ID, it, FormaAnotaAi, "Maria")
	contas, _ := r.ListContas(ev.ID)
	if len(contas) != 1 || contas[0].TotalPendente != 1500 {
		t.Fatalf("Maria deveria dever 1500, tem %d contas", len(contas))
	}
	maria := contas[0].ID

	// Cancela -> o débito some da conta (pendência 0) e ela sai dos pedidos da conta.
	if err := r.CancelarPedido(p1.ID); err != nil {
		t.Fatal(err)
	}
	contas, _ = r.ListContas(ev.ID)
	if len(contas) != 1 || contas[0].TotalPendente != 0 {
		t.Fatalf("débito da Maria deveria ter sido removido (pendente=%d)", contas[0].TotalPendente)
	}
	saldo, _ := r.ListContasSaldo(ev.ID)
	if len(saldo) != 1 || saldo[0].TotalPendente != 0 || saldo[0].NumAberto != 0 {
		t.Fatalf("saldo da Maria pós-cancelamento inesperado: pendente=%d aberto=%d",
			saldo[0].TotalPendente, saldo[0].NumAberto)
	}
	pedidosConta, _ := r.ListPedidosConta(maria)
	if len(pedidosConta) != 0 {
		t.Fatalf("pedido cancelado não deveria aparecer na conta, tem %d", len(pedidosConta))
	}

	// Uma venda anotada quitada e depois cancelada deixa de contar no quitado.
	p2, _ := r.CriarPedido(ev.ID)
	r.FecharPedido(p2.ID, it, FormaAnotaAi, "Maria")
	if err := r.QuitarConta(maria); err != nil {
		t.Fatal(err)
	}
	saldo, _ = r.ListContasSaldo(ev.ID)
	if saldo[0].TotalQuitado != 1500 {
		t.Fatalf("após quitar, quitado deveria ser 1500, tem %d", saldo[0].TotalQuitado)
	}
	if err := r.CancelarPedido(p2.ID); err != nil {
		t.Fatal(err)
	}
	saldo, _ = r.ListContasSaldo(ev.ID)
	if saldo[0].TotalQuitado != 0 {
		t.Fatalf("pedido quitado cancelado não deveria contar no quitado, tem %d", saldo[0].TotalQuitado)
	}
}
