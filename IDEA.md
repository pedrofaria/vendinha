# Vendinha — a ideia

## Origem

> No geral, será um ponto de vendas bem simples e impressão do que foi comprado
> em uma impressora MTP II térmica.

Essa foi a semente: um **PDV para eventos** — barraquinhas de igreja, festas,
feiras — rodando como app **desktop**, com venda rápida em tela touch e emissão
de um comprovante impresso em **térmica MTP II (ESC-POS)**.

De lá pra cá o app virou muito mais que isso, mantendo o espírito de "simples e
prático pra quem está atrás do balcão".

## O que o Vendinha é hoje

Um **ponto de venda por evento**, em **pt-BR**. Para cada evento você cadastra
os produtos que vai vender, organiza-os em grupos e faz as vendas numa tela
touch-friendly. Guarda histórico, controla estoque, lida com fiado e até com
venda de **cartelas (raspadinha)**.

É 100% local: os dados moram num SQLite em `%APPDATA%\vendinha\vendinha.db`.

## Visão completa (por pilar)

### 1. Venda rápida no caixa (PDV)
- Tela touch pensada pra quem atende: produtos agrupados e **coloridos por
  grupo**, com o pedido corrente do lado.
- **Estoque** por produto: **infinito** (quantidade irrelevante) ou **limitado**
  (com baixa automática ao fechar o pedido e aviso de "N restantes"/"Esgotado").
- **Atalhos de teclado** pra acelerar: `1–9` adiciona o produto, `Shift+1–9`
  remove; no diálogo de pagamento `1`/`2`/`3` escolhem a forma e `Enter` fecha.
- Três formas de pagamento:
  - **Dinheiro** — informa quanto recebeu e calcula o **troco**.
  - **Cartão** — marca a venda como paga.
  - **Anota aí** (fiado) — lança a venda na conta de um dono e vira sua pendência.
- Snapshot de nome e preço de cada item no momento da venda (histórico fiel).

### 2. Gestão do evento
- O evento virou um **hub** com 4 abas:
  - **Dashboard** — receita total, nº de pedidos, unidades vendidas e vendas por
    hora.
  - **Produtos** — CRUD de grupos e produtos, com reordenação por **drag &
    drop** e atribuição dos atalhos 1–9.
  - **Anota aí** — contas do fiado e pendências, com **quitação**.
  - **Pedidos** — histórico paginado e **cancelamento**.
- **Pedidos:** histórico dos fechados, com detalhes por pedido e **cancelamento**
  que devolve o estoque limitado e remove o débito do fiado. Cancelado some de
  toda soma e relatório.

### 3. Fiado ("Anota aí")
- Contas (donos) **únicas por evento**, sem diferenciar maiúsculas.
- A venda anotada entra na conta e vira **pendência**.
- Quitação é da **conta inteira** de uma vez; nova venda anotada reabre a conta
  como pendente sem perder o histórico já pago.

### 4. Cartelas (raspadinha)
- Tipo próprio (não é produto), **ilimitado** — imprime-se sob demanda.
- Denominações **R$ 10 / 20 / 50 / 100**, com layout do conteúdo embutido
  (ASCII) no app, como fonte única do que se imprime.
- Venda de cartelas é **opcional por evento** e **não baixa estoque**.

## Roadmap / o que ainda não está pronto

- **Impressão térmica (MTP II / ESC-POS) NÃO implementada.** Hoje a ficha/recibo
  só aparece numa simulação/janela em modo debug. Este é o próximo passo grande:
  imprimir de verdade o que foi comprado na térmica — é a origem do app.
- Baixa/quitação pontual de pedidos individuais do fiado (hoje a quitação é a
  conta inteira).

## Princípios que guiam o projeto

- **Dinheiro é sempre centavo (`int`)**, nunca float — troco, preço e total são
  aritmética inteira.
- **Simples primeiro**: nenhum campo/feature a mais do que a operação de evento
  precisa; a complexidade que existe (fiado, grupos, cancelamento) nasceu de
  necessidade real de uso.
- **Confiável offline**: dado local, sem depender de rede no balcão.
- **Baixa de estoque só no fechamento**; carrinho aberto é reserva local, nunca
  muta flag persistido de "esgotado".
- Todo o fluxo e as telas em **pt-BR**.

## Stack

- **Wails v2** — Go no back + SQLite puro (`modernc.org/sqlite`, sem CGO, uma
  única conexão).
- **Vue 3.5 + Vite 6 + Nuxt UI v4** (standalone) + vue-router (hash).
- App desktop nativo; rodar com `wails dev` e empacotar com `wails build`.
