package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

// =============================================================
// Tela de Configurações — banco de dados + impressora.
// As preferências do app (impressora, largura da linha...) vivem DENTRO do
// banco, na tabela `config` (chave/valor) — backup único, junto dos dados.
// Os bindings usam *a.repo (o banco aberto); a tabela não é apagada pelo
// "Zerar banco" (config não está na lista de DROP), então impressora/largura
// sobrevivem a um zero.
// =============================================================

// ImpressorasInfo é a resposta de ListImpressoras: nomes disponíveis, a
// impressora padrão do SO e a atualmente selecionada no app (a salva no banco;
// sem nada salvo, a padrão do sistema).
type ImpressorasInfo struct {
	Nomes       []string `json:"nomes"`
	Padrao      string   `json:"padrao"`
	Selecionada string   `json:"selecionada"`
}

// dbOrErr devolve o *sql.DB do app (garante inicialização).
func (a *App) dbOrErr() (*sql.DB, error) {
	r, err := a.repoOrErr()
	if err != nil {
		return nil, err
	}
	return r.db, nil
}

// configGet lê uma chave da tabela `config`. Chave ausente = valor vazio.
func configGet(db *sql.DB, chave string) (string, error) {
	var v sql.NullString
	err := db.QueryRow("SELECT valor FROM config WHERE chave = ?", chave).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return v.String, nil
}

// configSet grava (ou sobrescreve) uma chave na tabela `config`.
func configSet(db *sql.DB, chave, valor string) error {
	_, err := db.Exec(
		"INSERT INTO config(chave, valor) VALUES(?, ?) "+
			"ON CONFLICT(chave) DO UPDATE SET valor = excluded.valor",
		chave, valor)
	return err
}

// GetDBPath devolve o caminho completo do arquivo de banco de dados.
func (a *App) GetDBPath() (string, error) {
	return dbPath(), nil
}

// OpenDBFolder abre o gerenciador de arquivos do SO no diretório que contém o
// arquivo de banco de dados (onde o usuário pode fazer backup do arquivo).
func (a *App) OpenDBFolder() error {
	pasta := filepath.Dir(dbPath())
	if _, err := os.Stat(pasta); err != nil {
		return fmt.Errorf("pasta de dados não encontrada (%s): %w", pasta, err)
	}
	switch runtime.GOOS {
	case "windows":
		// explorer.exe é assíncrono: Start (não Run) pra não travar o app.
		return exec.Command("explorer", pasta).Start()
	case "darwin":
		return exec.Command("open", pasta).Start()
	default:
		return exec.Command("xdg-open", pasta).Start()
	}
}

// ZerarBanco remove TODOS os eventos, grupos, produtos, contas, pedidos e
// itens do banco e o recria vazio (autoincrements voltam a 1). As preferências
// da tela de Configurações (tabela config) são PRESERVADAS. É irreversível.
func (a *App) ZerarBanco() error {
	r, err := a.repoOrErr()
	if err != nil {
		return err
	}
	return zeroBanco(r.db)
}

// ListImpressoras devolve as impressoras disponíveis no sistema operacional.
func (a *App) ListImpressoras() (ImpressorasInfo, error) {
	out := ImpressorasInfo{Nomes: []string{}}
	db, err := a.dbOrErr()
	if err != nil {
		return out, err
	}
	nomes, err := listaImpressoras()
	if err != nil {
		return out, fmt.Errorf("não foi possível listar as impressoras: %w", err)
	}
	out.Nomes = nomes
	padrao, _ := impressoraPadrao()
	out.Padrao = padrao
	// Selecionada = a salva no banco; se ainda não escolheu, usa a padrão.
	sel, _ := configGet(db, "impressora")
	out.Selecionada = sel
	if out.Selecionada == "" {
		out.Selecionada = padrao
	}
	return out, nil
}

