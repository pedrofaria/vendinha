<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { dtBR, money } from '../lib/format'
import type { ListaPedidos, Pedido, PedidoResumo } from '../lib/types'

const route = useRoute()
const eventoId = computed(() => Number(route.params.eventoId))

const POR_PAGINA = 20

const lista = ref<ListaPedidos | null>(null)
const loading = ref(false)
const error = ref('')

// Paginação e busca (nº exato do pedido).
const pagina = ref(1)
const numeroBusca = ref('')

const pedidos = computed(() => lista.value?.pedidos ?? [])
const totalPaginas = computed(() => lista.value?.totalPaginas ?? 1)
const buscaNum = computed(() => {
  const n = parseInt(numeroBusca.value.replace(/\D/g, ''), 10)
  return Number.isFinite(n) && n > 0 ? n : 0
})

// Token evita que respostas fora de ordem (busca rápida) sobrescrevam a atual.
let reqToken = 0
async function load() {
  const token = ++reqToken
  loading.value = true
  error.value = ''
  try {
    const res = (await api().ListPedidos(eventoId.value, pagina.value, POR_PAGINA, buscaNum.value)) ?? null
    if (token !== reqToken) return
    lista.value = res
    if (res) pagina.value = res.pagina // servidor ajusta p/ página válida
  } catch (e) {
    if (token !== reqToken) return
    error.value = errMsg(e)
  } finally {
    if (token === reqToken) loading.value = false
  }
}

function irPara(p: number) {
  if (p < 1 || p > totalPaginas.value || p === pagina.value) return
  pagina.value = p
  load()
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(numeroBusca, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    pagina.value = 1
    load()
  }, 350)
})

watch(eventoId, () => {
  numeroBusca.value = ''
  pagina.value = 1
  load()
})

onMounted(load)

