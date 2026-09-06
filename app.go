package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx  context.Context
	db   *sql.DB
	repo *repo
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. Abre o banco e aplica o schema.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	path := dbPath()
	if dir := filepath.Dir(path); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	db, err := openDB(path)
	if err != nil {
		// Sem UI ainda neste ponto; loga e segue — as chamadas retornarão erro.
		println("Vendinha: falha ao abrir banco:", err.Error())
		a.db = nil
		a.repo = NewRepo(nil)
		return
	}
	a.db = db
	a.repo = NewRepo(db)
	// Migra preferências do antigo config.json (se houver) para a tabela config.
	_ = importaConfigJsonLegado(db)
}

// Greet kept from template (harmless), remove later if unused.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