// SetImpressora grava a impressora escolhida para imprimir no banco (tabela
// config). Vazio = volta à padrão.
func (a *App) SetImpressora(nome string) error {
	db, err := a.dbOrErr()
	if err != nil {
		return err
	}
	if nome != "" {
		nomes, err := listaImpressoras()
		if err != nil {
			return err
		}
		ok := false
		for _, n := range nomes {
			if n == nome {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("impressora %q não encontrada no sistema", nome)
		}
	}
	return configSet(db, "impressora", nome)
}

// LarguraLinhaPadrao é a largura em caracteres usada quando o usuário ainda
// não configurou. Térmica de 58mm (MTP II) costuma ter 32 colunas.
const LarguraLinhaPadrao = 32

// GetLarguraLinha devolve o nº de caracteres por linha configurado para a
// impressão (banco) ou LarguraLinhaPadrao quando não configurado.
func (a *App) GetLarguraLinha() (int64, error) {
	db, err := a.dbOrErr()
	if err != nil {
		return LarguraLinhaPadrao, err
	}
	v, _ := configGet(db, "larguraLinha")
	if v == "" {
		return LarguraLinhaPadrao, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return LarguraLinhaPadrao, nil
	}
	return n, nil
}

// SetLarguraLinha grava o nº de caracteres por linha da impressão no banco.
func (a *App) SetLarguraLinha(n int64) error {
	if n < 1 || n > 200 {
		return fmt.Errorf("largura de linha inválida (1 a 200 caracteres): %d", n)
	}
	db, err := a.dbOrErr()
	if err != nil {
		return err
	}
	return configSet(db, "larguraLinha", strconv.FormatInt(n, 10))
}

// ---- Modo debug (não imprime; mostra o recibo na tela) ----

// GetModoDebug devolve se o app está em "modo debug": quando ligado, o recibo
// NÃO é enviado à impressora e aparece apenas na tela (mesma formatação da
// impressão). É uma preferência global do app (tabela config).
func (a *App) GetModoDebug() (bool, error) {
	db, err := a.dbOrErr()
	if err != nil {
		return false, err
	}
	v, _ := configGet(db, "modoDebug")
	return v == "1", nil
}

// SetModoDebug liga/desliga o "modo debug" (preferência global do app).
func (a *App) SetModoDebug(ligado bool) error {
	db, err := a.dbOrErr()
	if err != nil {
		return err
	}
	valor := "0"
	if ligado {
		valor = "1"
	}
	return configSet(db, "modoDebug", valor)
}

// impressoraEscolhida devolve o nome da impressora a usar para imprimir: a que
// o usuário escolheu (tabela config) ou, sem nada salvo, a padrão do sistema.
func impressoraEscolhida(db *sql.DB) (string, error) {
	nome, _ := configGet(db, "impressora")
	if nome != "" {
		return nome, nil
	}
	return impressoraPadrao()
}

// ImprimirTexto envia `texto` (recibo já formatado) à impressora configurada,
// como job ESC/POS cru (sem diálogo). Usado quando o modo debug está DESLIGADO.
func (a *App) ImprimirTexto(texto string) error {
	db, err := a.dbOrErr()
	if err != nil {
		return err
	}
	nome, err := impressoraEscolhida(db)
	if err != nil {
		return fmt.Errorf("não foi possível obter a impressora: %w", err)
	}
	if nome == "" {
		return fmt.Errorf("nenhuma impressora configurada — escolha uma na tela de Configurações")
	}
	return imprimirTexto(nome, texto)
}

// ---- migração do antigo config.json (pré-1.0: preferências fora do banco) ----

// configJsonLegado devolve o caminho do antigo arquivo de config, ao lado do banco.
func configJsonLegado() string {
	return filepath.Join(filepath.Dir(dbPath()), "config.json")
}

// importaConfigJsonLegado copia as chaves do antigo config.json (se existir)
// para a tabela `config` do banco (só quando a chave ainda não está no banco) e
// então apaga o arquivo. Roda uma única vez no startup, migrando sem perder a
// impressora/largura já configuradas. Erros de leitura são ignorados (o app
// segue com os defaults).
func importaConfigJsonLegado(db *sql.DB) error {
	path := configJsonLegado()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // sem arquivo antigo: nada a migrar
		}
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil // arquivo corrompido: ignora
	}
	for chave, valor := range m {
		if valor == "" {
			continue
		}
		atual, _ := configGet(db, chave)
		if atual != "" {
			continue // banco já tem valor mais recente
		}
		_ = configSet(db, chave, valor)
	}
	_ = os.Remove(path) // migrou: não precisa mais do arquivo
	return nil
}
