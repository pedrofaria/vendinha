<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { money } from '../lib/format'
import type { Evento, Pedido, PedidoItem, ProdutoVenda } from '../lib/types'

const route = useRoute()
const router = useRouter()
const eventoId = computed(() => Number(route.params.eventoId))

const evento = ref<Evento | null>(null)
const produtos = ref<ProdutoVenda[]>([])
const loading = ref(false)
const error = ref('')

// Pedido em montagem (cliente local): produtoId -> linha
interface LinhaCarrinho {
  produtoId: number
  nome: string
  preco: number
  qtd: number
  esgotavel: boolean
  estoqueRestante: number
}
const carrinho = ref<LinhaCarrinho[]>([])
const buscando = ref<string>('')

// Resultado do pedido fechado + modal de recibo/impressão
const showReceipt = ref(false)
const pedidoFechado = ref<Pedido | null>(null)

const totalCarrinho = computed(() =>
  carrinho.value.reduce((s, l) => s + l.qtd * l.preco, 0)
)
const itensCount = computed(() => carrinho.value.reduce((s, l) => s + l.qtd, 0))

function addToCart(p: ProdutoVenda) {
  if (p.esgotado) return
  let linha = carrinho.value.find((l) => l.produtoId === p.id)
  if (!linha) {
    linha = {
      produtoId: p.id,
      nome: p.nome,
      preco: p.preco,
      qtd: 0,
      esgotavel: p.estoqueTipo === 'limitado',
      estoqueRestante: p.quantidade
    }
    carrinho.value.push(linha)
  }
  const max = linha.esgotavel ? linha.estoqueRestante : Infinity
  if (linha.qtd < max) linha.qtd++
  updateStockAfterEdit(linha, max)
}

function incQty(linha: LinhaCarrinho) {
  const max = linha.esgotavel ? linha.estoqueRestante : Infinity
  if (linha.qtd < max) {
    linha.qtd++
    updateStockAfterEdit(linha, max)
  }
}

function decQty(linha: LinhaCarrinho) {
  if (linha.qtd <= 1) {
    removeLine(linha)
    return
  }
  linha.qtd--
  updateStockAfterEdit(linha, linha.esgotavel ? linha.estoqueRestante : Infinity)
}

function updateStockAfterEdit(linha: LinhaCarrinho, max: number) {
  // marca produtos esgotados pela qtd já no carrinho
  if (linha.esgotavel && linha.qtd >= max) {
    // produto chega ao limite: atualiza esgotado p/ bloquear + no grid
    const p = produtos.value.find((x) => x.id === linha.produtoId)
    if (p) p.esgotado = true
  }
}

function removeLine(linha: LinhaCarrinho) {
  carrinho.value = carrinho.value.filter((l) => l.produtoId !== linha.produtoId)
  // reabilita no grid
  const p = produtos.value.find((x) => x.id === linha.produtoId)
  if (p && !(p.esgotado && linha.esgotavel)) {
    // se havia sobrado estoque, deixa clicável de novo conforme disponível
    p.esgotado = p.estoqueTipo === 'limitado' && p.quantidade <= 0
  }
}

function clearCart() {
  carrinho.value = []
  produtos.value.forEach((p) => {
    if (p.estoqueTipo === 'limitado') p.esgotado = p.quantidade <= 0
  })
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
    produtos.value = await api().ListProdutosVenda(eventoId.value)
    // aplica reserva já no carrinho, se houver (ex.: reload)
    carrinho.value = []
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

async function fecharPedido() {
  if (carrinho.value.length === 0) return
  error.value = ''
  buscando.value = 'Fechando pedido…'
  try {
    const pedido = await api().CriarPedido(eventoId.value)
    const itens: PedidoItem[] = carrinho.value.map((l) => ({
      id: 0,
      produtoId: l.produtoId,
      nome: l.nome,
      precoUnit: l.preco,
      qtd: l.qtd,
      subtotal: l.qtd * l.preco
    }))
    pedidoFechado.value = await api().FecharPedido(pedido.id, itens)
    showReceipt.value = true
    clearCart()
    await loadProdutos() // reflete estoque novo
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
  lines.push('')
  lines.push('Obrigado!')
  return lines.join('\n')
}

onMounted(() => {
  loadEvento()
  loadProdutos()
})
</script>

<template>
  <div class="flex h-full flex-col">
    <!-- barra do PDV -->
    <div class="flex items-center justify-between border-b border-neutral-200 bg-white px-4 py-2 dark:border-neutral-800 dark:bg-neutral-900">
      <div class="flex items-center gap-3">
        <UButton color="neutral" variant="ghost" icon="i-lucide-arrow-left" size="sm" @click="router.push('/eventos')">
          Eventos
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
        <div class="flex items-center gap-2 border-b border-neutral-100 px-3 py-2 dark:border-neutral-800">
          <UIcon name="i-lucide-package" class="size-4 text-neutral-400" />
          <span class="text-sm font-medium">Produtos</span>
        </div>
        <div class="min-h-0 flex-1 overflow-auto p-3">
          <div v-if="loading" class="flex justify-center py-10">
            <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
          </div>
          <div v-else-if="produtos.length === 0" class="py-10 text-center text-neutral-400">
            Nenhum produto disponível. Cadastre antes de vender.
          </div>
          <div v-else class="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            <button
              v-for="p in produtos"
              :key="p.id"
              type="button"
              class="flex min-h-[92px] flex-col justify-between rounded-xl border p-3 text-left transition
                focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
              :class="p.esgotado
                ? 'cursor-not-allowed border-neutral-200 bg-neutral-50 opacity-60 dark:border-neutral-800 dark:bg-neutral-900'
                : 'border-neutral-200 bg-white hover:border-emerald-400 hover:bg-emerald-50 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-emerald-600 dark:hover:bg-emerald-950/40'"
              :disabled="p.esgotado"
              @click="addToCart(p)"
            >
              <span class="text-sm font-medium leading-tight text-neutral-900 dark:text-neutral-100">{{ p.nome }}</span>
              <span class="mt-1 text-lg font-semibold text-emerald-700 dark:text-emerald-400">{{ money(p.preco) }}</span>
              <span v-if="p.estoqueTipo === 'limitado'" class="text-xs" :class="p.esgotado ? 'text-red-500' : 'text-neutral-400'">
                {{ p.esgotado ? 'Esgotado' : `${p.quantidade} restantes` }}
              </span>
            </button>
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
            label="Limpar" @click="clearCart" />
        </div>

        <div class="min-h-0 flex-1 overflow-auto p-3">
          <div v-if="carrinho.length === 0" class="flex h-full flex-col items-center justify-center gap-2 text-neutral-400">
            <UIcon name="i-lucide-shopping-bag" class="size-10" />
            <p class="text-sm">Toque nos produtos para montar o pedido</p>
          </div>
          <ul v-else class="space-y-2">
            <li
              v-for="l in carrinho"
              :key="l.produtoId"
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
            @click="fecharPedido"
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
        <p class="mt-3 text-sm text-neutral-500">
          Em modo debug, o recibo é exibido aqui. No futuro, será enviado à impressora térmica (MTP II).
        </p>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="primary" variant="soft" @click="showReceipt = false">Novo pedido</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
