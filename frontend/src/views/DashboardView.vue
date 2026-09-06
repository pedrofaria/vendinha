<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { money } from '../lib/format'
import type { ResumoEvento } from '../lib/types'

const route = useRoute()
const eventoId = computed(() => Number(route.params.eventoId))

const resumo = ref<ResumoEvento | null>(null)
const loading = ref(false)
const error = ref('')

const horas = computed(() => resumo.value?.vendasPorHora ?? [])
const maxHora = computed(() => Math.max(1, ...horas.value.map((h) => h.vendas)))
const totalVendas = computed(() => horas.value.reduce((s, h) => s + h.vendas, 0))

async function load() {
  loading.value = true
  error.value = ''
  try {
    resumo.value = await api().ResumoEvento(eventoId.value)
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

watch(eventoId, load)
onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-5xl p-6">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold">Dashboard</h2>
        <p class="text-sm text-neutral-500">Resumo das vendas fechadas deste evento.</p>
      </div>
      <UButton color="neutral" variant="soft" icon="i-lucide-refresh-cw" title="Atualizar" @click="load" />
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <template v-else-if="resumo">
      <!-- Cartões-resumo -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
          <div class="flex items-center gap-2 text-sm text-neutral-500 dark:text-neutral-400">
            <UIcon name="i-lucide-wallet" class="size-4 text-emerald-600 dark:text-emerald-400" />
            Total de vendas (R$)
          </div>
          <div class="mt-1 text-2xl font-semibold text-emerald-700 dark:text-emerald-400">{{ money(resumo.receitaTotal) }}</div>
        </div>
        <div class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
          <div class="flex items-center gap-2 text-sm text-neutral-500 dark:text-neutral-400">
            <UIcon name="i-lucide-package" class="size-4 text-emerald-600 dark:text-emerald-400" />
            Produtos vendidos
          </div>
          <div class="mt-1 text-2xl font-semibold text-neutral-900 dark:text-neutral-100">{{ resumo.numProdutosVendidos }}</div>
          <div class="text-xs text-neutral-400">unidades (sem cartelas)</div>
        </div>
        <div class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
          <div class="flex items-center gap-2 text-sm text-neutral-500 dark:text-neutral-400">
            <UIcon name="i-lucide-receipt-text" class="size-4 text-emerald-600 dark:text-emerald-400" />
            Nº de vendas
          </div>
          <div class="mt-1 text-2xl font-semibold text-neutral-900 dark:text-neutral-100">{{ resumo.numPedidos }}</div>
          <div class="text-xs text-neutral-400">pedidos fechados</div>
        </div>
      </div>

      <!-- Vendas por hora -->
      <div class="mt-6 rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
        <div class="mb-1 flex items-baseline justify-between gap-2">
          <h3 class="text-sm font-semibold text-neutral-900 dark:text-neutral-100">Vendas por hora</h3>
          <span class="text-xs text-neutral-400">total: {{ totalVendas }} {{ totalVendas === 1 ? 'venda' : 'vendas' }}</span>
        </div>

        <div v-if="totalVendas === 0" class="rounded-lg border border-dashed border-neutral-300 p-8 text-center text-sm text-neutral-400 dark:border-neutral-700">
          Nenhuma venda registrada ainda. As vendas aparecerão aqui por hora do dia.
        </div>

        <div v-else class="flex h-44 items-end gap-0.5 pt-4">
          <div
            v-for="h in horas"
            :key="h.hora"
            class="group relative flex h-full flex-1 flex-col items-center justify-end"
            :title="`${String(h.hora).padStart(2, '0')}h: ${h.vendas} ${h.vendas === 1 ? 'venda' : 'vendas'}`"
          >
            <div
              class="w-full rounded-t transition-colors"
              :class="h.vendas === maxHora && h.vendas > 0
                ? 'bg-emerald-500 dark:bg-emerald-400'
                : 'bg-emerald-200 hover:bg-emerald-300 dark:bg-emerald-900/70 dark:hover:bg-emerald-700'"
              :style="{ height: `${(h.vendas / maxHora) * 100}%` }"
            />
          </div>
        </div>

        <div class="mt-1 flex justify-between text-[10px] text-neutral-400">
          <span>0h</span>
          <span>6h</span>
          <span>12h</span>
          <span>18h</span>
          <span>23h</span>
        </div>
      </div>
    </template>
  </div>
</template>
