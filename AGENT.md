# AGENT.md — guia de trabalho no Vendinha

Guia para **agentes de IA / coding agents** que editam este repositório.
Stack: **Wails v2** (Go) + **SQLite puro** (`modernc.org/sqlite`, sem CGO) no
back; **Vue 3.5 + Vite 6 + Nuxt UI v4 (standalone) + vue-router (hash)** no
frontend, em `frontend/src/`. App desktop; interface e commits em **pt-BR**
(`feat:`/`fix:`/`docs:` convencionais).

## Arquitetura (em 10 linhas)

- Backend Go (todos `.go` na raiz): `main.go` embute `frontend/dist`; `app.go`
  abre o banco; `bindings.go` expõe os métodos chamáveis pelo frontend (um método
  `func (a *App) X(...)` = um binding); `repo.go` tem toda a lógica SQL;
  `models.go` os tipos; `db.go` o schema + migrações idempotentes.
- Frontend: `views/` = telas, `lib/` = helpers (`format.ts` dinheiro,
  `api.ts` bindings tipados, `types.ts`), `router/` rotas hash.
- Rotas: `/eventos` → `/eventos/:eventoId` (**hub** `EventoShell` com 4 abas
  filhas: Dashboard/Produtos/Anota aí/Pedidos) → `/pdv/:eventoId`.
- Banco local: `%APPDATA%\vendinha\vendinha.db`, **uma única conexão**
  (`db.SetMaxOpenConns(1)`).
- Bindings `frontend/wailsjs/go/...` são **auto-regenerados** pelo `wails dev`
  quando o `.go` muda — não edite à mão.

## Regras de trabalho (leia antes de editar)

- **Não lance `wails dev` por conta própria.** O usuário inicia/reinicia. Edições
  `.vue` fazem hot-reload via Vite; edições `.go` **não** recarregam num `wails dev`
  rodando — precisam de restart (decisão do usuário).
- Verificação sem mexer no dev: `go build ./... && go test ./...` (back) e
  `cd frontend && npx vue-tsc --noEmit` (front; não toca o watcher do Vite).
- Comit/`git add` durante um `wails dev` ativo pode derrubar o Vite (watcher
  storm) — pausar o dev antes de mexer no git quando for possível.
- Nunca simule execução: teste o que mudou e reporte o resultado real.

## Invariantes críticos

- **Dinheiro é centavo int**, nunca float. Use `frontend/src/lib/format.ts`
  (`money`, `centsToInput`, `inputToCents`) e `format.go` no back. Sempre.
- **Estoque só é debitado no fechamento** (`FecharPedido`). Carrinho aberto é
  reserva LOCAL no frontend: disponibilidade deriva do carrinho (`disponivel(p) =
  p.quantidade - qtdNoCarrinho(p.id)`, `esgotado = disponivel <= 0`). **Nunca**
  mutar um flag `esgotado` persistido para "devolver" estoque — derive do carrinho.
- **Conexão única → deadlock:** nunca abra um 2º `r.db.Query` enquanto um `rows`
  do mesmo `r.db` ainda estiver aberto no loop (a 2ª espera a 1ª). Materialize tudo
  do 1º SELECT num slice e feche `rows.Close()` **antes** de consultar o que depende
  dele. Padrão correto: `ListPedidosConta`.
- **`grupo_id`/`cartela_reais`/`quitado_em`/`atalho` são `NULL`-able**: escaneie-os
  como `sql.NullInt64`/`sql.NullString` — Scan de NULL em `int64`/`string` quebra.
- **Pedido cancelado sai de TODA agregação** (somas, relatórios, pendências): filtre
  `p.cancelado_em IS NULL` em toda query de soma/listagem. Cancelamento em TX única,
  devolvendo estoque limitado (materialize+close antes dos UPDATEs).
- **`UNIQUE (evento_id, nome COLLATE NOCASE)`** no fiado: a **coluna** também deve
  ser `COLLATE NOCASE` — se NOCASE estiver só no índice, o lookup `WHERE nome=?`
  fica BINARY e o INSERT bate no UNIQUE.

## Convenções de domínio

- `estoque_tipo`: `'ilimitado'` (quantidade irrelevante) | `'limitado'`.
  Produto inativo (`ativo=0`) sai da venda.
- Formas de pagamento (`pedidos.forma`): `'dinheiro' | 'cartao' | 'anotaai'`.
  `anotaai` grava `conta_id` (dono do fiado, do mesmo evento); as demais NULL.
  Troco/recebido ficam só no frontend + recibo (sem coluna).
- Cartela (raspadinha) = linha própria, NÃO produto (`produto_id NULL`,
  `cartela_reais = 10/20/50/100`, **não debita estoque**). No carrinho a linha tem
  `key` string `'p<produtoId>'` ou `'c<reais>'` (nunca indexar por produtoId: cartela
  tem produtoId 0 e colidiria). Catálogo fixo em `cartelas.go` com conteúdo ASCII via
  `//go:embed cartelas/*.txt` (fonte única).
- Atalho teclado `1–9` por produto: único **por evento** (ativos e inativos),
  `0` = sem atalho (DB NULL). Backstop = índice único parcial
  `(evento_id, atalho) WHERE atalho IS NOT NULL`; **DROP+CREATE** a cada abertura
  (um `IF NOT EXISTS` não conserta um índice antigo com definição errada).
- Quitar fiado = **conta inteira** (`quitado_em` em todos os pedidos anotaai
  abertos); nova venda anotada reabre a conta.
- Snapshot de nome/preço do item na venda preserva histórico.

## Navegação (frontend)

- Contrato: qualquer botão de voltar/voltar-ao-evento saindo do PDV mira
  `/eventos/:eventoId` (o hub), **nunca** a lista `/eventos` (só alcançável pelo
  topo/nav). A lista `/eventos` é alcançada do topo.
- "Fechar pedido" no PDV só abre o diálogo de pagamento — o registro de venda é
  feito por "Finalizar compra" (`CriarPedido` + `FecharPedido`). Dinheiro: trava
  "Finalizar" até `recebido >= total`. Cancelar/Limpar carrinho pede confirmação.
- Em `anotaai` use o seletor de conta (palette); `podeFinalizar` exige conta
  escolhida (nunca criar por engano). Reset do estado no `abrirPagamento`.
- Teclado no PDV: `1–9`/`Shift+1–9` = add/remove produto (sem foco em campo);
  `Enter` = Fechar pedido / avançar no diálogo; com diálogo aberto `1/2/3` =
  dinheiro/cartão/anota aí. PITFALL Windows: Shift+NumLock inverte o Numpad — mapeie
  os keyCodes de navegação (End/Home/PgUp/PgDn/arrows) de volta à tecla.

## Migrações (sem framework)

Não há sistema de migração. `db.go` roda o schema (CREATE IF NOT EXISTS) a cada
abertura e aplica colunas novas via `ensureColumn`/`PRAGMA table_info` + `ALTER
TABLE`, tudo idempotente. Coluna nova → siga esse padrão (função `migrateX`) e
mantenha o schema limpo (CREATE TABLE) em dia para bancos novos. Mexer numa
assinatura de método muda em cascata: repo + bindings + tests + `api.ts`.
