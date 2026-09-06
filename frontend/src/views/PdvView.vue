<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { centsToInput, inputToCents, money } from '../lib/format'
import type { Cartela, Conta, Evento, GrupoVenda, Pedido, PedidoItem, ProdutoVenda } from '../lib/types'

const route = useRoute()
const router = useRouter()
const eventoId = computed(() => Number(route.params.eventoId))

const evento = ref<Evento | null>(null)
// Produtos agrupados (ordem de exibição do PDV). Grupos vazios não vêm.
const grupos = ref<GrupoVenda[]>([])
// Visão "achatada" para as operações do carrinho (buscar produto por id etc.).
const todos = computed(() => grupos.value.flatMap((g) => g.produtos))
const loading = ref(false)
const error = ref('')

// Cartelas disponíveis (catálogo fixo) quando o evento vende cartelas.
const cartelas = ref<Cartela[]>([])

// Pedido em montagem (cliente local). Uma linha é de produto (produtoId > 0,
// cartelaReais 0) OU de cartela (cartelaReais > 0, produtoId 0). `key` é a
// identidade única no carrinho (p<id> p/ produto, c<reais> p/ cartela).
interface LinhaCarrinho {
  key: string
  produtoId: number
  cartelaReais: number
  nome: string
  preco: number
  qtd: number
}
const carrinho = ref<LinhaCarrinho[]>([])
const buscando = ref<string>('')

const chaveProduto = (id: number) => `p${id}`
const chaveCartela = (reais: number) => `c${reais}`

// Resultado do pedido fechado + modal de recibo/impressão
const showReceipt = ref(false)
const pedidoFechado = ref<Pedido | null>(null)

// Dialog de forma de pagamento antes de fechar de fato
type FormaPagamento = 'dinheiro' | 'cartao' | 'anotaai'
const showPagamento = ref(false)
const forma = ref<FormaPagamento | null>(null)
const recebidoText = ref('')
const showConfirmCancelar = ref(false)
const confirmarDePagamento = ref(false)
const pagamentoFechado = ref<{ forma: FormaPagamento; recebido: number; troco: number; contaNome?: string } | null>(null)

// "Anota aí" (fiado): o dono da conta é escolhido via CommandPalette. Contas
// existentes do evento aparecem na lista; se o texto digitado não casa com
// nenhuma, uma opção "Criar conta …" permite criar. O backend reusa a conta
// existente pelo nome exato (case-insensitive) ou cria na finalização.
const contas = ref<Conta[]>([])
const contaBusca = ref('') // termo digitado no CommandPalette (v-model:search-term)
const contaEscolhida = ref<{ nome: string; nova: boolean } | null>(null)

// Grupos do CommandPalette: contas existentes + (quando o texto não casa com uma
// existente) um item "Criar conta …" para cadastrar o novo dono.
const contaGrupos = computed<any[]>(() => {
  const grupos: any[] = []
  if (contas.value.length) {
    grupos.push({
      id: 'contas',
      label: 'Contas existentes deste evento',
      items: contas.value.map((c) => ({
        id: c.id,
        label: c.nome,
        slot: 'conta', // template #conta desenha nome à esquerda e total à direita
        valor: c.totalPendente,
        onSelect: () => escolherConta(c.nome, false)
      }))
    })
  }
  const q = contaBusca.value.trim()
  const casaExato = q !== '' && contas.value.some((c) => c.nome.toLowerCase() === q.toLowerCase())
  if (q && !casaExato) {
    grupos.push({
      id: 'nova',
      ignoreFilter: true, // não passa pelo fuse: mostra quando o texto não é uma conta exata
      items: [{ label: `Criar conta "${q}"`, icon: 'i-lucide-user-plus', onSelect: () => escolherConta(q, true) }]
    })
  }
  return grupos
})

function escolherConta(nome: string, nova: boolean) {
  contaEscolhida.value = { nome: nome.trim(), nova }
  contaBusca.value = ''
}

function limparConta() {
  contaEscolhida.value = null
  contaBusca.value = ''
}

const recebidoCents = computed(() => inputToCents(recebidoText.value))
const troco = computed(() => Math.max(0, recebidoCents.value - totalCarrinho.value))
const trocoSuficiente = computed(() => recebidoCents.value >= totalCarrinho.value)

const recebidoInput = ref<unknown>(null)

// foca um UInput após o próximo tick (o ref expõe inputRef nativo).
function focarInput(r: unknown) {
  nextTick(() => {
    const c = r as { inputRef?: unknown } | null
    const el = c?.inputRef
    const target =
      el && typeof (el as { focus?: unknown }).focus === 'function'
        ? el
        : (el as { value?: unknown } | null)?.value
    ;(target as { focus?: () => void } | null)?.focus?.()
  })
}

