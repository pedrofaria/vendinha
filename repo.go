package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// =============================================================
// Eventos
// =============================================================

func (r *repo) ListEventos() ([]Evento, error) {
	out := []Evento{}
	rows, err := r.db.Query(`SELECT id, nome, ativo, vende_cartela, criado_em FROM eventos ORDER BY criado_em DESC, id DESC`)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var e Evento
		var ativo, vc int
		if err := rows.Scan(&e.ID, &e.Nome, &ativo, &vc, &e.CriadoEm); err != nil {
			return out, err
		}
		e.Ativo = ativo == 1
		e.VendeCartela = vc == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *repo) GetEvento(id int64) (Evento, error) {
	var e Evento
	var ativo, vc int
	err := r.db.QueryRow(`SELECT id, nome, ativo, vende_cartela, criado_em FROM eventos WHERE id = ?`, id).
		Scan(&e.ID, &e.Nome, &ativo, &vc, &e.CriadoEm)
	if err != nil {
		return e, err
	}
	e.Ativo = ativo == 1
	e.VendeCartela = vc == 1
	return e, nil
}

func (r *repo) CreateEvento(nome string, vendeCartela bool) (Evento, error) {
	var e Evento
	vc := 0
	if vendeCartela {
		vc = 1
	}
	res, err := r.db.Exec(`INSERT INTO eventos (nome, vende_cartela) VALUES (?,?)`, nome, vc)
	if err != nil {
		return e, err
	}
	id, _ := res.LastInsertId()
	return r.GetEvento(id)
}

