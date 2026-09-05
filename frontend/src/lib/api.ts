import type { Evento, Pedido, PedidoItem, Produto, ProdutoVenda } from './types'

// Tipagem do objeto injetado pelo Wails no window.
declare global {
  interface Window {
    go: {
      main: {
        App: {
          Greet(name: string): Promise<string>
          ListEventos(): Promise<Evento[]>
          CreateEvento(nome: string): Promise<Evento>
          UpdateEvento(id: number, nome: string, ativo: boolean): Promise<void>
          DeleteEvento(id: number): Promise<void>
          ListProdutos(eventoID: number): Promise<Produto[]>
          ListProdutosVenda(eventoID: number): Promise<ProdutoVenda[]>
          CreateProduto(eventoID: number, nome: string, preco: number, estoqueTipo: string, quantidade: number): Promise<Produto>
          UpdateProduto(id: number, nome: string, preco: number, estoqueTipo: string, quantidade: number): Promise<void>
          SetProdutoAtivo(id: number, ativo: boolean): Promise<void>
          DeleteProduto(id: number): Promise<void>
          CriarPedido(eventoID: number): Promise<Pedido>
          GetPedido(id: number): Promise<Pedido>
          FecharPedido(pedidoID: number, itens: PedidoItem[]): Promise<Pedido>
        }
      }
    }
  }
}

export const api = () => window.go.main.App

/** Extrai a mensagem de erro amigável de uma promise rejeitada do Wails. */
export function errMsg(e: unknown): string {
  return (e as Error)?.message ?? String(e)
}