async function loadContas() {
  try {
    contas.value = await api().ListContas(eventoId.value)
  } catch {
    contas.value = []
  }
}

function selecionarForma(f: FormaPagamento) {
  forma.value = f
  // em dinheiro, joga o cursor direto no campo de valor recebido. Em "Anota aí",
  // o CommandPalette autofocusa o campo de busca quando é montado (v-if abaixo).
  if (f === 'dinheiro') {
    focarInput(recebidoInput.value)
  }
}

function podeFinalizar(): boolean {
  if (buscando.value) return false
  if (!forma.value) return false
  if (forma.value === 'dinheiro') return trocoSuficiente.value
  if (forma.value === 'anotaai') return !!contaEscolhida.value
  return true
}

async function abrirPagamento() {
  if (carrinho.value.length === 0) return
  forma.value = null
  recebidoText.value = ''
  contaEscolhida.value = null
  contaBusca.value = ''
  await loadContas() // autocomplete atualizado (contas criadas há pouco aparecem)
  showPagamento.value = true
}

function finalizarCompra() {
  const f = forma.value
  if (!f) return
  if (f === 'dinheiro' && !trocoSuficiente.value) {
    error.value = 'Valor recebido menor que o total'
    return
  }
  showPagamento.value = false
  fecharPedido(f)
}

function cancelarDaTela() {
  confirmarDePagamento.value = false
  showConfirmCancelar.value = true
}
function cancelarDoPagamento() {
  confirmarDePagamento.value = true
  showPagamento.value = false
  showConfirmCancelar.value = true
}
function voltarSemCancelar() {
  showConfirmCancelar.value = false
  if (confirmarDePagamento.value) showPagamento.value = true
}
function confirmarCancelar() {
  showConfirmCancelar.value = false
  clearCart()
}

const totalCarrinho = computed(() =>
  carrinho.value.reduce((s, l) => s + l.qtd * l.preco, 0)
)
const itensCount = computed(() => carrinho.value.reduce((s, l) => s + l.qtd, 0))

// Disponibilidade é sempre derivada do carrinho (reserva local; o banco só baixa
// no fechamento). Remover/decrementar um item "devolve" o estoque ao grid.
function qtdNoCarrinho(produtoId: number): number {
  return carrinho.value.find((l) => l.produtoId === produtoId)?.qtd ?? 0
}
function disponivel(p: ProdutoVenda): number {
  if (p.estoqueTipo !== 'limitado') return Infinity
  return p.quantidade - qtdNoCarrinho(p.id)
}
function esgotado(p: ProdutoVenda): boolean {
  return disponivel(p) <= 0
}

function acharLinha(key: string): LinhaCarrinho | undefined {
  return carrinho.value.find((l) => l.key === key)
}

function addToCart(p: ProdutoVenda) {
  if (esgotado(p)) return
  let linha = acharLinha(chaveProduto(p.id))
  if (!linha) {
    linha = {
      key: chaveProduto(p.id),
      produtoId: p.id,
      cartelaReais: 0,
      nome: p.nome,
      preco: p.preco,
      qtd: 0
    }
    carrinho.value.push(linha)
  }
  if (linha.qtd < p.quantidade || p.estoqueTipo !== 'limitado') linha.qtd++
}

// Adiciona 1 cartela ao carrinho. Cartela é ilimitada (imprime sob demanda).
function addCartela(c: Cartela) {
  let linha = acharLinha(chaveCartela(c.reais))
  if (!linha) {
    linha = {
      key: chaveCartela(c.reais),
      produtoId: 0,
      cartelaReais: c.reais,
      nome: c.nome,
      preco: c.preco,
      qtd: 0
    }
    carrinho.value.push(linha)
  }
  linha.qtd++
}

function incQty(linha: LinhaCarrinho) {
  // cartela é ilimitada: sempre cabe +1
  if (linha.cartelaReais > 0) {
    linha.qtd++
    return
  }
  const p = todos.value.find((x) => x.id === linha.produtoId)
  if (!p || esgotado(p)) return
  linha.qtd++
}

// Adiciona o produto cujo atalho é a tecla digitada (1-9). Produto esgotado
// não entra (addToCart já no-op); produto inativo nem chega a `todos`.
function addPorTecla(tecla: number) {
  const p = todos.value.find((x) => x.atalho === tecla)
  if (p) addToCart(p)
}

// Shift+tecla retira 1 unidade do produto do pedido (remove a linha se era a última).
function decPorTecla(tecla: number) {
  const p = todos.value.find((x) => x.atalho === tecla)
  if (!p) return
  const linha = carrinho.value.find((l) => l.produtoId === p.id)
  if (linha) decQty(linha)
}