func (r *repo) UpdateEvento(id int64, nome string, ativo bool, vendeCartela bool) error {
	a := 0
	if ativo {
		a = 1
	}
	vc := 0
	if vendeCartela {
		vc = 1
	}
	res, err := r.db.Exec(`UPDATE eventos SET nome = ?, ativo = ?, vende_cartela = ? WHERE id = ?`, nome, a, vc, id)
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

// ClonarEvento duplica um evento e sua estrutura (grupos e produtos), ignorando
// pedidos, vendas e contas do "Anota aí". O novo evento nasce ativo, com o mesmo
// nome + " (cópia)", e herda vende_cartela, grupos (nome/cor/ordem) e produtos
// (nome, preço, estoque, ativo, ordem, atalho e vínculo de grupo — remapeado
// para o novo grupo). Ordem e atalho são preservados porque o evento novo começa
// vazio (sem conflito de tecla). Tudo roda numa única transação.
func (r *repo) ClonarEvento(id int64) (Evento, error) {
	src, err := r.GetEvento(id)
	if err != nil {
		return Evento{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return Evento{}, err
	}
	defer tx.Rollback()

	vc := 0
	if src.VendeCartela {
		vc = 1
	}
	res, err := tx.Exec(`INSERT INTO eventos (nome, vende_cartela) VALUES (?, ?)`, src.Nome+" (cópia)", vc)
	if err != nil {
		return Evento{}, err
	}
	novoID, _ := res.LastInsertId()

	// ---- grupos: materializa e insere, mapeando id antigo -> novo ----
	gRows, err := tx.Query(`SELECT id, nome, cor, ordem FROM grupos WHERE evento_id = ?`, id)
	if err != nil {
		return Evento{}, err
	}
	type linhaGrupo struct {
		oldID int64
		nome  string
		cor   string
		ordem int64
	}
	grupos := []linhaGrupo{}
	for gRows.Next() {
		var g linhaGrupo
		if err := gRows.Scan(&g.oldID, &g.nome, &g.cor, &g.ordem); err != nil {
			gRows.Close()
			return Evento{}, err
		}
		grupos = append(grupos, g)
	}
	if err := gRows.Close(); err != nil {
		return Evento{}, err
	}
	grupoMap := map[int64]int64{} // grupo antigo -> novo
	for _, g := range grupos {
		gres, err := tx.Exec(`INSERT INTO grupos (evento_id, nome, cor, ordem) VALUES (?,?,?,?)`,
			novoID, g.nome, g.cor, g.ordem)
		if err != nil {
			return Evento{}, err
		}
		gid, _ := gres.LastInsertId()
		grupoMap[g.oldID] = gid
	}

	// ---- produtos: materializa e insere, remapeando grupo_id ----
	pRows, err := tx.Query(`SELECT nome, preco, estoque_tipo, quantidade, ativo, grupo_id, ordem, atalho
		FROM produtos WHERE evento_id = ?`, id)
	if err != nil {
		return Evento{}, err
	}
	type linhaProduto struct {
		nome       string
		preco      int64
		tipo       string
		quantidade int64
		ativo      int64
		grupo      sql.NullInt64
		ordem      int64
		atalho     sql.NullInt64
	}
	produtos := []linhaProduto{}
	for pRows.Next() {
		var p linhaProduto
		if err := pRows.Scan(&p.nome, &p.preco, &p.tipo, &p.quantidade, &p.ativo, &p.grupo, &p.ordem, &p.atalho); err != nil {
			pRows.Close()
			return Evento{}, err
		}
		produtos = append(produtos, p)
	}
	if err := pRows.Close(); err != nil {
		return Evento{}, err
	}
	for _, p := range produtos {
		var gkey any // NULL quando o produto era "sem grupo"
		if p.grupo.Valid {
			gkey = grupoMap[p.grupo.Int64]
		}
		akey := atalhoKey(int64(0)) // NULL
		if p.atalho.Valid {
			akey = p.atalho.Int64
		}
		if _, err := tx.Exec(`INSERT INTO produtos
			(evento_id, nome, preco, estoque_tipo, quantidade, ativo, grupo_id, ordem, atalho)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			novoID, p.nome, p.preco, p.tipo, p.quantidade, p.ativo, gkey, p.ordem, akey); err != nil {
			return Evento{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Evento{}, err
	}
	return r.GetEvento(novoID)
}

// =============================================================
// Grupos
// =============================================================

func (r *repo) gruposOrdenados(eventoID int64) ([]Grupo, error) {
	out := []Grupo{}
	rows, err := r.db.Query(`SELECT id, evento_id, nome, cor, ordem
		FROM grupos WHERE evento_id = ? ORDER BY ordem, id`, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var g Grupo
		if err := rows.Scan(&g.ID, &g.EventoID, &g.Nome, &g.Cor, &g.Ordem); err != nil {
			return out, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *repo) GetGrupo(id int64) (Grupo, error) {
	var g Grupo
	err := r.db.QueryRow(`SELECT id, evento_id, nome, cor, ordem FROM grupos WHERE id = ?`, id).
		Scan(&g.ID, &g.EventoID, &g.Nome, &g.Cor, &g.Ordem)
	if err != nil {
		return g, err
	}
	return g, nil
}

func (r *repo) CreateGrupo(eventoID int64, nome, cor string) (Grupo, error) {
	var g Grupo
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return g, errors.New("nome do grupo vazio")
	}
	cor = strings.TrimSpace(cor)
	if cor == "" {
		cor = CorPadraoGrupo
	}
	var prox int64
	if err := r.db.QueryRow(`SELECT COALESCE(MAX(ordem),0)+1 FROM grupos WHERE evento_id = ?`, eventoID).Scan(&prox); err != nil {
		return g, err
	}
	res, err := r.db.Exec(`INSERT INTO grupos (evento_id, nome, cor, ordem) VALUES (?,?,?,?)`, eventoID, nome, cor, prox)
	if err != nil {
		return g, err
	}
	id, _ := res.LastInsertId()
	return r.GetGrupo(id)
}

func (r *repo) UpdateGrupo(id int64, nome, cor string) error {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return errors.New("nome do grupo vazio")
	}
	cor = strings.TrimSpace(cor)
	if cor == "" {
		cor = CorPadraoGrupo
	}
	res, err := r.db.Exec(`UPDATE grupos SET nome = ?, cor = ? WHERE id = ?`, nome, cor, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "grupo")
}

// DeleteGrupo remove o grupo; os produtos dele passam a "Sem grupo"
// (FK ON DELETE SET NULL).
func (r *repo) DeleteGrupo(id int64) error {
	res, err := r.db.Exec(`DELETE FROM grupos WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res, "grupo")
}

// MoverGrupo desloca um grupo uma posição na ordem (delta -1 sobe, +1 desce),
// trocando a ordem com o vizinho na mesma direção. No-op quando já está na borda.
func (r *repo) MoverGrupo(id, delta int64) error {
	g, err := r.GetGrupo(id)
	if err != nil {
		return err
	}
	ids, err := r.grupoIDs(g.EventoID)
	if err != nil {
		return err
	}
	idx := indexOf(ids, id)
	to := int64(idx) + delta
	if idx < 0 || to < 0 || to >= int64(len(ids)) || idx == int(to) {
		return nil // borda ou delta 0
	}
	return r.swapOrdemGrupos(id, ids[to])
}

func (r *repo) grupoIDs(eventoID int64) ([]int64, error) {
	out := []int64{}
	rows, err := r.db.Query(`SELECT id FROM grupos WHERE evento_id = ? ORDER BY ordem, id`, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return out, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ReorderGrupos grava a nova ordem de exibição dos grupos de um evento a partir
// de ids na ordem final (usado pelo drag & drop). ids vêm da própria tela.
func (r *repo) ReorderGrupos(eventoID int64, ids []int64) error {
	for i, id := range ids {
		if _, err := r.db.Exec(`UPDATE grupos SET ordem = ? WHERE id = ? AND evento_id = ?`, int64(i+1), id, eventoID); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) swapOrdemGrupos(a, b int64) error {
	var oa, ob int64
	if err := r.db.QueryRow(`SELECT ordem FROM grupos WHERE id = ?`, a).Scan(&oa); err != nil {
		return err
	}
	if err := r.db.QueryRow(`SELECT ordem FROM grupos WHERE id = ?`, b).Scan(&ob); err != nil {
		return err
	}
	if _, err := r.db.Exec(`UPDATE grupos SET ordem = ? WHERE id = ?`, ob, a); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE grupos SET ordem = ? WHERE id = ?`, oa, b)
	return err
}

// =============================================================
// Produtos
// =============================================================

// ListProdutos retorna os grupos de um evento (ordenados) com seus produtos
// já dentro (ordenados), para a tela de gerenciamento. O grupo virtual
// "Sem grupo" (produtos sem grupo_id) aparece por último, só quando tem itens.
func (r *repo) ListProdutos(eventoID int64) ([]GrupoProdutos, error) {
	grupos, err := r.gruposOrdenados(eventoID)
	if err != nil {
		return nil, err
	}
	out := make([]GrupoProdutos, 0, len(grupos)+1)
	byID := make(map[int64]int, len(grupos))
	for i := range grupos {
		out = append(out, GrupoProdutos{Grupo: grupos[i], Produtos: []Produto{}})
		byID[grupos[i].ID] = i
	}
	semGrupo := []Produto{}
	rows, err := r.db.Query(`SELECT id, evento_id, nome, preco, estoque_tipo, quantidade, ativo, grupo_id, atalho, ordem
		FROM produtos WHERE evento_id = ? ORDER BY ordem, id`, eventoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, err
		}
		if i, ok := byID[p.GrupoID]; ok {
			out[i].Produtos = append(out[i].Produtos, p)
		} else {
			semGrupo = append(semGrupo, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(semGrupo) > 0 {
		out = append(out, GrupoProdutos{
			Grupo:    Grupo{ID: SemGrupoID, EventoID: eventoID, Nome: "Sem grupo", Cor: CorSemGrupo},
			Produtos: semGrupo,
		})
	}
	return out, nil
}

// ListProdutosVenda retorna os grupos ativos de um evento com seus produtos
// à venda (ativos, ordenados), como o PDV renderiza. Ignora grupos que ficaram
// sem nenhum produto ativo; o bucket "Sem grupo" só aparece se tiver itens.
func (r *repo) ListProdutosVenda(eventoID int64) ([]GrupoVenda, error) {
	grupos, err := r.gruposOrdenados(eventoID)
	if err != nil {
		return nil, err
	}
	out := make([]GrupoVenda, 0, len(grupos)+1)
	byID := make(map[int64]int, len(grupos))
	for i := range grupos {
		out = append(out, GrupoVenda{Grupo: grupos[i], Produtos: []ProdutoVenda{}})
		byID[grupos[i].ID] = i
	}
	semGrupo := []ProdutoVenda{}
	rows, err := r.db.Query(`SELECT id, nome, preco, estoque_tipo, quantidade, ativo, grupo_id, atalho
		FROM produtos WHERE evento_id = ? ORDER BY ordem, id`, eventoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pv ProdutoVenda
		var tipo string
		var qtd, ativo int64
		var g, a sql.NullInt64
		if err := rows.Scan(&pv.ID, &pv.Nome, &pv.Preco, &tipo, &qtd, &ativo, &g, &a); err != nil {
			return nil, err
		}
		if ativo == 0 {
			continue // não oferece produtos inativos na venda
		}
		pv.EstoqueTipo = tipo
		pv.Quantidade = qtd
		pv.Esgotado = tipo == EstoqueLimitado && qtd <= 0
		if a.Valid {
			pv.Atalho = a.Int64
		}
		grupoID := int64(0)
		if g.Valid {
			grupoID = g.Int64
		}
		if i, ok := byID[grupoID]; ok {
			out[i].Produtos = append(out[i].Produtos, pv)
		} else {
			semGrupo = append(semGrupo, pv)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := len(out) - 1; i >= 0; i-- {
		if len(out[i].Produtos) == 0 {
			out = append(out[:i], out[i+1:]...) // descarta grupo sem produto ativo
		}
	}
	if len(semGrupo) > 0 {
		out = append(out, GrupoVenda{
			Grupo:    Grupo{ID: SemGrupoID, EventoID: eventoID, Nome: "Sem grupo", Cor: CorSemGrupo},
			Produtos: semGrupo,
		})
	}
	return out, nil
}

func (r *repo) GetProduto(id int64) (Produto, error) {
	row := r.db.QueryRow(`SELECT id, evento_id, nome, preco, estoque_tipo, quantidade, ativo, grupo_id, atalho, ordem
		FROM produtos WHERE id = ?`, id)
	return scanProduto(row)
}

// nextOrdemProduto devolve a próxima ordem dentro do grupo de um produto
// (grupoID 0 = sem grupo; ordem é relativa entre irmãos do mesmo grupo).
func (r *repo) nextOrdemProduto(eventoID, grupoID int64) (int64, error) {
	var key any
	if grupoID != SemGrupoID {
		key = grupoID
	}
	var prox int64
	err := r.db.QueryRow(`SELECT COALESCE(MAX(ordem),0)+1 FROM produtos
		WHERE evento_id = ? AND (grupo_id IS ?)`, eventoID, key).Scan(&prox)
	return prox, err
}

func (r *repo) CreateProduto(eventoID int64, nome string, preco int64, tipo string, quantidade, grupoID, atalho int64) (Produto, error) {
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
	if err := r.exigeAtalhoLivre(eventoID, atalho, 0); err != nil {
		return p, err
	}
	ordem, err := r.nextOrdemProduto(eventoID, grupoID)
	if err != nil {
		return p, err
	}
	var key any
	if grupoID != SemGrupoID {
		key = grupoID
	}
	res, err := r.db.Exec(`INSERT INTO produtos (evento_id, nome, preco, estoque_tipo, quantidade, grupo_id, ordem, atalho)
		VALUES (?,?,?,?,?,?,?,?)`, eventoID, nome, preco, tipo, quantidade, key, ordem, atalhoKey(atalho))
	if err != nil {
		return p, err
	}
	id, _ := res.LastInsertId()
	return r.GetProduto(id)
}

func (r *repo) UpdateProduto(id int64, nome string, preco int64, tipo string, quantidade, grupoID, atalho int64) error {
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
	cur, err := r.GetProduto(id)
	if err != nil {
		return err
	}
	if err := r.exigeAtalhoLivre(cur.EventoID, atalho, id); err != nil {
		return err
	}
	// Se mudou de grupo, joga o produto para o fim do novo grupo para não
	// "atravessar" a ordem; se permanece no mesmo grupo, mantém a ordem atual.
	ordem := cur.Ordem
	if cur.GrupoID != grupoID {
		ordem, err = r.nextOrdemProduto(cur.EventoID, grupoID)
		if err != nil {
			return err
		}
	}
	var key any
	if grupoID != SemGrupoID {
		key = grupoID
	}
	res, err := r.db.Exec(`UPDATE produtos SET nome = ?, preco = ?, estoque_tipo = ?, quantidade = ?, grupo_id = ?, ordem = ?, atalho = ? WHERE id = ?`,
		nome, preco, tipo, quantidade, key, ordem, atalhoKey(atalho), id)
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

// MoverProduto desloca um produto uma posição dentro do próprio grupo
// (delta -1 sobe, +1 desce), trocando a ordem com o vizinho irmão.
func (r *repo) MoverProduto(id, delta int64) error {
	cur, err := r.GetProduto(id)
	if err != nil {
		return err
	}
	ids, err := r.produtoIrmaos(cur.EventoID, cur.GrupoID)
	if err != nil {
		return err
	}
	idx := indexOf(ids, id)
	to := int64(idx) + delta
	if idx < 0 || to < 0 || to >= int64(len(ids)) || idx == int(to) {
		return nil // borda ou delta 0
	}
	return r.swapOrdemProdutos(id, ids[to])
}

// produtoIrmaos lista os ids dos produtos do mesmo grupo (ordem relativa).
// GrupoID 0 = "Sem grupo" (grupo_id IS NULL).
func (r *repo) produtoIrmaos(eventoID, grupoID int64) ([]int64, error) {
	var key any
	if grupoID != SemGrupoID {
		key = grupoID
	}
	out := []int64{}
	rows, err := r.db.Query(`SELECT id FROM produtos
		WHERE evento_id = ? AND (grupo_id IS ?) ORDER BY ordem, id`, eventoID, key)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return out, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ReorderProdutos grava a ordem dos produtos de um mesmo grupo a partir de ids
// na ordem final (drag & drop). ids = conjunto completo (ordenado) dos irmãos.
func (r *repo) ReorderProdutos(ids []int64) error {
	for i, id := range ids {
		if _, err := r.db.Exec(`UPDATE produtos SET ordem = ? WHERE id = ?`, int64(i+1), id); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) swapOrdemProdutos(a, b int64) error {
	var oa, ob int64
	if err := r.db.QueryRow(`SELECT ordem FROM produtos WHERE id = ?`, a).Scan(&oa); err != nil {
		return err
	}
	if err := r.db.QueryRow(`SELECT ordem FROM produtos WHERE id = ?`, b).Scan(&ob); err != nil {
		return err
	}
	if _, err := r.db.Exec(`UPDATE produtos SET ordem = ? WHERE id = ?`, ob, a); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE produtos SET ordem = ? WHERE id = ?`, oa, b)
	return err
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

// scanPedidoBase lê uma linha de pedidos (mesmas colunas de pedidoScanCols) em
// um Pedido. Converte NULLs (conta_id, quitado_em) e normaliza forma vazia.
type pedidoScaner interface {
	Scan(dest ...any) error
}

const pedidoScanCols = `id, evento_id, numero, total, forma, conta_id, criado_em, quitado_em, cancelado_em`

func scanPedidoBase(s pedidoScaner) (Pedido, error) {
	var p Pedido
	var conta sql.NullInt64
	var quitado, cancelado sql.NullString
	if err := s.Scan(&p.ID, &p.EventoID, &p.Numero, &p.Total, &p.Forma, &conta, &p.CriadoEm, &quitado, &cancelado); err != nil {
		return p, err
	}
	if conta.Valid {
		p.ContaID = conta.Int64
	}
	if quitado.Valid {
		p.QuitadoEm = quitado.String
	}
	if cancelado.Valid {
		p.CanceladoEm = cancelado.String
	}
	if p.Forma == "" {
		p.Forma = FormaDinheiro // pedidos anteriores à coluna forma
	}
	return p, nil
}

func (r *repo) GetPedido(id int64) (Pedido, error) {
	p, err := scanPedidoBase(r.db.QueryRow(`SELECT `+pedidoScanCols+` FROM pedidos WHERE id = ?`, id))
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

// ListPedidosConta devolve os pedidos "Anota aí" de uma conta (em aberto e já
// quitados), mais recentes primeiro, com os itens de cada um.
//
// O banco roda com MaxOpenConns(1), então NÃO podemos buscar os itens enquanto
// o rows do SELECT pedidos estiver aberto (segunda query esperaria a conexão
// única → deadlock). Primeiro materializa os pedidos e fecha o rows; depois
// anexa os itens de cada um.
func (r *repo) ListPedidosConta(contaID int64) ([]Pedido, error) {
	out := []Pedido{}
	rows, err := r.db.Query(`SELECT `+pedidoScanCols+` FROM pedidos
		WHERE conta_id = ? AND forma = ? AND cancelado_em IS NULL
		ORDER BY criado_em DESC, id DESC`, contaID, FormaAnotaAi)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		p, err := scanPedidoBase(rows)
		if err != nil {
			rows.Close()
			return out, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return out, err
	}
	rows.Close()
	for i := range out {
		itens, err := r.pedidoItens(out[i].ID)
		if err != nil {
			return out, err
		}
		out[i].Itens = itens
	}
	return out, nil
}

func (r *repo) pedidoItens(pedidoID int64) ([]PedidoItem, error) {
	out := []PedidoItem{}
	rows, err := r.db.Query(`SELECT id, produto_id, cartela_reais, nome, preco_unit, qtd, subtotal
		FROM pedido_itens WHERE pedido_id = ?`, pedidoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var it PedidoItem
		var pid, creais sql.NullInt64 // NULL quando a linha é o outro tipo
		if err := rows.Scan(&it.ID, &pid, &creais, &it.Nome, &it.PrecoUnit, &it.Qtd, &it.Subtotal); err != nil {
			return out, err
		}
		if pid.Valid {
			it.ProdutoID = pid.Int64
		}
		if creais.Valid {
			it.CartelaReais = creais.Int64
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Por padrão a listagem de pedidos traz 20 por página.
const pedidosPorPaginaPadrao int64 = 20

// ListPedidos devolve uma página dos pedidos fechados de um evento (total > 0),
// do mais recente para o mais antigo, com dados leves e o nome da conta do
// "Anota aí" quando houver. Filtra por número quando numeroBusca > 0. Os itens
// não vêm aqui — a UI chama GetPedido para o detalhe de um pedido.
func (r *repo) ListPedidos(eventoID, pagina, porPagina, numeroBusca int64) (ListaPedidos, error) {
	out := ListaPedidos{Pedidos: []PedidoResumo{}}
	if pagina < 1 {
		pagina = 1
	}
	if porPagina < 1 {
		porPagina = pedidosPorPaginaPadrao
	}
	if porPagina > 100 {
		porPagina = 100
	}
	// Só pedidos fechados (total > 0). Cancelados continuam na listagem
	// (histórico), marcados por cancelado_em — não há filtro por status aqui.
	where := `WHERE p.evento_id = ? AND p.total > 0`
	args := []any{eventoID}
	if numeroBusca > 0 {
		where += ` AND p.numero = ?`
		args = append(args, numeroBusca)
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM pedidos p `+where, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	out.TotalPaginas = (out.Total + porPagina - 1) / porPagina
	if out.TotalPaginas < 1 {
		out.TotalPaginas = 1
	}
	if pagina > out.TotalPaginas {
		pagina = out.TotalPaginas
	}
	out.Pagina = pagina
	pageArgs := append(append([]any{}, args...), porPagina, (pagina-1)*porPagina)
	rows, err := r.db.Query(`SELECT p.id, p.evento_id, p.numero, p.total, p.forma,
			p.conta_id, p.criado_em, p.quitado_em, p.cancelado_em, c.nome
		FROM pedidos p
		LEFT JOIN contas c ON c.id = p.conta_id `+where+`
		ORDER BY p.criado_em DESC, p.id DESC
		LIMIT ? OFFSET ?`, pageArgs...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var pr PedidoResumo
		var conta sql.NullInt64
		var quitado, cancelado, nome sql.NullString
		if err := rows.Scan(&pr.ID, &pr.EventoID, &pr.Numero, &pr.Total, &pr.Forma,
			&conta, &pr.CriadoEm, &quitado, &cancelado, &nome); err != nil {
			return out, err
		}
		if conta.Valid {
			pr.ContaID = conta.Int64
		}
		if quitado.Valid {
			pr.QuitadoEm = quitado.String
		}
		if cancelado.Valid {
			pr.CanceladoEm = cancelado.String
		}
		if nome.Valid {
			pr.ContaNome = nome.String
		}
		if pr.Forma == "" {
			pr.Forma = FormaDinheiro
		}
		out.Pedidos = append(out.Pedidos, pr)
	}
	return out, rows.Err()
}

// CancelarPedido cancela um pedido fechado (total > 0) e ainda não cancelado:
// devolve o estoque dos produtos limitados baixados nele e marca cancelado_em.
// Se a venda for "Anota aí", o pedido deixa de contar na pendência da conta
// automaticamente (as consultas de saldo ignoram pedidos cancelados) — quando a
// venda ainda estava em aberto, o débito some da conta.
//
// Tudo numa transação. O banco roda com MaxOpenConns(1): materializa as linhas
// do pedido e fecha o rows antes de executar os UPDATEs de estoque.
func (r *repo) CancelarPedido(pedidoID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Valida: pedido existe, já fechado e não foi cancelado antes.
	var total int64
	var cancelado sql.NullString
	if err := tx.QueryRow(`SELECT total, cancelado_em FROM pedidos WHERE id = ?`, pedidoID).
		Scan(&total, &cancelado); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("pedido não encontrado")
		}
		return err
	}
	if cancelado.Valid && cancelado.String != "" {
		return errors.New("pedido já cancelado")
	}
	if total <= 0 {
		return errors.New("pedido ainda não foi fechado")
	}

	// Devolve o estoque dos produtos limitados do pedido (ilimitado não tem
	// estoque p/ devolver; cartela também não baixou).
	type reposicao struct {
		produtoID, qtd int64
	}
	var devolver []reposicao
	rows, err := tx.Query(`SELECT produto_id, qtd FROM pedido_itens
		WHERE pedido_id = ? AND produto_id IS NOT NULL`, pedidoID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var rp reposicao
		if err := rows.Scan(&rp.produtoID, &rp.qtd); err != nil {
			rows.Close()
			return err
		}
		devolver = append(devolver, rp)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, rp := range devolver {
		if _, err := tx.Exec(`UPDATE produtos SET quantidade = quantidade + ?
			WHERE id = ? AND estoque_tipo = 'limitado'`, rp.qtd, rp.produtoID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE pedidos SET cancelado_em = datetime('now','localtime')
		WHERE id = ?`, pedidoID); err != nil {
		return err
	}
	return tx.Commit()
}

// ListContas devolve as contas "Anota aí" de um evento, ordenadas por nome, já
// com o TotalPendente (soma dos totais dos pedidos anotaai). O autocomplete do
// PDV usa essa lista para sugerir donos existentes e quanto já está anotado.
func (r *repo) ListContas(eventoID int64) ([]Conta, error) {
	out := []Conta{}
	rows, err := r.db.Query(`SELECT c.id, c.evento_id, c.nome, c.criado_em,
			COALESCE(SUM(p.total), 0)
		FROM contas c
		LEFT JOIN pedidos p ON p.conta_id = c.id AND p.forma = ? AND p.quitado_em IS NULL AND p.cancelado_em IS NULL
		WHERE c.evento_id = ?
		GROUP BY c.id
		ORDER BY c.nome COLLATE NOCASE`, FormaAnotaAi, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var c Conta
		if err := rows.Scan(&c.ID, &c.EventoID, &c.Nome, &c.CriadoEm, &c.TotalPendente); err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ListContasSaldo devolve as contas "Anota aí" de um evento com o saldo para a
// tela de gerenciamento: quanto cada uma deve em aberto (vendas anotadas não
// quitadas), quantas vendas em aberto e quanto já quitou. Ordenadas por valor
// pendente decrescente (maior devedor primeiro) e depois por nome.
func (r *repo) ListContasSaldo(eventoID int64) ([]ContaSaldo, error) {
	out := []ContaSaldo{}
	rows, err := r.db.Query(`SELECT c.id, c.evento_id, c.nome, c.criado_em,
			COALESCE(SUM(CASE WHEN p.quitado_em IS NULL THEN p.total END), 0),
			COALESCE(SUM(CASE WHEN p.id IS NOT NULL AND p.quitado_em IS NULL THEN 1 END), 0),
			COALESCE(SUM(CASE WHEN p.quitado_em IS NOT NULL THEN p.total END), 0)
		FROM contas c
		LEFT JOIN pedidos p ON p.conta_id = c.id AND p.forma = ? AND p.cancelado_em IS NULL
		WHERE c.evento_id = ?
		GROUP BY c.id
		ORDER BY 5 DESC, c.nome COLLATE NOCASE`, FormaAnotaAi, eventoID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var c ContaSaldo
		if err := rows.Scan(&c.ID, &c.EventoID, &c.Nome, &c.CriadoEm, &c.TotalPendente, &c.NumAberto, &c.TotalQuitado); err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// QuitarConta marca como pagas as vendas anotadas em aberto de uma conta
// (define quitado_em = agora). Nova venda anotada volta a deixar a conta
// pendente automaticamente.
func (r *repo) QuitarConta(contaID int64) error {
	_, err := r.db.Exec(`UPDATE pedidos SET quitado_em = datetime('now','localtime')
		WHERE conta_id = ? AND forma = ? AND quitado_em IS NULL AND cancelado_em IS NULL`, contaID, FormaAnotaAi)
	return err
}

// ResumoEvento agrega as vendas fechadas (total > 0) de um evento para o
// Dashboard: receita, nº de pedidos, unidades de produto vendidas e o
// histograma de vendas por hora do dia (0–23). Cartelas entram na receita e no
// nº de pedidos, mas não em NumProdutosVendidos (não são produto).
func (r *repo) ResumoEvento(eventoID int64) (ResumoEvento, error) {
	var res ResumoEvento
	// Receita total e nº de pedidos fechados.
	if err := r.db.QueryRow(`SELECT COALESCE(SUM(total),0), COUNT(*)
		FROM pedidos WHERE evento_id = ? AND total > 0 AND cancelado_em IS NULL`, eventoID).
		Scan(&res.ReceitaTotal, &res.NumPedidos); err != nil {
		return res, err
	}
	// Unidades de produto vendidas (linhas com produto; cartelas ficam de fora).
	if err := r.db.QueryRow(`SELECT COALESCE(SUM(pi.qtd),0)
		FROM pedido_itens pi
		JOIN pedidos p ON p.id = pi.pedido_id
		WHERE p.evento_id = ? AND p.total > 0 AND p.cancelado_em IS NULL AND pi.produto_id IS NOT NULL`, eventoID).
		Scan(&res.NumProdutosVendidos); err != nil {
		return res, err
	}
	// Hora em que o eixo do gráfico deve começar, para respeitar a linha do
	// tempo quando a noite cruza a meia-noite: se a 1ª e a última venda caíram em
	// dias diferentes, o eixo começa na hora da 1ª venda (festa 14h→01h mostra
	// 14h…23h, 0h, 1h); se tudo foi no mesmo dia, começa em 0h como de costume.
	res.VendasInicioHora = 0
	var minEm, maxEm sql.NullString
	if err := r.db.QueryRow(`SELECT MIN(criado_em), MAX(criado_em) FROM pedidos
		WHERE evento_id = ? AND total > 0 AND cancelado_em IS NULL`, eventoID).
		Scan(&minEm, &maxEm); err != nil {
		return res, err
	}
	if minEm.Valid && maxEm.Valid {
		const layout = "2006-01-02 15:04:05" // datetime('now','localtime') no SQLite
		ini, errIni := time.Parse(layout, minEm.String)
		fim, errFim := time.Parse(layout, maxEm.String)
		if errIni == nil && errFim == nil {
			y1, m1, d1 := ini.Date()
			y2, m2, d2 := fim.Date()
			if y1 != y2 || m1 != m2 || d1 != d2 {
				res.VendasInicioHora = int64(ini.Hour())
			}
		}
	}
	// Histograma por hora do dia (0–23), zerado onde não houve venda.
	porHora := make([]HoraVendas, 24)
	for i := range porHora {
		porHora[i].Hora = int64(i)
	}
	rows, err := r.db.Query(`SELECT CAST(strftime('%H', criado_em) AS INTEGER), COUNT(*)
		FROM pedidos WHERE evento_id = ? AND total > 0 AND cancelado_em IS NULL
		GROUP BY CAST(strftime('%H', criado_em) AS INTEGER)`, eventoID)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var hora, n int64
		if err := rows.Scan(&hora, &n); err != nil {
			return res, err
		}
		if hora >= 0 && hora < 24 {
			porHora[hora].Vendas = n
		}
	}
	res.VendasPorHora = porHora
	return res, rows.Err()
}

// queryer é o subconjunto de *sql.DB/*sql.Tx usado na resolução de contas
// dentro da transação do FecharPedido.
type queryer interface {
	QueryRow(query string, args ...any) *sql.Row
}

// contaPorNomeTx busca a conta de eventoID cujo nome bate (case-insensitive,
// por causa da COLLATE NOCASE da coluna), ou devolve (nil, nil) se não existe.
func contaPorNomeTx(q queryer, eventoID int64, nome string) (*Conta, error) {
	var c Conta
	err := q.QueryRow(`SELECT id, evento_id, nome, criado_em
		FROM contas WHERE evento_id = ? AND nome = ? COLLATE NOCASE`, eventoID, nome).
		Scan(&c.ID, &c.EventoID, &c.Nome, &c.CriadoEm)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// validaForma aceita as três formas de pagamento persistidas em pedidos.forma.
func validaForma(forma string) error {
	switch forma {
	case FormaDinheiro, FormaCartao, FormaAnotaAi:
		return nil
	}
	return errors.New("forma de pagamento inválida")
}

// FecharPedido valida e grava os itens do pedido, baixa o estoque dos
// produtos limitados, recalcula o total e persiste a forma de pagamento.
// Tudo numa transação.
// Cada item é de produto (ProdutoID > 0) OU de cartela (CartelaReais > 0);
// itens de cartela não mexem em estoque.
// `forma` é 'dinheiro' | 'cartao' | 'anotaai'. Em 'anotaai', `contaNome` é o dono
// da conta: reusa a conta existente do evento (comparação case-insensitive) ou a
// cria — o pedido fica vinculado a ela (vira a pendência da conta). Nas demais
// formas o pedido não tem conta.
func (r *repo) FecharPedido(pedidoID int64, itens []PedidoItem, forma, contaNome string) (Pedido, error) {
	var p Pedido
	tx, err := r.db.Begin()
	if err != nil {
		return p, err
	}
	defer tx.Rollback()

	if err := validaForma(forma); err != nil {
		return p, err
	}

	// valida itens
	if len(itens) == 0 {
		return p, errors.New("pedido sem itens")
	}
	for i := range itens {
		it := &itens[i]
		if it.Qtd <= 0 {
			return p, errors.New("quantidade inválida")
		}
		ehProduto := it.ProdutoID > 0
		ehCartela := it.CartelaReais > 0
		switch {
		case ehProduto && ehCartela:
			return p, errors.New("item inválido")
		case ehProduto:
		case ehCartela:
			if !ehDenominacaoCartela(it.CartelaReais) {
				return p, fmt.Errorf("cartela de valor inválido: R$ %d", it.CartelaReais)
			}
		default:
			return p, errors.New("item sem produto")
		}
	}

	// Resolve a conta do "Anota aí": evento do pedido + reusa existente (nome
	// case-insensitive) ou cria. contaID fica nil (NULL no banco) nas demais formas.
	var eventoID int64
	if err := tx.QueryRow(`SELECT evento_id FROM pedidos WHERE id = ?`, pedidoID).Scan(&eventoID); err != nil {
		return p, err
	}
	var contaID any
	if forma == FormaAnotaAi {
		nome := strings.TrimSpace(contaNome)
		if nome == "" {
			return p, errors.New("informe o nome do dono da conta para anotar a venda")
		}
		c, err := contaPorNomeTx(tx, eventoID, nome)
		if err != nil {
			return p, err
		}
		if c == nil {
			res, err := tx.Exec(`INSERT INTO contas (evento_id, nome) VALUES (?,?)`, eventoID, nome)
			if err != nil {
				return p, err
			}
			id, _ := res.LastInsertId()
			contaID = id
		} else {
			contaID = c.ID
		}
	}

	if _, err := tx.Exec(`DELETE FROM pedido_itens WHERE pedido_id = ?`, pedidoID); err != nil {
		return p, err
	}
	var total int64
	for _, it := range itens {
		it.Subtotal = it.PrecoUnit * it.Qtd
		total += it.Subtotal
		// produto_id fica NULL em itens de cartela (não referencia produtos).
		var pid any
		if it.ProdutoID > 0 {
			pid = it.ProdutoID
		}
		var creais any
		if it.CartelaReais > 0 {
			creais = it.CartelaReais
		}
		if _, err := tx.Exec(`INSERT INTO pedido_itens (pedido_id, produto_id, cartela_reais, nome, preco_unit, qtd, subtotal)
			VALUES (?,?,?,?,?,?,?)`, pedidoID, pid, creais, it.Nome, it.PrecoUnit, it.Qtd, it.Subtotal); err != nil {
			return p, err
		}
		// baixa estoque (se limitado); ilimitado não mexe. Cartela não baixa.
		if it.ProdutoID > 0 {
			var tipo string
			if err := tx.QueryRow(`SELECT estoque_tipo FROM produtos WHERE id = ?`, it.ProdutoID).Scan(&tipo); err == nil && tipo == EstoqueLimitado {
				if err := baixaEstoqueTx(tx, it.ProdutoID, it.Qtd); err != nil {
					return p, fmt.Errorf("%s: %w", it.Nome, err)
				}
			}
		}
	}
	if _, err := tx.Exec(`UPDATE pedidos SET total = ?, forma = ?, conta_id = ? WHERE id = ?`, total, forma, contaID, pedidoID); err != nil {
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
	var g, a sql.NullInt64
	err := s.Scan(&p.ID, &p.EventoID, &p.Nome, &p.Preco, &p.EstoqueTipo, &p.Quantidade, &ativo, &g, &a, &p.Ordem)
	if err != nil {
		return p, err
	}
	p.Ativo = ativo == 1
	if g.Valid {
		p.GrupoID = g.Int64
	}
	if a.Valid {
		p.Atalho = a.Int64
	}
	return p, nil
}

// atalhoKey converte o atalho do modelo (0 = sem atalho) para o valor de banco:
// NULL quando não há tecla, o inteiro 1-9 caso contrário.
func atalhoKey(atalho int64) any {
	if atalho == 0 {
		return nil
	}
	return atalho
}

// validaAtalho aceita 0 (sem atalho) ou um dígito 1-9.
func validaAtalho(atalho int64) error {
	if atalho == 0 {
		return nil
	}
	if atalho < 1 || atalho > 9 {
		return errors.New("atalho deve ser uma tecla de 1 a 9 (ou 0 para nenhum)")
	}
	return nil
}

// atalhoEmUso devolve o nome de outro produto do mesmo evento que já usa a tecla
// `atalho` (ignorando ignoreID), ou "" se estiver livre.
func (r *repo) atalhoEmUso(eventoID, atalho, ignoreID int64) (string, error) {
	if atalho == 0 {
		return "", nil
	}
	var nome string
	err := r.db.QueryRow(`SELECT nome FROM produtos
		WHERE evento_id = ? AND atalho = ? AND id <> ? ORDER BY id LIMIT 1`, eventoID, atalho, ignoreID).Scan(&nome)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return nome, nil
}

// exigeAtalhoLivre valida o valor e garante que a tecla não pertence a outro
// produto do mesmo evento, retornando um erro amigável quando já está em uso.
func (r *repo) exigeAtalhoLivre(eventoID, atalho, ignoreID int64) error {
	if err := validaAtalho(atalho); err != nil {
		return err
	}
	holder, err := r.atalhoEmUso(eventoID, atalho, ignoreID)
	if err != nil {
		return err
	}
	if holder != "" {
		return fmt.Errorf("atalho %d já é usado pelo produto \"%s\"", atalho, holder)
	}
	return nil
}

// indexOf retorna o índice de v em slice, ou -1 se ausente.
func indexOf(s []int64, v int64) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
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
