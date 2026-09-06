package main

import "fmt"

// =============================================================
// Camada de bindings Wails — delega para *repo. Garante a inicialização
// de slices vazios (JSON [] ao invés de null).
// =============================================================

func (a *App) repoOrErr() (*repo, error) {
	if a.repo == nil || a.repo.db == nil {
		return nil, fmt.Errorf("banco de dados indisponível")
	}
	return a.repo, nil
}

// ---- Eventos ----
func (a *App) ListEventos() ([]Evento, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []Evento{}, err
	}
	return r.ListEventos()
}

func (a *App) CreateEvento(nome string, vendeCartela bool) (Evento, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Evento{}, err
	}
	return r.CreateEvento(nome, vendeCartela)
}

func (a *App) UpdateEvento(id int64, nome string, ativo bool, vendeCartela bool) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.UpdateEvento(id, nome, ativo, vendeCartela)
}

func (a *App) DeleteEvento(id int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.DeleteEvento(id)
}

// ClonarEvento duplica o evento (com grupos e produtos), sem pedidos/contas.
func (a *App) ClonarEvento(id int64) (Evento, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Evento{}, err
	}
	return r.ClonarEvento(id)
}

// ---- Cartelas ----
// ListCartelas devolve o catálogo fixo de cartelas (R$ 10/20/50/100) com o
// layout ASCII de cada uma (lido dos arquivos embutidos cartelas/*.txt).
func (a *App) ListCartelas() ([]Cartela, error) {
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

// ---- Grupos ----
func (a *App) ListGrupos(eventoID int64) ([]Grupo, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []Grupo{}, err
	}
	return r.gruposOrdenados(eventoID)
}

func (a *App) CreateGrupo(eventoID int64, nome, cor string) (Grupo, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Grupo{}, err
	}
	return r.CreateGrupo(eventoID, nome, cor)
}

func (a *App) UpdateGrupo(id int64, nome, cor string) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.UpdateGrupo(id, nome, cor)
}

func (a *App) DeleteGrupo(id int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.DeleteGrupo(id)
}

func (a *App) MoverGrupo(id, delta int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.MoverGrupo(id, delta)
}

func (a *App) ReorderGrupos(eventoID int64, ids []int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.ReorderGrupos(eventoID, ids)
}

// ---- Produtos ----
func (a *App) ListProdutos(eventoID int64) ([]GrupoProdutos, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []GrupoProdutos{}, err
	}
	return r.ListProdutos(eventoID)
}

func (a *App) ListProdutosVenda(eventoID int64) ([]GrupoVenda, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []GrupoVenda{}, err
	}
	return r.ListProdutosVenda(eventoID)
}

func (a *App) CreateProduto(eventoID int64, nome string, preco int64, estoqueTipo string, quantidade, grupoID, atalho int64) (Produto, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Produto{}, err
	}
	return r.CreateProduto(eventoID, nome, preco, estoqueTipo, quantidade, grupoID, atalho)
}

func (a *App) UpdateProduto(id int64, nome string, preco int64, estoqueTipo string, quantidade, grupoID, atalho int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.UpdateProduto(id, nome, preco, estoqueTipo, quantidade, grupoID, atalho)
}

func (a *App) SetProdutoAtivo(id int64, ativo bool) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.SetProdutoAtivo(id, ativo)
}

func (a *App) DeleteProduto(id int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.DeleteProduto(id)
}

func (a *App) MoverProduto(id, delta int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.MoverProduto(id, delta)
}

func (a *App) ReorderProdutos(ids []int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.ReorderProdutos(ids)
}

// ---- Pedidos ----
func (a *App) CriarPedido(eventoID int64) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.CriarPedido(eventoID)
}

// ListPedidos devolve uma página dos pedidos fechados de um evento (mais
// recentes primeiro), filtrando pelo número quando numeroBusca > 0.
func (a *App) ListPedidos(eventoID, pagina, porPagina, numeroBusca int64) (ListaPedidos, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return ListaPedidos{}, err
	}
	return r.ListPedidos(eventoID, pagina, porPagina, numeroBusca)
}

// CancelarPedido cancela um pedido fechado: devolve o estoque dos produtos
// limitados e remove o débito do "Anota aí" se a venda foi anotada em aberto.
func (a *App) CancelarPedido(pedidoID int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.CancelarPedido(pedidoID)
}

// ---- Contas ("Anota aí") ----
// ListContas devolve as contas de um evento (autocomplete do PDV).
func (a *App) ListContas(eventoID int64) ([]Conta, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []Conta{}, err
	}
	return r.ListContas(eventoID)
}

func (a *App) GetPedido(id int64) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.GetPedido(id)
}

func (a *App) FecharPedido(pedidoID int64, itens []PedidoItem, forma, contaNome string) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.FecharPedido(pedidoID, itens, forma, contaNome)
}

// ---- Dashboard / quitação do "Anota aí" ----

// ListContasSaldo devolve as contas de um evento com o saldo (pendente, nº em
// aberto, quitado) para a tela de gerenciamento do "Anota aí".
func (a *App) ListContasSaldo(eventoID int64) ([]ContaSaldo, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []ContaSaldo{}, err
	}
	return r.ListContasSaldo(eventoID)
}

// QuitarConta marca como pagas as vendas anotadas em aberto de uma conta.
func (a *App) QuitarConta(contaID int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.QuitarConta(contaID)
}

// ListPedidosConta devolve os pedidos "Anota aí" de uma conta (em aberto e já
// quitados), com os itens de cada um.
func (a *App) ListPedidosConta(contaID int64) ([]Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []Pedido{}, err
	}
	return r.ListPedidosConta(contaID)
}

// ResumoEvento agrega as vendas fechadas de um evento para o Dashboard.
func (a *App) ResumoEvento(eventoID int64) (ResumoEvento, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return ResumoEvento{}, err
	}
	return r.ResumoEvento(eventoID)
}
