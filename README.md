# Vendinha

Ponto de venda (PDV) para eventos, em **pt-BR**. Para cada evento você cadastra
produtos — com estoque **infinito** ou **limitado** — e faz vendas em uma tela
touch-friendly. Suporta três formas de pagamento (dinheiro, cartão e **"anota
aí"**, o fiado) e, por evento, a venda de **cartelas** (raspadinha).

> O app é um aplicativo **desktop** Wails. A impressão em **térmica MTP II**
> ainda **não está implementada** — por ora a ficha/recibo só é exibida numa
> janela/simulação em modo debug.

## Stack

- **Wails v2** (Go backend, `v2.15`) + **SQLite** puro (`modernc.org/sqlite`,
  sem CGO), uma única conexão (`db.SetMaxOpenConns(1)`)
- **Frontend:** Vue 3.5 + Vite 6 + Nuxt UI v4 (standalone, fora do Nuxt) +
  vue-router (hash) + vuedraggable (SortableJS)
- Dados locais em `%APPDATA%\vendinha\vendinha.db`
- Dinheiro é sempre tratado como **centavos** (`int`), nunca float

## Modelo de dados

- `eventos` — um evento onde ocorrem vendas; pode habilitar a venda de cartelas
  (`vende_cartela`)
- `grupos` — categorias de um evento (ex.: Comida, Bebida, Sobremesa), cada uma
  com uma **cor** para o PDV e uma **ordem** de exibição. `grupo_id 0` = "Sem grupo"
- `produtos` — itens à venda; `estoque_tipo ∈ 'ilimitado' | 'limitado'` com
  `quantidade`. Pertencem a um grupo, têm `ordem` e um **atalho de teclado 1–9**
  (único por evento) que o adiciona no PDV
- `pedidos` / `pedido_itens` — vendas fechadas; nome/preço são **snapshot** no
  momento da venda. Ao fechar, o estoque dos produtos limitados é **baixado**
- `contas` — donos do fiado ("anota aí"), únicos por evento (case-insensitive)
- **Cartelas** — tipo próprio (não produto), ilimitadas; denominações R$ 10/20/50/100
  com layout ASCII embutido em `cartelas/*.txt`

## Funcionalidades

- **Formas de pagamento:** dinheiro (com troco), cartão e **anota aí** (lança a
  venda na conta de um dono do mesmo evento e vira sua pendência).
- **Contas / fiado:** tela "Anota aí" lista pendências por conta com **quitação**
  (baixa a conta inteira, reabre automaticamente em nova venda).
- **Pedidos:** histórico paginado dos fechados, com detalhes e **cancelamento**
  (devolve estoque limitado e remove o débito do fiado; cancelados saem de todo
  relatório).
- **Dashboard** por evento: receita, nº de pedidos/produtos e vendas por hora.
- **Cartelas (raspadinha):** opcional por evento, ilimitadas, impressas sob
  demanda — não baixam estoque.
- **Atalhos no PDV:** `1–9` adiciona produto, `Shift+1–9` remove; no diálogo de
  pagamento `1` = Dinheiro, `2` = Cartão, `3` = Anota aí; `Enter` = Fechar pedido.

## Telas / rotas

- `/eventos` — CRUD de eventos (gerenciamento)
- `/eventos/:eventoId` — **hub** do evento com 4 abas:
  - **Dashboard** (padrão): receita e vendas
  - **Produtos**: CRUD de grupos e produtos, reordenação por drag & drop e atalhos
  - **Anota aí**: contas do fiado e pendências
  - **Pedidos**: histórico e cancelamento
- `/pdv/:eventoId` — tela de venda (touchscreen), produtos agrupados por cor de um
  lado, pedido do outro; fechar pedido exibe o recibo/ficha (debug, sem impressora).

## Comandos

- `go test ./...` — testa a camada de dados (banco em memória), incl. baixa de estoque
- `cd frontend && npx vue-tsc --noEmit` — type-check do frontend (seguro durante o dev)
- `wails dev` — desenvolvimento com hot-reload do frontend (back `.go` **não**
  recarrega sozinho; reinicie o dev após mudar Go)
- `wails build` — gera o executável

> Impressão real (MTP II / ESC-POS) virá em iterações futuras.
