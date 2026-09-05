// Tipos espelhando os modelos Go (JSON tags). Preços em centavos.

export interface Evento {
  id: number
  nome: string
  ativo: boolean
  criadoEm: string
}

export type EstoqueTipo = 'ilimitado' | 'limitado'

export interface Produto {
  id: number
  eventoId: number
  nome: string
  preco: number // centavos
  estoqueTipo: EstoqueTipo
  quantidade: number
  ativo: boolean
}

export interface ProdutoVenda {
  id: number
  nome: string
  preco: number
  estoqueTipo: EstoqueTipo
  quantidade: number
  esgotado: boolean
}

export interface PedidoItem {
  id: number
  produtoId: number
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
  criadoEm: string
  itens: PedidoItem[]
}
