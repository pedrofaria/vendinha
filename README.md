# Vendinha

Ponto de venda (PDV) para eventos. Para cada evento você cadastra produtos —
alguns com estoque **infinito**, outros com quantidade **limitada** — e faz vendas.
Após fechar um pedido, o app imprime uma **ficha** em impressora térmica (MTP II);
em modo debug, mostra o recibo numa janela de simulação.

## Stack
- **Wails v2** (Go backend) + **SQLite** puro (`modernc.org/sqlite`, sem CGO)
- **Frontend:** Vue 3 + Vite + Nuxt UI v4 (standalone, fora do Nuxt) + vue-router (hash)
- Dados locais em `%APPDATA%\vendinha\vendinha.db`

## Modelo de dados
- `eventos` — um evento onde ocorrem vendas
- `produtos` — itens à venda; `estoque_tipo` ∈ `'ilimitado' | 'limitado'`, com `quantidade`
- `pedidos` / `pedido_itens` — vendas fechadas; nome e preço dos itens são **snapshot**
  no momento da venda (histórico preservado). Ao fechar, estoque dos produtos limitados é baixado.

## Comandos
- `go test ./...` — testa a camada de dados (banco em memória) incl. baixa de estoque.
- `cd frontend && npx vue-tsc --noEmit` — type-check do frontend.
- `wails build` — gera `build/bin/vendinha.exe`.
- `wails dev` — desenvolvimento com hot-reload (back Go não recarrega sozinho;
  reinicie o dev após mudar `.go`).

## Telas (esqueleto inicial)
- `/eventos` — CRUD de eventos (gerenciamento)
- `/eventos/:eventoId` — CRUD de produtos do evento (gerenciamento)
- `/pdv/:eventoId` — tela de venda: produtos de um lado, pedido do outro
  (touchscreen-friendly). Fechar pedido exibe o recibo em modo debug.

> Impressão real (MTP II / ESC-POS) e demais telas virão em iterações futuras.
