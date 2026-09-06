<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../lib/api'
import type { Evento } from '../lib/types'

const route = useRoute()
const router = useRouter()
const eventoId = computed(() => Number(route.params.eventoId))

const evento = ref<Evento | null>(null)
const nome = computed(() => evento.value?.nome ?? 'Evento')

const abas = computed(() => [
  { name: 'evento-dashboard', label: 'Dashboard', icon: 'i-lucide-layout-dashboard', to: `/eventos/${eventoId.value}` },
  { name: 'evento-produtos', label: 'Produtos', icon: 'i-lucide-package', to: `/eventos/${eventoId.value}/produtos` },
  { name: 'evento-contas', label: 'Anota aí', icon: 'i-lucide-notebook-pen', to: `/eventos/${eventoId.value}/contas` },
  { name: 'evento-pedidos', label: 'Pedidos', icon: 'i-lucide-clipboard-list', to: `/eventos/${eventoId.value}/pedidos` }
])

// Aba ativa pelo nome da rota resolvida (filha) — ex.: 'evento-produtos'.
const ativa = computed(() => route.name as string)

async function loadEvento() {
  try {
    const all = await api().ListEventos()
    evento.value = all.find((e) => e.id === eventoId.value) ?? null
  } catch {
    evento.value = null
  }
}

onMounted(loadEvento)
</script>

<template>
  <div class="flex h-full flex-col">
    <!-- Cabeçalho do evento -->
    <header class="border-b border-neutral-200 px-4 py-3 dark:border-neutral-800">
      <div class="mx-auto flex max-w-5xl items-center justify-between gap-3">
        <div class="flex min-w-0 items-center gap-3">
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-lucide-arrow-left"
            size="sm"
            title="Voltar para a lista de eventos"
            @click="router.push('/eventos')"
          />
          <div class="min-w-0">
            <div class="text-[11px] uppercase tracking-wide text-neutral-400">Evento</div>
            <h1 class="truncate text-lg font-semibold text-neutral-900 dark:text-neutral-100">{{ nome }}</h1>
          </div>
        </div>
        <UButton
          color="primary"
          icon="i-lucide-shopping-bag"
          @click="router.push(`/pdv/${eventoId}`)"
        >
          Ir para venda
        </UButton>
      </div>
    </header>

    <!-- Abas -->
    <nav class="border-b border-neutral-200 px-4 dark:border-neutral-800">
      <div class="mx-auto flex max-w-5xl gap-1">
        <RouterLink
          v-for="a in abas"
          :key="a.name"
          :to="a.to"
          class="flex items-center gap-2 border-b-2 px-3 py-2.5 text-sm font-medium transition
            focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
          :class="ativa === a.name
            ? 'border-emerald-600 text-emerald-700 dark:border-emerald-400 dark:text-emerald-300'
            : 'border-transparent text-neutral-500 hover:border-neutral-300 hover:text-neutral-800 dark:hover:border-neutral-700 dark:hover:text-neutral-200'"
        >
          <UIcon :name="a.icon" class="size-4" />
          {{ a.label }}
        </RouterLink>
      </div>
    </nav>

    <!-- Conteúdo da subpágina -->
    <main class="min-h-0 flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