// Devolve a tecla 1-9 pressionada, independente da origem (fileira superior ou
// Numpad). Alguns WebView reportam o Numpad com code vazio/'Unidentified', então
// aceita também o e.key quando é um dígito. Retorna null se não for 1-9.
function teclaDigitada(e: KeyboardEvent): number | null {
  const dm = /^Digit([1-9])$/.exec(e.code) // fileira superior (estável mesmo com Shift)
  if (dm) return Number(dm[1])
  const nm = /^Numpad([1-9])$/.exec(e.code) // numpad por code
  if (nm) return Number(nm[1])
  if (/^[1-9]$/.test(e.key)) return Number(e.key) // fallback (cobre numpad com NumLock ligado)
  // Windows "inverte" o Numpad ao segurar Shift com NumLock ligado: o dígito chega
  // como função de navegação (Numpad1->End, Numpad7->Home...). Mapeia de volta.
  if (e.shiftKey) {
    const navParaTecla: Record<number, number> = { 33: 9, 34: 3, 35: 1, 36: 7, 37: 4, 39: 6, 38: 8, 40: 2 }
    const tecla = navParaTecla[e.keyCode]
    if (tecla !== undefined) return tecla
  }
  return null
}

// Campo de texto/editable (onde Enter deve digitar, não agir como comando).
function editavel(el: HTMLElement | null): boolean {
  return !!el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.tagName === 'SELECT' || el.isContentEditable)
}

// Elemento interativo (botão/link): Enter nativo ativa o foco — não duplicar aqui.
function interativo(el: HTMLElement | null): boolean {
  return !!el && !!el.closest('button, a, input, textarea, select, [role="button"], [contenteditable="true"]')
}

// Atalhos de teclado no PDV:
//  - tecla 1-9: adiciona o produto ao pedido; Shift+tecla retira 1;
//  - Enter: "Fechar pedido" — abre o pagamento na tela, finaliza dentro do
//    diálogo de pagamento, e confirma/fecha nos modais de confirmação/recibo.
function onGlobalKey(e: KeyboardEvent) {
  if (e.repeat) return
  if (e.ctrlKey || e.metaKey || e.altKey) return
  const t = e.target as HTMLElement | null

  if (e.key === 'Enter') {
    // no campo "valor recebido" (pagamento em dinheiro): Enter = finalizar.
    // NO campo de busca do CommandPalette (anota aí) o Enter deve selecionar o
    // item destacado (reka lida com o evento) — não finalizar direto.
    if (showPagamento.value && editavel(t)) {
      if (forma.value === 'dinheiro' && podeFinalizar()) {
        e.preventDefault()
        finalizarCompra()
      }
      return
    }
    // botão/link focado: deixa o Enter nativo ativar o elemento
    if (interativo(t)) return
    if (showPagamento.value) {
      if (podeFinalizar()) {
        e.preventDefault()
        finalizarCompra()
      }
      return
    }
    if (showConfirmCancelar.value) {
      e.preventDefault()
      confirmarCancelar()
      return
    }
    if (showReceipt.value) {
      e.preventDefault()
      showReceipt.value = false
      return
    }
    // tela principal do PDV: Enter = "Fechar pedido"
    if (!buscando.value && carrinho.value.length) {
      e.preventDefault()
      abrirPagamento()
    }
    return
  }

  // dígitos: no diálogo de pagamento, 1 = Dinheiro, 2 = Cartão e 3 = Anota aí;
  // fora dele, 1-9 adiciona/remove produto. Nunca engole digitação em campo.
  if (editavel(t)) return
  const tecla = teclaDigitada(e)
  if (tecla === null) return

  if (showPagamento.value) {
    if (tecla === 1) selecionarForma('dinheiro')
    else if (tecla === 2) selecionarForma('cartao')
    else if (tecla === 3) selecionarForma('anotaai')
    return
  }
  if (showConfirmCancelar.value || showReceipt.value || !!buscando.value) return
  if (e.shiftKey) decPorTecla(tecla)
  else addPorTecla(tecla)
}

function decQty(linha: LinhaCarrinho) {
  if (linha.qtd <= 1) {
    removeLine(linha)
    return
  }
  linha.qtd--
}

function removeLine(linha: LinhaCarrinho) {
  carrinho.value = carrinho.value.filter((l) => l.key !== linha.key)
  // esgotado do grid é recalculado automaticamente (estoque "devolvido")
}

function clearCart() {
  carrinho.value = []
}

