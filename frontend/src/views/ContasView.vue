<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { dtBR, money } from '../lib/format'
import type { ContaSaldo, Pedido } from '../lib/types'

const route = useRoute()
const eventoId = computed(() => Number(route.params.eventoId))

const contas = ref<ContaSaldo[]>([])
const loading = ref(false)
const error = ref('')

// Filtro da lista: 'pendentes' (devendo) ou 'quitadas' (em dia).
const aba = ref<'pendentes' | 'quitadas'>('pendentes')

const pendentes = computed(() => contas.value.filter((c) => c.totalPendente > 0))
const quitadas = computed(() => contas.value.filter((c) => c.totalPendente === 0))
const aReceber = computed(() => pendentes.value.reduce((s, c) => s + c.totalPendente, 0))

async function load() {
  loading.value = true
  error.value = ''
  try {
    contas.value = (await api().ListContasSaldo(eventoId.value)) ?? []
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

// ---- modal de detalhe da conta (lista de pedidos) ----
const showModal = ref(false)
const modalConta = ref<ContaSaldo | null>(null)
const pedidos = ref<Pedido[]>([])
const pedLoading = ref(false)
const pedError = ref('')

const numAbertos = computed(() => pedidos.value.filter((p) => !p.quitadoEm).length)
const numPagos = computed(() => pedidos.value.filter((p) => p.quitadoEm).length)
const totalEmAberto = computed(() =>
  pedidos.value.filter((p) => !p.quitadoEm).reduce((s, p) => s + p.total, 0)
)

async function abrirDetalhe(c: ContaSaldo) {
  modalConta.value = c
  pedidos.value = []
  pedError.value = ''
  showModal.value = true
  pedLoading.value = true
  try {
    pedidos.value = (await api().ListPedidosConta(c.id)) ?? []
  } catch (e) {
    pedError.value = errMsg(e)
  } finally {
    pedLoading.value = false
  }
}

function fechar() {
  showModal.value = false
  modalConta.value = null
  pedidos.value = []
}

async function receber(c: ContaSaldo) {
  if (
    !confirm(
      `Receber ${money(c.totalPendente)} de "${c.nome}"?\n\n` +
        `Marca como pagas as ${c.numAberto} ${c.numAberto === 1 ? 'venda anotada em aberto' : 'vendas anotadas em aberto'} da conta.`
    )
  ) {
    return
  }
  error.value = ''
  try {
    await api().QuitarConta(c.id)
    await load()
    if (modalConta.value?.id === c.id) fechar()
  } catch (e) {
    error.value = errMsg(e)
  }
}

watch(eventoId, load)
onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-5xl p-6">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold">Anota aí</h2>
        <p class="text-sm text-neutral-500">Vendas anotadas (fiado) deste evento, do maior devedor para o menor.</p>
      </div>
      <UButton color="neutral" variant="soft" icon="i-lucide-refresh-cw" title="Atualizar" @click="load" />
    </div>

    <!-- Resumo + filtro -->
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 dark:border-amber-900/60 dark:bg-amber-950/30">
        <UIcon name="i-lucide-hand-coins" class="size-4 text-amber-600 dark:text-amber-400" />
        <span class="text-sm text-amber-800 dark:text-amber-200">
          A receber: <strong class="font-semibold">{{ money(aReceber) }}</strong>
          <span v-if="pendentes.length" class="text-amber-600/80 dark:text-amber-300/70"> · {{ pendentes.length }} {{ pendentes.length === 1 ? 'conta' : 'contas' }}</span>
        </span>
      </div>
      <div class="flex rounded-lg border border-neutral-200 p-0.5 dark:border-neutral-800">
        <button
          type="button"
          class="rounded-md px-3 py-1.5 text-sm font-medium transition"
          :class="aba === 'pendentes' ? 'bg-emerald-600 text-white' : 'text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800'"
          @click="aba = 'pendentes'"
        >
          Pendentes ({{ pendentes.length }})
        </button>
        <button
          type="button"
          class="rounded-md px-3 py-1.5 text-sm font-medium transition"
          :class="aba === 'quitadas' ? 'bg-emerald-600 text-white' : 'text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800'"
          @click="aba = 'quitadas'"
        >
          Em dia ({{ quitadas.length }})
        </button>
      </div>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <div v-else-if="contas.length === 0" class="rounded-xl border border-dashed border-neutral-300 p-10 text-center text-neutral-500 dark:border-neutral-700">
      <UIcon name="i-lucide-notebook-pen" class="mx-auto size-8 text-neutral-300 dark:text-neutral-600" />
      <p class="mt-2 text-lg font-medium">Nenhuma conta anotada ainda.</p>
      <p class="mt-1 text-sm">As vendas fechadas como <strong>Anota aí</strong> no PDV criam a conta aqui automaticamente.</p>
    </div>

    <!-- Lista de contas pendentes -->
    <ul v-else-if="aba === 'pendentes'" class="space-y-2">
      <li v-if="pendentes.length === 0" class="rounded-lg border border-dashed border-neutral-300 p-6 text-center text-sm text-neutral-400 dark:border-neutral-700">
        Tudo recebido! Nenhuma conta em aberto.
      </li>
      <li
        v-for="c in pendentes"
        :key="c.id"
        class="flex cursor-pointer items-center gap-3 rounded-xl border border-neutral-200 bg-white p-3 transition hover:border-emerald-300 hover:bg-emerald-50/40 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-emerald-800 dark:hover:bg-emerald-950/20"
        role="button"
        title="Ver pedidos desta conta"
        @click="abrirDetalhe(c)"
      >
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
          <UIcon name="i-lucide-user" class="size-5" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="truncate font-medium text-neutral-900 dark:text-neutral-100">{{ c.nome }}</div>
          <div class="text-xs text-neutral-500">
            {{ c.numAberto }} {{ c.numAberto === 1 ? 'venda em aberto' : 'vendas em aberto' }}
            <span v-if="c.totalQuitado > 0"> · já quitou {{ money(c.totalQuitado) }}</span>
          </div>
        </div>
        <div class="shrink-0 text-right">
          <div class="text-lg font-semibold text-amber-600 dark:text-amber-400">{{ money(c.totalPendente) }}</div>
        </div>
        <UButton
          color="primary"
          icon="i-lucide-check"
          :label="`Receber ${money(c.totalPendente)}`"
          size="sm"
          @click.stop="receber(c)"
        />
      </li>
    </ul>

    <!-- Lista de contas em dia (quitadas) -->
    <ul v-else class="space-y-2">
      <li v-if="quitadas.length === 0" class="rounded-lg border border-dashed border-neutral-300 p-6 text-center text-sm text-neutral-400 dark:border-neutral-700">
        Nenhuma conta quitada ainda.
      </li>
      <li
        v-for="c in quitadas"
        :key="c.id"
        class="flex cursor-pointer items-center gap-3 rounded-xl border border-neutral-200 bg-white p-3 opacity-80 transition hover:border-emerald-300 hover:bg-emerald-50/40 dark:border-neutral-800 dark:bg-neutral-900 dark:hover:border-emerald-800 dark:hover:bg-emerald-950/20"
        role="button"
        title="Ver pedidos desta conta"
        @click="abrirDetalhe(c)"
      >
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-300">
          <UIcon name="i-lucide-user" class="size-5" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="truncate font-medium text-neutral-900 dark:text-neutral-100">{{ c.nome }}</span>
            <span class="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
              Em dia
            </span>
          </div>
          <div v-if="c.totalQuitado > 0" class="text-xs text-neutral-500">Já quitou {{ money(c.totalQuitado) }}</div>
        </div>
        <div class="shrink-0 text-right text-lg font-semibold text-emerald-600 dark:text-emerald-400">R$ 0,00</div>
      </li>
    </ul>

    <!-- Modal: pedidos da conta -->
    <UModal v-model:open="showModal" :title="`Conta de ${modalConta?.nome ?? ''}`" :ui="{ content: 'max-w-2xl' }">
      <template #body>
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <span class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-800 dark:bg-amber-950 dark:text-amber-300">
            {{ numAbertos }} em aberto · {{ money(totalEmAberto) }}
          </span>
          <span class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300">
            {{ numPagos }} pagos
          </span>
        </div>

        <p v-if="pedError" class="mb-3 text-sm text-red-600 dark:text-red-400">{{ pedError }}</p>

        <div v-if="pedLoading" class="flex justify-center py-8">
          <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
        </div>

        <div v-else-if="pedidos.length === 0" class="rounded-lg border border-dashed border-neutral-300 p-6 text-center text-sm text-neutral-400 dark:border-neutral-700">
          Nenhum pedido nesta conta.
        </div>

        <ul v-else class="max-h-[50vh] space-y-2 overflow-auto pr-1">
          <li v-for="p in pedidos" :key="p.id" class="rounded-lg border border-neutral-200 p-3 dark:border-neutral-800">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div class="text-sm font-medium text-neutral-900 dark:text-neutral-100">
                Pedido #{{ p.numero }}
                <span class="ml-1 font-normal text-neutral-400">{{ dtBR(p.criadoEm) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="rounded-full px-2 py-0.5 text-[11px] font-medium"
                  :class="p.quitadoEm
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300'"
                >
                  {{ p.quitadoEm ? `Pago em ${dtBR(p.quitadoEm)}` : 'Em aberto' }}
                </span>
                <span class="text-base font-semibold text-neutral-900 dark:text-neutral-100">{{ money(p.total) }}</span>
              </div>
            </div>
            <ul v-if="p.itens.length" class="mt-2 space-y-0.5 border-t border-neutral-100 pt-2 text-sm text-neutral-600 dark:border-neutral-800 dark:text-neutral-300">
              <li v-for="it in p.itens" :key="it.id" class="flex items-center justify-between gap-2">
                <span class="min-w-0 truncate">{{ it.nome }} <span class="text-neutral-400">× {{ it.qtd }}</span></span>
                <span class="shrink-0 text-neutral-500 dark:text-neutral-400">{{ money(it.subtotal) }}</span>
              </li>
            </ul>
          </li>
        </ul>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton
            v-if="modalConta && modalConta.totalPendente > 0"
            color="primary"
            icon="i-lucide-check"
            :label="`Receber ${money(modalConta.totalPendente)}`"
            @click="modalConta && receber(modalConta)"
          />
          <UButton color="neutral" variant="outline" @click="fechar">Fechar</UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
