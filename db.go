package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS eventos (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	nome          TEXT    NOT NULL,
	ativo         INTEGER NOT NULL DEFAULT 1,
	vende_cartela INTEGER NOT NULL DEFAULT 0,   -- habilita venda de cartelas no PDV
	criado_em     TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS grupos (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id  INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
	nome       TEXT    NOT NULL,
	cor        TEXT    NOT NULL DEFAULT '#10b981',
	ordem      INTEGER NOT NULL DEFAULT 0,
	criado_em  TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS produtos (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id    INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
	nome         TEXT    NOT NULL,
	preco        INTEGER NOT NULL DEFAULT 0,          -- em centavos
	estoque_tipo TEXT    NOT NULL DEFAULT 'ilimitado', -- 'ilimitado' | 'limitado'
	quantidade   INTEGER NOT NULL DEFAULT 0,
	ativo        INTEGER NOT NULL DEFAULT 1,
	grupo_id     INTEGER REFERENCES grupos(id) ON DELETE SET NULL,
	ordem        INTEGER NOT NULL DEFAULT 0,
	atalho       INTEGER,                              -- tecla 1-9 no PDV (NULL = sem)
	criado_em    TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS contas (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id  INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
	nome       TEXT    NOT NULL COLLATE NOCASE,        -- comparação e unicidade case-insensitive
	criado_em  TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
	UNIQUE (evento_id, nome COLLATE NOCASE)
);

CREATE TABLE IF NOT EXISTS pedidos (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id  INTEGER NOT NULL REFERENCES eventos(id),
	numero     INTEGER NOT NULL,                        -- sequencial por evento
	total      INTEGER NOT NULL DEFAULT 0,              -- centavos
	forma      TEXT    NOT NULL DEFAULT 'dinheiro',     -- 'dinheiro' | 'cartao' | 'anotaai'
	conta_id   INTEGER REFERENCES contas(id) ON DELETE SET NULL, -- conta no "Anota aí"
	quitado_em TEXT,   -- preenchido quando a venda anotada é quitada (NULL = em aberto)
	cancelado_em TEXT, -- preenchido quando o pedido é cancelado (NULL = ativo)
	criado_em  TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
	UNIQUE (evento_id, numero)
);

CREATE TABLE IF NOT EXISTS pedido_itens (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	pedido_id      INTEGER NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
	produto_id     INTEGER REFERENCES produtos(id),  -- NULL quando é cartela
	cartela_reais  INTEGER,                          -- denominação da cartela (NULL = produto)
	nome           TEXT    NOT NULL,   -- snapshot do nome na venda
	preco_unit     INTEGER NOT NULL,
	qtd            INTEGER NOT NULL,
	subtotal       INTEGER NOT NULL
);
`

// openDB abre o banco SQLite em path (":memory:" ou arquivo), aplicando PRAGMAs.
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// PRAGMAs são por-conexão; MaxOpenConns(1) garante que persistam.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	// Bancos criados antes das colunas de grupo/ordem ganham-nas aqui.
	// Sem sistema de migração: verifica PRAGMA table_info e faz ALTER quando falta.
	if err := migrateProdutos(db); err != nil {
		db.Close()
		return nil, err
	}
	// Bancos antigos não têm vende_cartela (eventos) nem cartela_reais
	// (pedido_itens). Adiciona via ALTER quando faltam.
	if err := migrateVendeCartela(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrateAnotaAi(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrateQuitadoPedido(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := migrateCanceladoPedido(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// ensureColumn adiciona uma coluna em `table` via ALTER TABLE se ainda não existir.
func ensureColumn(db *sql.DB, table, name, def string) error {
	cols, err := tableColumns(db, table)
	if err != nil {
		return err
	}
	for _, c := range cols {
		if c == name {
			return nil
		}
	}
	_, err = db.Exec("ALTER TABLE " + table + " ADD COLUMN " + def)
	return err
}

// migrateVendeCartela garante as colunas do recurso de cartelas em bancos
// antigos: eventos.vende_cartela e pedido_itens.cartela_reais.
func migrateVendeCartela(db *sql.DB) error {
	if err := ensureColumn(db, "eventos", "vende_cartela", "vende_cartela INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate eventos.vende_cartela: %w", err)
	}
	if err := ensureColumn(db, "pedido_itens", "cartela_reais", "cartela_reais INTEGER"); err != nil {
		return fmt.Errorf("migrate pedido_itens.cartela_reais: %w", err)
	}
	return nil
}

// migrateAnotaAi garante as colunas do recurso "Anota aí" (fiado) em bancos
// antigos: pedidos.forma e pedidos.conta_id. A tabela `contas` é criada pelo
// schema em toda abertura (CREATE TABLE IF NOT EXISTS), então não precisa de
// migração própria.
func migrateAnotaAi(db *sql.DB) error {
	if err := ensureColumn(db, "pedidos", "forma", "forma TEXT NOT NULL DEFAULT 'dinheiro'"); err != nil {
		return fmt.Errorf("migrate pedidos.forma: %w", err)
	}
	if err := ensureColumn(db, "pedidos", "conta_id", "conta_id INTEGER REFERENCES contas(id) ON DELETE SET NULL"); err != nil {
		return fmt.Errorf("migrate pedidos.conta_id: %w", err)
	}
	return nil
}

// migrateQuitadoPedido garante a coluna pedidos.quitado_em (quitação do
// "Anota aí") em bancos antigos.
func migrateQuitadoPedido(db *sql.DB) error {
	if err := ensureColumn(db, "pedidos", "quitado_em", "quitado_em TEXT"); err != nil {
		return fmt.Errorf("migrate pedidos.quitado_em: %w", err)
	}
	return nil
}

// migrateCanceladoPedido garante a coluna pedidos.cancelado_em (cancelamento de
// pedidos fechados) em bancos antigos.
func migrateCanceladoPedido(db *sql.DB) error {
	if err := ensureColumn(db, "pedidos", "cancelado_em", "cancelado_em TEXT"); err != nil {
		return fmt.Errorf("migrate pedidos.cancelado_em: %w", err)
	}
	return nil
}

// migrateProdutos garante as colunas de agrupamento/atalho em produtos,
// adicionando-as via ALTER TABLE quando um banco antigo ainda não as tem
// (idempotente), e cria o índice único parcial do atalho (por evento).
func migrateProdutos(db *sql.DB) error {
	cols, err := tableColumns(db, "produtos")
	if err != nil {
		return fmt.Errorf("migrate produtos: %w", err)
	}
	add := func(name, def string) error {
		for _, c := range cols {
			if c == name {
				return nil
			}
		}
		_, err := db.Exec("ALTER TABLE produtos ADD COLUMN " + def)
		return err
	}
	if err := add("grupo_id", "grupo_id INTEGER REFERENCES grupos(id) ON DELETE SET NULL"); err != nil {
		return fmt.Errorf("migrate produtos.grupo_id: %w", err)
	}
	if err := add("ordem", "ordem INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("migrate produtos.ordem: %w", err)
	}
	if err := add("atalho", "atalho INTEGER"); err != nil {
		return fmt.Errorf("migrate produtos.atalho: %w", err)
	}
	// Índice único parcial: garante que um mesmo evento não tenha dois produtos
	// com a mesma tecla de atalho. NULL (sem atalho) não colide entre si.
	// DROP + CREATE a cada abertura: corrige bancos que já criaram a versão
	// errada (só `evento_id`) — IF NOT EXISTS não trocaria a definição antiga.
	if _, err := db.Exec(`DROP INDEX IF EXISTS idx_produtos_atalho`); err != nil {
		return fmt.Errorf("migrate produtos.atalho index: %w", err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_produtos_atalho
		ON produtos(evento_id, atalho) WHERE atalho IS NOT NULL`); err != nil {
		return fmt.Errorf("migrate produtos.atalho index: %w", err)
	}
	return nil
}

// tableColumns lista os nomes das colunas de uma tabela.
func tableColumns(db *sql.DB, table string) ([]string, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// dbPath resolve o caminho do arquivo de dados em %APPDATA%\vendinha\vendinha.db.
func dbPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "vendinha", "vendinha.db")
}
