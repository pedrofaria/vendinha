import type { Cartela, Conta, ContaSaldo, Evento, Grupo, GrupoProdutos, GrupoVenda, ListaPedidos, Pedido, PedidoItem, Produto, ResumoEvento } from './types'

// Tipagem do objeto injetado pelo Wails no window.
declare global {
  interface Window {
    go: {
      main: {
        App: {
          Greet(name: string): Promise<string>
          ListEventos(): Promise<Evento[]>
          CreateEvento(nome: string, vendeCartela: boolean): Promise<Evento>
          UpdateEvento(id: number, nome: string, ativo: boolean, vendeCartela: boolean): Promise<void>
          DeleteEvento(id: number): Promise<void>
          ListCartelas(): Promise<Cartela[]>
          ListGrupos(eventoID: number): Promise<Grupo[]>
          CreateGrupo(eventoID: number, nome: string, cor: string): Promise<Grupo>
          UpdateGrupo(id: number, nome: string, cor: string): Promise<void>
          DeleteGrupo(id: number): Promise<void>
          MoverGrupo(id: number, delta: number): Promise<void>
          ReorderGrupos(eventoID: number, ids: number[]): Promise<void>
          ListProdutos(eventoID: number): Promise<GrupoProdutos[]>
          ListProdutosVenda(eventoID: number): Promise<GrupoVenda[]>
          CreateProduto(eventoID: number, nome: string, preco: number, estoqueTipo: string, quantidade: number, grupoId: number, atalho: number): Promise<Produto>
          UpdateProduto(id: number, nome: string, preco: number, estoqueTipo: string, quantidade: number, grupoId: number, atalho: number): Promise<void>
          SetProdutoAtivo(id: number, ativo: boolean): Promise<void>
          DeleteProduto(id: number): Promise<void>
          MoverProduto(id: number, delta: number): Promise<void>
          ReorderProdutos(ids: number[]): Promise<void>
          CriarPedido(eventoID: number): Promise<Pedido>
          ListPedidos(eventoID: number, pagina: number, porPagina: number, numeroBusca: number): Promise<ListaPedidos>
          CancelarPedido(pedidoID: number): Promise<void>
          ListContas(eventoID: number): Promise<Conta[]>
          GetPedido(id: number): Promise<Pedido>
          FecharPedido(pedidoID: number, itens: PedidoItem[], forma: string, contaNome: string): Promise<Pedido>
          ListContasSaldo(eventoID: number): Promise<ContaSaldo[]>
          QuitarConta(contaID: number): Promise<void>
          ListPedidosConta(contaID: number): Promise<Pedido[]>
          ResumoEvento(eventoID: number): Promise<ResumoEvento>
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
