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

func (a *App) CreateEvento(nome string) (Evento, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Evento{}, err
	}
	return r.CreateEvento(nome)
}

func (a *App) UpdateEvento(id int64, nome string, ativo bool) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.UpdateEvento(id, nome, ativo)
}

func (a *App) DeleteEvento(id int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.DeleteEvento(id)
}

// ---- Produtos ----
func (a *App) ListProdutos(eventoID int64) ([]Produto, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []Produto{}, err
	}
	return r.ListProdutos(eventoID)
}

func (a *App) ListProdutosVenda(eventoID int64) ([]ProdutoVenda, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return []ProdutoVenda{}, err
	}
	return r.ListProdutosVenda(eventoID)
}

func (a *App) CreateProduto(eventoID int64, nome string, preco int64, estoqueTipo string, quantidade int64) (Produto, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Produto{}, err
	}
	return r.CreateProduto(eventoID, nome, preco, estoqueTipo, quantidade)
}

func (a *App) UpdateProduto(id int64, nome string, preco int64, estoqueTipo string, quantidade int64) error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return r.UpdateProduto(id, nome, preco, estoqueTipo, quantidade)
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

// ---- Pedidos ----
func (a *App) CriarPedido(eventoID int64) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.CriarPedido(eventoID)
}

func (a *App) GetPedido(id int64) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.GetPedido(id)
}

func (a *App) FecharPedido(pedidoID int64, itens []PedidoItem) (Pedido, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return Pedido{}, err
	}
	return r.FecharPedido(pedidoID, itens)
}
