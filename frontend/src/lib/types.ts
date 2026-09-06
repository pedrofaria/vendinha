// Tipos espelhando os modelos Go (JSON tags). Preços em centavos.

export interface Evento {
  id: number
  nome: string
  ativo: boolean
  vendeCartela: boolean
  criadoEm: string
}

// Cartela (raspadinha) vendida por valor; ilimitada (imprime sob demanda).
// conteudo é o layout ASCII a imprimir (cartelas/<reais>.txt).
export interface Cartela {
  reais: number
  preco: number // centavos = reais * 100
  nome: string
  conteudo: string
}

export type EstoqueTipo = 'ilimitado' | 'limitado'

// 0 = "Sem grupo" (produto sem grupo_id). Os grupos reais têm id > 0.
export const SemGrupo = 0

export interface Grupo {
  id: number
  eventoId: number
  nome: string
  cor: string
  ordem: number
}

export interface Produto {
  id: number
  eventoId: number
  nome: string
  preco: number // centavos
  estoqueTipo: EstoqueTipo
  quantidade: number
  ativo: boolean
  grupoId: number // 0 = sem grupo
  atalho: number // 0 = sem atalho; tecla 1-9 no PDV
  ordem: number
}

// Grupo com produtos dentro, para a tela de gerenciamento. GrupoProdutos.id
// == SemGrupo para o bucket virtual "Sem grupo".
export interface GrupoProdutos extends Grupo {
  produtos: Produto[]
}

export interface ProdutoVenda {
  id: number
  nome: string
  preco: number
  estoqueTipo: EstoqueTipo
  quantidade: number
  esgotado: boolean
  atalho: number // 0 = sem atalho; tecla 1-9
}

// Grupo com produtos à venda, como o PDV renderiza.
export interface GrupoVenda extends Grupo {
  produtos: ProdutoVenda[]
}

// Conta ("Anota aí" / fiado) de um evento: quem deve na venda anotada. O nome é
// único por evento (case-insensitive); nomes diferentes são contas distintas.
// totalPendente = soma dos totais dos pedidos anotaai da conta (centavos).
export interface Conta {
  id: number
  eventoId: number
  nome: string
  totalPendente: number
  criadoEm: string
}

// Saldo de uma conta na tela "Anota aí": o que ela deve em aberto (vendas
// anotadas não quitadas), quantas vendas em aberto e quanto já quitou.
export interface ContaSaldo {
  id: number
  eventoId: number
  nome: string
  totalPendente: number // centavos ainda devidos
  numAberto: number // nº de vendas anotadas em aberto
  totalQuitado: number // centavos já quitados
  criadoEm: string
}

// Relatório do Dashboard de um evento.
export interface ResumoEvento {
  receitaTotal: number // soma dos totais dos pedidos fechados (centavos)
  numPedidos: number // nº de vendas fechadas
  numProdutosVendidos: number // unidades de produto vendidas (sem cartelas)
  vendasInicioHora: number // hora (0-23) em que o eixo do gráfico começa (virada de meia-noite)
  vendasPorHora: HoraVendas[] // histograma por hora do dia (0-23)
}

export interface HoraVendas {
  hora: number
  vendas: number
}

export interface PedidoItem {
  id: number
  produtoId: number
  cartelaReais: number // > 0 quando a linha é uma cartela (0 = produto)
  nome: string
  precoUnit: number
  qtd: number
  subtotal: number
}

export interface Pedido {
  id: number
  eventoId: number
  numero: number
  total: number
  forma: string // 'dinheiro' | 'cartao' | 'anotaai'
  contaId: number // > 0 quando a venda foi anotada (forma 'anotaai')
  quitadoEm: string // '' = em aberto; preenchido quando a venda anotada foi paga
  canceladoEm: string // '' = ativo; preenchido quando o pedido foi cancelado
  criadoEm: string
  itens: PedidoItem[]
}

// Linha leve da listagem de pedidos: traz o nome da conta do "Anota aí"
// (contaNome) e os status, sem os itens (carregados via GetPedido).
export interface PedidoResumo {
  id: number
  eventoId: number
  numero: number
  total: number
  forma: string
  contaId: number
  contaNome: string // preenchido quando a forma é 'anotaai'
  quitadoEm: string
  canceladoEm: string
  criadoEm: string
}

// Resposta paginada da listagem de pedidos (mais recentes primeiro).
export interface ListaPedidos {
  pedidos: PedidoResumo[]
  total: number
  pagina: number
  totalPaginas: number
}

// Impressoras disponíveis no SO + a selecionada para imprimir (tela
// Configurações). padrao = impressora padrão do Windows; selecionada = a salva
// no config.json ou, sem nada salvo, a padrão do sistema.
export interface ImpressorasInfo {
  nomes: string[]
  padrao: string
  selecionada: string
}
