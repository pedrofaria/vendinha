<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'
import type { Evento } from '../lib/types'
import { dtBR } from '../lib/format'

const router = useRouter()

const eventos = ref<Evento[]>([])
const loading = ref(false)
const error = ref('')

// ---- modal de evento ----
const showModal = ref(false)
const editing = ref<Evento | null>(null)
const formNome = ref('')
const formVendeCartela = ref(false)

function openNew() {
  editing.value = null
  formNome.value = ''
  formVendeCartela.value = false
  showModal.value = true
}
function openEdit(e: Evento) {
  editing.value = e
  formNome.value = e.nome
  formVendeCartela.value = e.vendeCartela
  showModal.value = true
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    eventos.value = await api().ListEventos()
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  const nome = formNome.value.trim()
  if (!nome) return
  error.value = ''
  try {
    if (editing.value) {
      await api().UpdateEvento(editing.value.id, nome, editing.value.ativo, formVendeCartela.value)
    } else {
      await api().CreateEvento(nome, formVendeCartela.value)
    }
    showModal.value = false
    await load()
  } catch (e) {
    error.value = errMsg(e)
  }
}

async function toggleAtivo(e: Evento) {
  error.value = ''
  try {
    await api().UpdateEvento(e.id, e.nome, !e.ativo, e.vendeCartela)
    await load()
  } catch (err) {
    error.value = errMsg(err)
  }
}

async function del(e: Evento) {
  if (!confirm(`Excluir o evento "${e.nome}"? Produtos e vendas dele serão removidos.`)) return
  error.value = ''
  try {
    await api().DeleteEvento(e.id)
    await load()
  } catch (err) {
    error.value = errMsg(err)
  }
}

onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-4xl p-6">
    <div class="mb-5 flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold">Eventos</h1>
        <p class="text-sm text-neutral-500">Cadastre os eventos em que haverá venda.</p>
      </div>
      <UButton color="primary" icon="i-lucide-plus" size="md" @click="openNew">
        Novo evento
      </UButton>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <div v-else-if="eventos.length === 0" class="rounded-xl border border-dashed border-neutral-300 p-10 text-center text-neutral-500 dark:border-neutral-700">
      Nenhum evento ainda. Crie o primeiro para começar.
    </div>

    <ul v-else class="space-y-2">
      <li
        v-for="e in eventos"
        :key="e.id"
        class="flex items-center gap-3 rounded-lg border border-neutral-200 bg-white p-3 dark:border-neutral-800 dark:bg-neutral-900"
        :class="!e.ativo && 'opacity-60'"
      >
        <button class="min-w-0 flex-1 text-left" type="button" @click="router.push(`/eventos/${e.id}`)">
          <div class="truncate text-base font-medium text-neutral-900 dark:text-neutral-100">{{ e.nome }}</div>
          <div class="mt-0.5 flex items-center gap-2">
            <span class="text-xs text-neutral-400">
              {{ e.ativo ? 'Ativo' : 'Inativo' }} · criado em {{ dtBR(e.criadoEm) }}
            </span>
            <span v-if="e.vendeCartela"
              class="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:bg-emerald-950 dark:text-emerald-400">
              Cartelas
            </span>
          </div>
        </button>
        <div class="flex shrink-0 items-center gap-1.5">
          <UButton color="primary" variant="soft" icon="i-lucide-shopping-bag" size="sm" title="Vender neste evento"
            @click="router.push(`/pdv/${e.id}`)">
            Vender
          </UButton>
          <UButton color="primary" variant="outline" icon="i-lucide-package" size="sm" title="Gerenciar produtos"
            @click="router.push(`/eventos/${e.id}/produtos`)">
            Produtos
          </UButton>
          <UButton color="neutral" variant="ghost" :icon="e.ativo ? 'i-lucide-eye-off' : 'i-lucide-eye'" size="sm"
            :title="e.ativo ? 'Desativar' : 'Ativar'" @click="toggleAtivo(e)" />
          <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="sm" title="Renomear" @click="openEdit(e)" />
          <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" title="Excluir" @click="del(e)" />
        </div>
      </li>
    </ul>

    <!-- Modal novo/editar evento -->
    <UModal v-model:open="showModal" :title="editing ? 'Editar evento' : 'Novo evento'" :ui="{ content: 'max-w-md' }">
      <template #body>
        <form id="evento-form" class="space-y-4" @submit.prevent="save">
          <UFormField label="Nome do evento">
            <UInput v-model="formNome" placeholder="Ex.: Retiro de Carnaval 2026" size="lg" autofocus />
          </UFormField>
          <div class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 p-3 dark:border-neutral-800">
            <div class="min-w-0">
              <p class="text-sm font-medium text-neutral-900 dark:text-neutral-100">Vender cartelas</p>
              <p class="text-xs text-neutral-500">Cartelas de raspadinha (R$ 10, 20, 50 e 100) disponíveis no PDV deste evento.</p>
            </div>
            <USwitch v-model="formVendeCartela" aria-label="Vender cartelas neste evento" />
          </div>
        </form>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="outline" @click="showModal = false">Cancelar</UButton>
          <UButton color="primary" type="submit" form="evento-form" :disabled="!formNome.trim()">
            {{ editing ? 'Salvar' : 'Criar' }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
