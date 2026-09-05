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
	id        INTEGER PRIMARY KEY AUTOINCREMENT,
	nome      TEXT    NOT NULL,
	ativo     INTEGER NOT NULL DEFAULT 1,
	criado_em TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS produtos (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id    INTEGER NOT NULL REFERENCES eventos(id) ON DELETE CASCADE,
	nome         TEXT    NOT NULL,
	preco        INTEGER NOT NULL DEFAULT 0,          -- em centavos
	estoque_tipo TEXT    NOT NULL DEFAULT 'ilimitado', -- 'ilimitado' | 'limitado'
	quantidade   INTEGER NOT NULL DEFAULT 0,
	ativo        INTEGER NOT NULL DEFAULT 1,
	criado_em    TEXT    NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS pedidos (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	evento_id  INTEGER NOT NULL REFERENCES eventos(id),
	numero     INTEGER NOT NULL,                        -- sequencial por evento
	total      INTEGER NOT NULL DEFAULT 0,              -- centavos
	criado_em  TEXT    NOT NULL DEFAULT (datetime('now','localtime')),
	UNIQUE (evento_id, numero)
);

CREATE TABLE IF NOT EXISTS pedido_itens (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	pedido_id   INTEGER NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
	produto_id  INTEGER REFERENCES produtos(id),
	nome        TEXT    NOT NULL,   -- snapshot do nome na venda
	preco_unit  INTEGER NOT NULL,
	qtd         INTEGER NOT NULL,
	subtotal    INTEGER NOT NULL
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
	return db, nil
}

// dbPath resolve o caminho do arquivo de dados em %APPDATA%\vendinha\vendinha.db.
func dbPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "vendinha", "vendinha.db")
}