async function loadEvento() {
  try {
    const all = await api().ListEventos()
    evento.value = all.find((e) => e.id === eventoId.value) ?? null
  } catch (e) {
    error.value = errMsg(e)
  }
}

async function loadProdutos() {
  loading.value = true
  error.value = ''
  try {
    grupos.value = await api().ListProdutosVenda(eventoId.value)
    // aplica reserva já no carrinho, se houver (ex.: reload)
    carrinho.value = []
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

async function loadCartelas() {
  try {
    cartelas.value = await api().ListCartelas()
  } catch (e) {
    error.value = errMsg(e)
  }
}

// Cartelas vendidas no pedido fechado, para exibir o layout a imprimir
// (1 cópia por unidade da denominação).
const cartelasImprimir = computed(() => {
  const p = pedidoFechado.value
  if (!p) return []
  const porReais = new Map<number, number>()
  for (const it of p.itens) {
    if (it.cartelaReais > 0) porReais.set(it.cartelaReais, (porReais.get(it.cartelaReais) ?? 0) + it.qtd)
  }
  const out: { reais: number; nome: string; qtd: number; conteudo: string }[] = []
  for (const [reais, qtd] of porReais) {
    const c = cartelas.value.find((x) => x.reais === reais)
    if (!c) continue
    out.push({ reais, nome: c.nome, qtd, conteudo: c.conteudo })
  }
  // preserva a ordem do catálogo (10, 20, 50, 100)
  return out.sort((a, b) => a.reais - b.reais)
})

async function fecharPedido(forma: FormaPagamento) {
  if (carrinho.value.length === 0) return
  error.value = ''
  buscando.value = 'Finalizando…'
  // captura antes de qualquer await (carrinho ainda intacto)
  const recebido = forma === 'dinheiro' ? recebidoCents.value : totalCarrinho.value
  const trocoValor = forma === 'dinheiro' ? troco.value : 0
  const nomeConta = forma === 'anotaai' ? (contaEscolhida.value?.nome ?? '') : ''
  try {
    const pedido = await api().CriarPedido(eventoId.value)
    const itens: PedidoItem[] = carrinho.value.map((l) => ({
      id: 0,
      produtoId: l.produtoId,
      cartelaReais: l.cartelaReais,
      nome: l.nome,
      precoUnit: l.preco,
      qtd: l.qtd,
      subtotal: l.qtd * l.preco
    }))
    pedidoFechado.value = await api().FecharPedido(pedido.id, itens, forma, nomeConta)
    pagamentoFechado.value = { forma, recebido, troco: trocoValor, contaNome: nomeConta || undefined }
    showReceipt.value = true
    clearCart()
    await loadProdutos() // reflete estoque novo
    if (forma === 'anotaai') await loadContas() // conta nova (se criou) entra no autocomplete
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    buscando.value = ''
  }
}

// Receipt text formatado para impressão térmica (ficha).
function receiptText(): string {
  const p = pedidoFechado.value
  if (!p) return ''
  const lines: string[] = []
  lines.push('          VENDINHA')
  lines.push('--------------------------------')
  lines.push(`Evento: ${evento.value?.nome ?? ''}`)
  lines.push(`Ficha/Pedido: #${p.numero}`)
  lines.push(`Data: ${p.criadoEm}`)
  lines.push('--------------------------------')
  for (const it of p.itens) {
    lines.push(`${it.qtd}x ${it.nome}`)
    lines.push(`    ${money(it.subtotal)}`)
  }
  lines.push('--------------------------------')
  lines.push(`TOTAL: ${money(p.total)}`)
  const pg = pagamentoFechado.value
  if (pg) {
    if (pg.forma === 'dinheiro') {
      lines.push(`Pagamento: Dinheiro`)
      lines.push(`Recebido: ${money(pg.recebido)}`)
      if (pg.troco > 0) lines.push(`Troco: ${money(pg.troco)}`)
    } else if (pg.forma === 'anotaai') {
      lines.push('Pagamento: Anota aí')
      lines.push(`Conta: ${pg.contaNome ?? ''}`)
    } else {
      lines.push('Pagamento: Cartão')
    }
  }
  lines.push('')
  lines.push('Obrigado!')
  return lines.join('\n')
}

onMounted(() => {
  window.addEventListener('keydown', onGlobalKey)
  loadEvento()
  loadProdutos()
  loadCartelas()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onGlobalKey)
})
</script>

<template>
  <div class="flex h-full flex-col">
    <!-- barra do PDV -->
    <div class="flex items-center justify-between border-b border-neutral-200 bg-white px-4 py-2 dark:border-neutral-800 dark:bg-neutral-900">
      <div class="flex items-center gap-3">
        <UButton color="neutral" variant="ghost" icon="i-lucide-arrow-left" size="sm" @click="router.push(`/eventos/${eventoId}`)">
          Voltar ao evento
        </UButton>
        <div class="text-sm">
          <span class="font-medium">{{ evento?.nome ?? 'Venda' }}</span>
        </div>
      </div>
      <div class="text-sm text-neutral-500 dark:text-neutral-400">
        {{ itensCount }} {{ itensCount === 1 ? 'item' : 'itens' }} · {{ money(totalCarrinho) }}
      </div>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="m-3" />

    <div class="grid min-h-0 flex-1 grid-cols-[1fr_360px] gap-3 p-3">
      <!-- Produtos (esquerda) -->
      <div class="flex min-h-0 flex-col overflow-hidden rounded-xl border border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900">
        <div class="flex items-center justify-between border-b border-neutral-100 px-3 py-2 dark:border-neutral-800">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-package" class="size-4 text-neutral-400" />
            <span class="text-sm font-medium">Produtos</span>
          </div>
          <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="xs"
            label="Editar" @click="router.push(`/eventos/${eventoId}`)" />
        </div>
        <div class="min-h-0 flex-1 overflow-auto p-3">
          <!-- Cartelas (habilitadas no evento): ilimitadas, imprimem ao fechar o pedido -->
          <section v-if="evento?.vendeCartela && cartelas.length"
            class="mb-4 overflow-hidden rounded-xl border border-violet-200 bg-violet-50/70 dark:border-violet-900 dark:bg-violet-950/25">
            <div class="flex items-center gap-2 border-b border-violet-200/70 px-3 py-2 dark:border-violet-900">
              <UIcon name="i-lucide-ticket" class="size-4 text-violet-600 dark:text-violet-400" />
              <span class="text-sm font-semibold">Cartelas</span>
              <span class="ml-auto text-[11px] text-neutral-500 dark:text-neutral-400">imprime cada cartela ao fechar</span>
            </div>
            <div class="grid grid-cols-2 gap-2 p-3 sm:grid-cols-4">
              <button
                v-for="c in cartelas"
                :key="c.reais"
                type="button"
                class="flex min-h-[76px] flex-col items-center justify-center gap-0.5 rounded-xl border border-neutral-200 bg-white px-1 text-center transition hover:border-violet-400 dark:border-neutral-700 dark:bg-neutral-900 dark:hover:border-violet-500"
                :title="`Adicionar ${c.nome} ao pedido`"
                @click="addCartela(c)"
              >
                <span class="text-[11px] uppercase tracking-wide text-neutral-400">Cartela</span>
                <span class="text-lg font-bold text-violet-700 dark:text-violet-400">{{ money(c.preco) }}</span>
              </button>
            </div>
          </section>

          <div v-if="loading" class="flex justify-center py-10">
            <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
          </div>
          <div v-else-if="grupos.length === 0 && !(evento?.vendeCartela && cartelas.length)" class="py-10 text-center text-neutral-400">
            Nenhum produto disponível. Cadastre antes de vender.
          </div>
          <div v-else class="space-y-4">
            <!-- seção por grupo: cabeçalho com a cor de identificação -->
            <section v-for="g in grupos" :key="`grp-${g.id}`">
              <div class="mb-2 flex items-center gap-2">
                <span class="size-3 shrink-0 rounded-full ring-1 ring-black/10" :style="{ backgroundColor: g.cor }" />
                <h3 class="truncate text-xs font-semibold uppercase tracking-wide text-neutral-500 dark:text-neutral-400">{{ g.nome }}</h3>
              </div>
              <div class="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
                <button
                  v-for="p in g.produtos"
                  :key="p.id"
                  type="button"
                  class="relative flex min-h-[108px] flex-col overflow-hidden rounded-xl border p-3 pl-4 text-left transition
                    focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                  :class="esgotado(p)
                    ? 'cursor-not-allowed border-neutral-200 bg-neutral-50 opacity-60 dark:border-neutral-800 dark:bg-neutral-900'
                    : 'border-neutral-200 bg-white hover:border-emerald-400 hover:bg-emerald-50 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-emerald-600 dark:hover:bg-emerald-950/40'"
                  :disabled="esgotado(p)"
                  @click="addToCart(p)"
                >
                  <!-- faixa lateral com a cor do grupo -->
                  <span class="absolute inset-y-0 left-0 w-1" :style="{ backgroundColor: g.cor }" />
                  <!-- selo da tecla de atalho (1-9) -->
                  <span
                    v-if="p.atalho > 0"
                    class="absolute right-2 top-2 flex size-6 items-center justify-center rounded-md bg-neutral-900 text-xs font-bold text-white shadow-sm dark:bg-neutral-200 dark:text-neutral-900"
                    :title="`Tecla ${p.atalho}: tecla adiciona 1 · Shift+${p.atalho} remove 1`"
                  >{{ p.atalho }}</span>
                  <span :class="['line-clamp-2 text-base font-semibold leading-snug text-neutral-900 dark:text-neutral-100', p.atalho > 0 ? 'pr-6' : '']">{{ p.nome }}</span>
                  <!-- mt-auto ancora preço + linha de estoque na base, mantendo o preço sempre alinhado -->
                  <span class="mt-auto text-xl font-semibold text-emerald-700 dark:text-emerald-400">{{ money(p.preco) }}</span>
                  <!-- linha reservada com altura fixa: preço não "desce" quando não há estoque p/ exibir -->
                  <span class="mt-1 block h-4 text-xs leading-none" :class="esgotado(p) ? 'text-red-500' : 'text-neutral-400'">
                    {{ p.estoqueTipo === 'limitado' ? (esgotado(p) ? 'Esgotado' : `${disponivel(p)} restantes`) : '' }}
                  </span>
                </button>
              </div>
            </section>
          </div>
        </div>
      </div>

      <!-- Pedido (direita) -->
      <div class="flex min-h-0 flex-col overflow-hidden rounded-xl border border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900">
        <div class="flex items-center justify-between border-b border-neutral-100 px-3 py-2 dark:border-neutral-800">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-shopping-cart" class="size-4 text-neutral-400" />
            <span class="text-sm font-medium">Pedido</span>
          </div>
          <UButton v-if="carrinho.length" color="neutral" variant="ghost" icon="i-lucide-trash-2" size="xs"
            label="Limpar" @click="cancelarDaTela" />
        </div>

        <div class="min-h-0 flex-1 overflow-auto p-3">
          <div v-if="carrinho.length === 0" class="flex h-full flex-col items-center justify-center gap-2 text-neutral-400">
            <UIcon name="i-lucide-shopping-bag" class="size-10" />
            <p class="text-sm">Toque nos produtos para montar o pedido</p>
          </div>
          <ul v-else class="space-y-2">
            <li
              v-for="l in carrinho"
              :key="l.key"
              class="flex items-center gap-2 rounded-lg border border-neutral-200 p-2 dark:border-neutral-800"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm font-medium text-neutral-900 dark:text-neutral-100">{{ l.nome }}</div>
                <div class="text-xs text-neutral-400">{{ money(l.preco) }} × {{ l.qtd }}</div>
              </div>
              <div class="flex items-center gap-1">
                <UButton color="neutral" variant="soft" icon="i-lucide-minus" size="xs" :aria-label="`Remover ${l.nome}`" @click="decQty(l)" />
                <span class="w-6 text-center text-sm font-semibold tabular-nums">{{ l.qtd }}</span>
                <UButton color="neutral" variant="soft" icon="i-lucide-plus" size="xs" :aria-label="`Adicionar ${l.nome}`" @click="incQty(l)" />
              </div>
              <div class="w-20 text-right text-sm font-semibold tabular-nums text-neutral-900 dark:text-neutral-100">
                {{ money(l.qtd * l.preco) }}
              </div>
              <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs"
                :aria-label="`Remover ${l.nome} do pedido`" title="Remover item" @click="removeLine(l)" />
            </li>
          </ul>
        </div>

        <div class="border-t border-neutral-100 p-3 dark:border-neutral-800">
          <div class="mb-2 flex items-center justify-between text-lg font-semibold">
            <span>Total</span>
            <span class="tabular-nums text-emerald-700 dark:text-emerald-400">{{ money(totalCarrinho) }}</span>
          </div>
          <UButton
            color="primary"
            size="lg"
            class="w-full"
            icon="i-lucide-check-check"
            :loading="!!buscando"
            :disabled="carrinho.length === 0 || !!buscando"
            @click="abrirPagamento"
          >
            {{ buscando || 'Fechar pedido' }}
          </UButton>
        </div>
      </div>
    </div>

    <!-- Recibo / simulação de impressão -->
    <UModal v-model:open="showReceipt" title="Pedido fechado" :ui="{ content: 'max-w-md' }">
      <template #body>
        <div v-if="pedidoFechado" class="rounded border border-dashed border-neutral-300 bg-neutral-50 p-4 font-mono text-[13px] leading-relaxed whitespace-pre dark:border-neutral-700 dark:bg-neutral-950">
          {{ receiptText() }}
        </div>

        <!-- Cartelas vendidas no pedido: layout ASCII a imprimir (1 cópia por unidade) -->
        <div v-if="cartelasImprimir.length" class="mt-4 space-y-4">
          <div>
            <h4 class="mb-1 text-sm font-semibold text-neutral-900 dark:text-neutral-100">Cartelas para imprimir</h4>
            <p class="mb-3 text-xs text-neutral-500">
              O conteúdo abaixo é o que vai à impressão térmica (MTP II) — ainda exibido aqui em modo debug. Uma cópia por cartela vendida.
            </p>
          </div>
          <div v-for="cp in cartelasImprimir" :key="cp.reais">
            <p class="mb-1 text-xs font-medium text-neutral-500 dark:text-neutral-400">
              {{ cp.nome }} · {{ cp.qtd }} {{ cp.qtd === 1 ? 'cópia' : 'cópias' }}
            </p>
            <pre class="overflow-x-auto rounded border border-dashed border-violet-300 bg-neutral-50 p-3 font-mono text-[12px] leading-tight whitespace-pre text-neutral-800 dark:border-violet-800 dark:bg-neutral-950 dark:text-neutral-200">{{ cp.conteudo }}</pre>
          </div>
        </div>

        <p v-if="!cartelasImprimir.length" class="mt-3 text-sm text-neutral-500">
          Em modo debug, o recibo é exibido aqui. No futuro, será enviado à impressora térmica (MTP II).
        </p>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="primary" variant="soft" @click="showReceipt = false">Novo pedido</UButton>
        </div>
      </template>
    </UModal>

    <!-- Forma de pagamento -->
    <UModal v-model:open="showPagamento" title="Forma de pagamento" :ui="{ content: 'max-w-md' }" :dismissible="!buscando">
      <template #body>
        <div class="space-y-4">
          <div class="grid grid-cols-3 gap-2">
            <button
              type="button"
              @click="selecionarForma('dinheiro')"
              class="flex flex-col items-center gap-2 rounded-xl border p-3 text-sm font-medium transition
                focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              :class="forma === 'dinheiro'
                ? 'border-emerald-500 bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400'
                : 'border-neutral-200 text-neutral-600 hover:border-neutral-300 dark:border-neutral-700 dark:text-neutral-300'"
            >
              <UIcon name="i-lucide-banknote" class="size-6" />
              Dinheiro
            </button>
            <button
              type="button"
              @click="selecionarForma('cartao')"
              class="flex flex-col items-center gap-2 rounded-xl border p-3 text-sm font-medium transition
                focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              :class="forma === 'cartao'
                ? 'border-blue-500 bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-400'
                : 'border-neutral-200 text-neutral-600 hover:border-neutral-300 dark:border-neutral-700 dark:text-neutral-300'"
            >
              <UIcon name="i-lucide-credit-card" class="size-6" />
              Cartão
            </button>
            <button
              type="button"
              @click="selecionarForma('anotaai')"
              class="flex flex-col items-center gap-2 rounded-xl border p-3 text-sm font-medium transition
                focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500"
              :class="forma === 'anotaai'
                ? 'border-amber-500 bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-400'
                : 'border-neutral-200 text-neutral-600 hover:border-neutral-300 dark:border-neutral-700 dark:text-neutral-300'"
            >
              <UIcon name="i-lucide-notebook-pen" class="size-6" />
              Anota aí
            </button>
          </div>
          <p class="text-center text-xs text-neutral-400">Tecla <kbd class="rounded border border-neutral-300 px-1 font-mono dark:border-neutral-600">1</kbd> Dinheiro · <kbd class="rounded border border-neutral-300 px-1 font-mono dark:border-neutral-600">2</kbd> Cartão · <kbd class="rounded border border-neutral-300 px-1 font-mono dark:border-neutral-600">3</kbd> Anota aí</p>

          <div class="flex items-center justify-between rounded-lg bg-neutral-50 px-3 py-2 text-sm dark:bg-neutral-900">
            <span class="text-neutral-500">Total</span>
            <span class="text-lg font-semibold tabular-nums text-neutral-900 dark:text-neutral-100">{{ money(totalCarrinho) }}</span>
          </div>

          <!-- Dinheiro: valor recebido + troco -->
          <div v-if="forma === 'dinheiro'" class="space-y-3">
            <UFormField label="Valor recebido (R$)">
              <div class="flex items-center gap-2">
                <UInput ref="recebidoInput" v-model="recebidoText" inputmode="decimal" placeholder="0,00" size="lg" class="flex-1" />
                <UButton color="neutral" variant="soft" label="Exato" size="lg"
                  @click="recebidoText = centsToInput(totalCarrinho)" />
              </div>
            </UFormField>
            <div v-if="recebidoCents > 0" class="flex items-center justify-between px-1 text-sm">
              <span class="text-neutral-500">Troco</span>
              <span class="text-base font-semibold tabular-nums" :class="trocoSuficiente ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500'">
                {{ trocoSuficiente ? money(troco) : 'Valor insuficiente' }}
              </span>
            </div>
          </div>

          <!-- Cartão: aguardando pagamento -->
          <div v-else-if="forma === 'cartao'"
            class="flex items-center gap-3 rounded-lg bg-blue-50 px-3 py-3 text-sm text-blue-700 dark:bg-blue-950/40 dark:text-blue-300">
            <UIcon name="i-lucide-loader-circle" class="size-5 shrink-0 animate-spin" />
            Aguardando pagamento na maquininha. Assim que aprovar, toque em “Finalizar compra”.
          </div>

          <!-- Anota aí (fiado): escolhe o dono da conta via CommandPalette.
               Contas existentes do evento são buscáveis; se o texto digitado não
               casa com nenhuma, "Criar conta …" cadastra o novo dono. A venda é
               lançada na conta escolhida (vira pendência dela). -->
          <div v-else-if="forma === 'anotaai'" class="space-y-3">
            <p class="text-xs text-neutral-500 dark:text-neutral-400">
              Quem vai pagar depois? Escolha a conta da pessoa (se já existe neste evento) ou crie uma nova digitando o nome.
            </p>

            <UCommandPalette
              v-if="!contaEscolhida"
              v-model:search-term="contaBusca"
              :groups="contaGrupos"
              :fuse="{ resultLimit: 4 }"
              placeholder="Busque ou digite o nome…"
              :ui="{
                root: 'w-full',
                content: 'max-h-64 overflow-y-auto',
                input: 'px-3 py-2',
                empty: 'px-3 py-4 text-center text-sm text-neutral-500 dark:text-neutral-400'
              }"
            >
              <template #conta="{ item }">
                <span class="flex w-full items-center justify-between gap-3">
                  <span class="flex min-w-0 items-center gap-2">
                    <UIcon name="i-lucide-user" class="size-4 shrink-0 text-neutral-400" />
                    <span class="truncate">{{ item.label }}</span>
                  </span>
                  <span class="shrink-0 text-xs font-medium tabular-nums text-neutral-500 dark:text-neutral-400">
                    {{ money(item.valor) }}
                  </span>
                </span>
              </template>

              <template #empty="{ searchTerm }">
                <span v-if="contas.length === 0">Ainda não há contas neste evento — digite um nome para criar a primeira.</span>
                <span v-else>Digite o nome para <strong>“Criar conta {{ searchTerm }}”</strong> aparecer na lista.</span>
              </template>
            </UCommandPalette>

            <!-- conta escolhida: mostra quem e se é nova ou existente -->
            <div v-else
              class="flex items-center gap-3 rounded-lg border px-3 py-3 text-sm"
              :class="contaEscolhida.nova
                ? 'border-amber-300 bg-amber-50 text-amber-800 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-300'
                : 'border-emerald-300 bg-emerald-50 text-emerald-800 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-300'"
            >
              <UIcon :name="contaEscolhida.nova ? 'i-lucide-user-plus' : 'i-lucide-user-check'" class="size-5 shrink-0" />
              <div class="min-w-0 flex-1">
                <p class="truncate font-semibold">{{ contaEscolhida.nome }}</p>
                <p class="text-xs opacity-80">
                  {{ contaEscolhida.nova ? 'Nova conta — será criada e a venda lançada nela.' : 'Conta existente — a venda será lançada nela.' }}
                </p>
              </div>
              <UButton color="neutral" variant="soft" size="xs" icon="i-lucide-rotate-ccw" label="Trocar" @click="limparConta" />
            </div>
          </div>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-between gap-2">
          <UButton color="error" variant="ghost" icon="i-lucide-x" label="Cancelar pedido" @click="cancelarDoPagamento" />
          <UButton color="primary" icon="i-lucide-check" label="Finalizar compra"
            :loading="!!buscando" :disabled="!podeFinalizar()" @click="finalizarCompra" />
        </div>
      </template>
    </UModal>

    <!-- Confirmação de cancelamento do pedido -->
    <UModal v-model:open="showConfirmCancelar" title="Cancelar pedido?" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <p class="text-sm text-neutral-500">
          O pedido em montagem será descartado e o estoque devolvido. Tem certeza que deseja cancelar?
        </p>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="soft" label="Voltar" @click="voltarSemCancelar" />
          <UButton color="error" label="Sim, cancelar" icon="i-lucide-trash-2" @click="confirmarCancelar" />
        </div>
      </template>
    </UModal>
  </div>
</template>
