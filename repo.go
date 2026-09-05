package main

import (
	"database/sql"
	"errors"
	"fmt"
)

// =============================================================
// Eventos
// =============================================================

func (r *repo) ListEventos() ([]Evento, error) {
	out := []Evento{}
	rows, err := r.db.Query(`SELECT id, nome, ativo, criado_em FROM eventos ORDER BY criado_em DESC, id DESC`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var e Evento
		var ativo int
		if err := rows.Scan(&e.ID, &e.Nome, &ativo, &e.CriadoEm); err != nil {
			return out, err
		}
		e.Ativo = ativo == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *repo) GetEvento(id int64) (Evento, error) {
	var e Evento
	var ativo int
	err := r.db.QueryRow(`SELECT id, nome, ativo, criado_em FROM eventos WHERE id = ?`, id).
		Scan(&e.ID, &e.Nome, &ativo, &e.CriadoEm)
	if err != nil {
		return e, err
	}
	e.Ativo = ativo == 1
	return e, nil
}

func (r *repo) CreateEvento(nome string) (Evento, error) {
	var e Evento
	res, err := r.db.Exec(`INSERT INTO eventos (nome) VALUES (?)`, nome)
	if err != nil {
		return e, err
	}
	id, _ := res.LastInsertId()
	return r.GetEvento(id)
}

func (r *repo) UpdateEvento(id int64, nome string, ativo bool) error {
	a := 0
	if ativo {
		a = 1
	}
	res, err := r.db.Exec(`UPDATE eventos SET nome = ?, ativo = ? WHERE id = ?`, nome, a, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "evento")
}

func (r *repo) DeleteEvento(id int64) error {
	res, err := r.db.Exec(`DELETE FROM eventos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "evento")
}

// =============================================================
// Produtos
// =============================================================

func (r *repo) ListProdutos(eventoID int64) ([]Produto, error) {
	out := []Produto{}
	rows, err := r.db.Query(`SELECT id, evento_id, nome, preco, estoque_tipo, quantidade, ativo
		FROM produtos WHERE evento_id = ? ORDER BY nome COLLATE NOCASE`, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return out, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListProdutosVenda retorna produtos ativos de um evento na visão do PDV,
// já ordenados por nome e com flag esgotado computado.
func (r *repo) ListProdutosVenda(eventoID int64) ([]ProdutoVenda, error) {
	out := []ProdutoVenda{}
	rows, err := r.db.Query(`SELECT id, nome, preco, estoque_tipo, quantidade, ativo
		FROM produtos WHERE evento_id = ? ORDER BY nome COLLATE NOCASE`, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var pv ProdutoVenda
		var tipo string
		var qtd, ativo int64
		if err := rows.Scan(&pv.ID, &pv.Nome, &pv.Preco, &tipo, &qtd, &ativo); err != nil {
			return out, err
		}
		pv.EstoqueTipo = tipo
		pv.Quantidade = qtd
		if ativo == 0 {
			continue // não oferece produtos inativos na venda
		}
		pv.Esgotado = tipo == EstoqueLimitado && qtd <= 0
		out = append(out, pv)
	}
	return out, rows.Err()
}

func (r *repo) GetProduto(id int64) (Produto, error) {
	row := r.db.QueryRow(`SELECT id, evento_id, nome, preco, estoque_tipo, quantidade, ativo
		FROM produtos WHERE id = ?`, id)
	return scanProduto(row)
}

func (r *repo) CreateProduto(eventoID int64, nome string, preco int64, tipo string, quantidade int64) (Produto, error) {
	var p Produto
	if preco < 0 {
		return p, errors.New("preço não pode ser negativo")
	}
	if tipo != EstoqueIlimitado && tipo != EstoqueLimitado {
		return p, errors.New("tipo de estoque inválido")
	}
	if tipo == EstoqueLimitado && quantidade < 0 {
		return p, errors.New("quantidade não pode ser negativa")
	}
	if tipo == EstoqueIlimitado {
		quantidade = 0
	}
	res, err := r.db.Exec(`INSERT INTO produtos (evento_id, nome, preco, estoque_tipo, quantidade) VALUES (?,?,?,?,?)`,
		eventoID, nome, preco, tipo, quantidade)
	if err != nil {
		return p, err
	}
	id, _ := res.LastInsertId()
	return r.GetProduto(id)
}

func (r *repo) UpdateProduto(id int64, nome string, preco int64, tipo string, quantidade int64) error {
	if preco < 0 {
		return errors.New("preço não pode ser negativo")
	}
	if tipo != EstoqueIlimitado && tipo != EstoqueLimitado {
		return errors.New("tipo de estoque inválido")
	}
	if tipo == EstoqueIlimitado {
		quantidade = 0
	} else if quantidade < 0 {
		return errors.New("quantidade não pode ser negativa")
	}
	res, err := r.db.Exec(`UPDATE produtos SET nome = ?, preco = ?, estoque_tipo = ?, quantidade = ? WHERE id = ?`,
		nome, preco, tipo, quantidade, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "produto")
}

func (r *repo) SetProdutoAtivo(id int64, ativo bool) error {
	a := 0
	if ativo {
		a = 1
	}
	res, err := r.db.Exec(`UPDATE produtos SET ativo = ? WHERE id = ?`, a, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "produto")
}

func (r *repo) DeleteProduto(id int64) error {
	res, err := r.db.Exec(`DELETE FROM produtos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "produto")
}

// BaixaEstoque subtrai qtd de um produto limitado, garantindo que não
// fique negativo (transação única já vem da chamada FecharPedido).
func (r *repo) baixaEstoque(produtoID, qtd int64) error {
	res, err := r.db.Exec(`UPDATE produtos SET quantidade = quantidade - ? WHERE id = ? AND estoque_tipo = 'limitado' AND quantidade >= ?`,
		qtd, produtoID, qtd)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("estoque insuficiente")
	}
	return nil
}