// ---- rótulos / badges ----
interface FormaInfo {
  label: string
  icon: string
  cls: string
}
function formaInfo(f: string): FormaInfo {
  switch (f) {
    case 'cartao':
      return { label: 'Cartão', icon: 'i-lucide-credit-card', cls: 'bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300' }
    case 'anotaai':
      return { label: 'Anota aí', icon: 'i-lucide-notebook-pen', cls: 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' }
    default:
      return { label: 'Dinheiro', icon: 'i-lucide-banknote', cls: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300' }
  }
}

function statusBadge(p: PedidoResumo): { label: string; cls: string } | null {
  if (p.canceladoEm) {
    return { label: 'Cancelado', cls: 'bg-red-100 text-red-700 dark:bg-red-950 dark:text-red-300' }
  }
  if (p.forma === 'anotaai') {
    return p.quitadoEm
      ? { label: `Pago em ${dtBR(p.quitadoEm)}`, cls: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300' }
      : { label: 'Em aberto', cls: 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' }
  }
  return null
}

// ---- modal: detalhes do pedido (passo 'detalhe') → cancelamento (passo 'confirmar') ----
const modalOpen = ref(false)
const passo = ref<'detalhe' | 'confirmar'>('detalhe')
const modalPedido = ref<PedidoResumo | null>(null)
const detalhe = ref<Pedido | null>(null)
const detalheLoading = ref(false)
const detalheError = ref('')

const cancelBusy = ref(false)
const cancelError = ref('')

const modalTitulo = computed(() => {
  if (!modalPedido.value) return ''
  return passo.value === 'confirmar'
    ? `Cancelar pedido #${modalPedido.value.numero}?`
    : `Pedido #${modalPedido.value.numero}`
})

const temProduto = computed(() => !!detalhe.value?.itens.some((i) => i.produtoId > 0))
const soCartelas = computed(() => !!detalhe.value?.itens.some((i) => i.cartelaReais > 0))

// Nota do efeito do cancelamento (venda anotada em aberto → remove o débito).
function notaCancelamento(p: PedidoResumo | null): string | null {
  if (p?.forma === 'anotaai' && !p.quitadoEm) {
    const dono = p.contaNome ? ` de "${p.contaNome}"` : ''
    return `Esta venda está anotada em aberto${dono}. Ao cancelar, o débito de ${money(p.total)} é removido da conta.`
  }
  return null
}

async function abrirDetalhe(p: PedidoResumo) {
  modalPedido.value = p
  passo.value = 'detalhe'
  detalhe.value = null
  detalheError.value = ''
  cancelError.value = ''
  modalOpen.value = true
  detalheLoading.value = true
  try {
    detalhe.value = await api().GetPedido(p.id)
  } catch (e) {
    detalheError.value = errMsg(e)
  } finally {
    detalheLoading.value = false
  }
}

function proporCancelar() {
  cancelError.value = ''
  passo.value = 'confirmar'
}

function voltarAoDetalhe() {
  cancelError.value = ''
  passo.value = 'detalhe'
}

function fecharModal() {
  modalOpen.value = false
  modalPedido.value = null
  detalhe.value = null
}

async function confirmarCancelar() {
  if (!modalPedido.value) return
  cancelBusy.value = true
  cancelError.value = ''
  try {
    await api().CancelarPedido(modalPedido.value.id)
    fecharModal()
    await load() // recarrega a página atual (o cancelado segue listado com selo)
  } catch (e) {
    cancelError.value = errMsg(e)
  } finally {
    cancelBusy.value = false
  }
}

// Janela de páginas exibidas na paginação.
const paginasVisiveis = computed(() => {
  const total = totalPaginas.value
  const atual = pagina.value
  const inicio = Math.max(1, Math.min(atual - 2, total - 4))
  const fim = Math.min(total, inicio + 4)
  const out: number[] = []
  for (let i = inicio; i <= fim; i++) out.push(i)
  return out
})
</script>

<template>
  <div class="mx-auto max-w-5xl p-6">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold">Pedidos</h2>
        <p class="text-sm text-neutral-500">Vendas fechadas deste evento, da mais recente para a mais antiga.</p>
      </div>
      <UButton color="neutral" variant="soft" icon="i-lucide-refresh-cw" title="Atualizar" @click="load" />
    </div>

    <!-- Busca por nº do pedido -->
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <UInput
        v-model="numeroBusca"
        icon="i-lucide-search"
        placeholder="Buscar pelo nº do pedido"
        class="w-64"
        clearable
        :aria-label="'Buscar pelo número do pedido'"
      />
      <span v-if="lista" class="text-sm text-neutral-500">
        {{ lista.total }} {{ lista.total === 1 ? 'pedido' : 'pedidos' }}
        <template v-if="buscaNum"> para o nº {{ buscaNum }}</template>
      </span>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <div
      v-else-if="pedidos.length === 0"
      class="rounded-xl border border-dashed border-neutral-300 p-10 text-center text-neutral-500 dark:border-neutral-700"
    >
      <UIcon name="i-lucide-clipboard-list" class="mx-auto size-8 text-neutral-300 dark:text-neutral-600" />
      <p class="mt-2 text-lg font-medium">
        {{ buscaNum ? 'Nenhum pedido com esse número.' : 'Nenhum pedido fechado ainda.' }}
      </p>
      <p v-if="!buscaNum" class="mt-1 text-sm">As vendas fechadas no PDV aparecem aqui, com os pedidos mais recentes primeiro.</p>
    </div>

    <template v-else>
      <ul class="space-y-2">
        <li
          v-for="p in pedidos"
          :key="p.id"
          role="button"
          title="Ver detalhes do pedido"
          class="flex cursor-pointer items-center gap-4 rounded-xl border border-neutral-200 bg-white p-3 transition hover:border-neutral-300 hover:bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-neutral-700 dark:hover:bg-neutral-800/40"
          :class="p.canceladoEm ? 'opacity-75' : ''"
          @click="abrirDetalhe(p)"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="font-semibold text-neutral-900 dark:text-neutral-100">#{{ p.numero }}</span>
              <span class="text-xs text-neutral-400">{{ dtBR(p.criadoEm) }}</span>
              <span
                v-if="statusBadge(p)"
                class="rounded-full px-2 py-0.5 text-[11px] font-medium"
                :class="statusBadge(p)!.cls"
              >
                {{ statusBadge(p)!.label }}
              </span>
            </div>
            <div class="mt-1 flex flex-wrap items-center gap-1.5">
              <span
                class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium"
                :class="formaInfo(p.forma).cls"
              >
                <UIcon :name="formaInfo(p.forma).icon" class="size-3" />
                {{ formaInfo(p.forma).label }}
                <template v-if="p.forma === 'anotaai' && p.contaNome"> · {{ p.contaNome }}</template>
              </span>
            </div>
          </div>

          <div
            class="shrink-0 text-right text-lg font-semibold text-neutral-900 dark:text-neutral-100"
            :class="p.canceladoEm ? 'text-neutral-400 line-through dark:text-neutral-500' : ''"
          >
            {{ money(p.total) }}
          </div>

          <UIcon name="i-lucide-chevron-right" class="shrink-0 size-4 text-neutral-300 dark:text-neutral-600" />
        </li>
      </ul>

      <!-- Paginação -->
      <div v-if="totalPaginas > 1" class="mt-5 flex items-center justify-center gap-1">
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-lucide-chevron-left"
          size="sm"
          :disabled="pagina <= 1"
          aria-label="Página anterior"
          @click="irPara(pagina - 1)"
        />
        <button
          v-for="pg in paginasVisiveis"
          :key="pg"
          type="button"
          class="h-8 min-w-8 rounded-md px-2 text-sm font-medium transition"
          :class="pg === pagina ? 'bg-emerald-600 text-white' : 'text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800'"
          @click="irPara(pg)"
        >
          {{ pg }}
        </button>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-lucide-chevron-right"
          size="sm"
          :disabled="pagina >= totalPaginas"
          aria-label="Próxima página"
          @click="irPara(pagina + 1)"
        />
      </div>
    </template>

    <!-- Modal de detalhes do pedido (passo 'confirmar' = confirmação de cancelamento) -->
    <UModal v-model:open="modalOpen" :title="modalTitulo" :ui="{ content: 'max-w-lg' }">
      <template #body>
        <!-- Cabeçalho: status + forma + conta + data -->
        <div v-if="modalPedido" class="flex flex-wrap items-center gap-2">
          <span
            v-if="statusBadge(modalPedido)"
            class="rounded-full px-2 py-0.5 text-[11px] font-medium"
            :class="statusBadge(modalPedido)!.cls"
          >
            {{ statusBadge(modalPedido)!.label }}
          </span>
          <span
            class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium"
            :class="formaInfo(modalPedido.forma).cls"
          >
            <UIcon :name="formaInfo(modalPedido.forma).icon" class="size-3" />
            {{ formaInfo(modalPedido.forma).label }}
          </span>
          <span v-if="modalPedido.forma === 'anotaai' && modalPedido.contaNome" class="text-sm text-neutral-500">
            Conta de {{ modalPedido.contaNome }}
          </span>
          <span class="text-xs text-neutral-400">{{ dtBR(modalPedido.criadoEm) }}</span>
        </div>

        <div
          v-if="modalPedido?.canceladoEm"
          class="mt-3 flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/60 dark:bg-red-950/40 dark:text-red-300"
        >
          <UIcon name="i-lucide-ban" class="mt-0.5 size-4 shrink-0 text-red-500" />
          <span>Pedido cancelado em {{ dtBR(modalPedido.canceladoEm) }}. Não é possível reativar.</span>
        </div>

        <!-- Passo confirmar: avisos antes de cancelar -->
        <template v-if="passo === 'confirmar'">
          <p class="mt-3 text-sm text-neutral-500">
            Esta ação não pode ser desfeita. Pedidos não podem ser editados — para corrigir, cancele e refaça a venda.
          </p>
          <div
            v-if="notaCancelamento(modalPedido)"
            class="mt-3 flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-900/60 dark:bg-amber-950/40 dark:text-amber-200"
          >
            <UIcon name="i-lucide-hand-coins" class="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400" />
            <span>{{ notaCancelamento(modalPedido) }}</span>
          </div>
          <p v-if="cancelError" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ cancelError }}</p>
        </template>

        <div v-if="detalheLoading" class="flex justify-center py-8">
          <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
        </div>

        <template v-else-if="detalhe">
          <ul class="mt-3 max-h-[45vh] space-y-1 overflow-auto rounded-lg border border-neutral-200 p-3 dark:border-neutral-800">
            <li v-for="it in detalhe.itens" :key="it.id" class="flex items-center justify-between gap-2 text-sm">
              <span class="min-w-0 truncate text-neutral-700 dark:text-neutral-200">
                {{ it.nome }} <span class="text-neutral-400">× {{ it.qtd }}</span>
              </span>
              <span class="shrink-0 text-neutral-500 dark:text-neutral-400">{{ money(it.subtotal) }}</span>
            </li>
          </ul>

          <div class="mt-3 flex items-center justify-between gap-2 text-sm">
            <span v-if="passo === 'confirmar'" class="text-neutral-500">
              <template v-if="temProduto">Se houver, o estoque dos produtos limitados deste pedido é devolvido.</template>
              <template v-else-if="soCartelas">Cartelas são ilimitadas (não há estoque a devolver).</template>
              <template v-else>Nenhum item neste pedido.</template>
            </span>
            <span class="text-lg font-semibold text-neutral-900 dark:text-neutral-100">{{ money(detalhe.total) }}</span>
          </div>
        </template>

        <p v-else-if="detalheError" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ detalheError }}</p>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2">
          <!-- passo detalhe: só visualização; cancelar vai para a confirmação -->
          <template v-if="passo === 'detalhe'">
            <UButton color="neutral" variant="ghost" label="Fechar" @click="fecharModal" />
            <UButton
              v-if="modalPedido && !modalPedido.canceladoEm"
              color="error"
              variant="soft"
              icon="i-lucide-ban"
              label="Cancelar pedido"
              @click="proporCancelar"
            />
          </template>
          <!-- passo confirmar -->
          <template v-else>
            <UButton color="neutral" variant="soft" label="Voltar" @click="voltarAoDetalhe" />
            <UButton
              color="error"
              icon="i-lucide-ban"
              :label="cancelBusy ? 'Cancelando…' : 'Sim, cancelar pedido'"
              :loading="cancelBusy"
              @click="confirmarCancelar"
            />
          </template>
        </div>
      </template>
    </UModal>
  </div>
</template>