// =============================================================
// Pedidos
// =============================================================

// CriarPedido abre um novo pedido para o evento com o próximo número
// sequencial. Retorna o pedido vazio (sem itens) já com ID/numero.
func (r *repo) CriarPedido(eventoID int64) (Pedido, error) {
	var p Pedido
	tx, err := r.db.Begin()
	if err != nil {
		return p, err
	}
	defer tx.Rollback()

	var prox int64
	err = tx.QueryRow(`SELECT COALESCE(MAX(numero),0)+1 FROM pedidos WHERE evento_id = ?`, eventoID).Scan(&prox)
	if err != nil {
		return p, err
	}
	res, err := tx.Exec(`INSERT INTO pedidos (evento_id, numero) VALUES (?,?)`, eventoID, prox)
	if err != nil {
		return p, err
	}
	id, _ := res.LastInsertId()
	if err := tx.Commit(); err != nil {
		return p, err
	}
	p.ID = id
	p.EventoID = eventoID
	p.Numero = prox
	p.Itens = []PedidoItem{}
	return p, nil
}

func (r *repo) GetPedido(id int64) (Pedido, error) {
	var p Pedido
	err := r.db.QueryRow(`SELECT id, evento_id, numero, total, criado_em FROM pedidos WHERE id = ?`, id).
		Scan(&p.ID, &p.EventoID, &p.Numero, &p.Total, &p.CriadoEm)
	if err != nil {
		return p, err
	}
	itens, err := r.pedidoItens(id)
	if err != nil {
		return p, err
	}
	p.Itens = itens
	return p, nil
}

func (r *repo) pedidoItens(pedidoID int64) ([]PedidoItem, error) {
	out := []PedidoItem{}
	rows, err := r.db.Query(`SELECT id, produto_id, nome, preco_unit, qtd, subtotal
		FROM pedido_itens WHERE pedido_id = ?`, pedidoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var it PedidoItem
		if err := rows.Scan(&it.ID, &it.ProdutoID, &it.Nome, &it.PrecoUnit, &it.Qtd, &it.Subtotal); err != nil {
			return out, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// FecharPedido valida e grava os itens do pedido, baixa o estoque dos
// produtos limitados e recalcula o total. Tudo numa transação.
func (r *repo) FecharPedido(pedidoID int64, itens []PedidoItem) (Pedido, error) {
	var p Pedido
	tx, err := r.db.Begin()
	if err != nil {
		return p, err
	}
	defer tx.Rollback()

	// valida itens
	if len(itens) == 0 {
		return p, errors.New("pedido sem itens")
	}
	for i := range itens {
		it := &itens[i]
		if it.Qtd <= 0 {
			return p, errors.New("quantidade inválida")
		}
		if it.ProdutoID <= 0 {
			return p, errors.New("item sem produto")
		}
	}

	if _, err := tx.Exec(`DELETE FROM pedido_itens WHERE pedido_id = ?`, pedidoID); err != nil {
		return p, err
	}
	var total int64
	for _, it := range itens {
		it.Subtotal = it.PrecoUnit * it.Qtd
		total += it.Subtotal
		if _, err := tx.Exec(`INSERT INTO pedido_itens (pedido_id, produto_id, nome, preco_unit, qtd, subtotal)
			VALUES (?,?,?,?,?,?)`, pedidoID, it.ProdutoID, it.Nome, it.PrecoUnit, it.Qtd, it.Subtotal); err != nil {
			return p, err
		}
		// baixa estoque (se limitado); ilimitado não mexe.
		var tipo string
		if err := tx.QueryRow(`SELECT estoque_tipo FROM produtos WHERE id = ?`, it.ProdutoID).Scan(&tipo); err == nil && tipo == EstoqueLimitado {
			if err := baixaEstoqueTx(tx, it.ProdutoID, it.Qtd); err != nil {
				return p, fmt.Errorf("%s: %w", it.Nome, err)
			}
		}
	}
	if _, err := tx.Exec(`UPDATE pedidos SET total = ? WHERE id = ?`, total, pedidoID); err != nil {
		return p, err
	}
	if err := tx.Commit(); err != nil {
		return p, err
	}
	return r.GetPedido(pedidoID)
}

func baixaEstoqueTx(tx *sql.Tx, produtoID, qtd int64) error {
	res, err := tx.Exec(`UPDATE produtos SET quantidade = quantidade - ? WHERE id = ? AND estoque_tipo = 'limitado' AND quantidade >= ?`,
		qtd, produtoID, qtd)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("estoque insuficiente")
	}
	return nil
}

// =============================================================
// helpers
// =============================================================

type scanner interface {
	Scan(dest ...any) error
}

func scanProduto(s scanner) (Produto, error) {
	var p Produto
	var ativo int64
	err := s.Scan(&p.ID, &p.EventoID, &p.Nome, &p.Preco, &p.EstoqueTipo, &p.Quantidade, &ativo)
	if err != nil {
		return p, err
	}
	p.Ativo = ativo == 1
	return p, nil
}

func requireAffected(res sql.Result, what string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s não encontrado", what)
	}
	return nil
}
