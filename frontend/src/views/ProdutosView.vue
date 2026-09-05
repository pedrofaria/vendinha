<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'
import { centsToInput, inputToCents, money } from '../lib/format'
import type { Evento, EstoqueTipo, Produto } from '../lib/types'

const route = useRoute()
const router = useRouter()
const eventoId = computed(() => Number(route.params.eventoId))

const evento = ref<Evento | null>(null)
const produtos = ref<Produto[]>([])
const loading = ref(false)
const error = ref('')

// ---- modal produto ----
const showModal = ref(false)
const editing = ref<Produto | null>(null)
const form = ref({ nome: '', preco: '', estoqueTipo: 'ilimitado' as EstoqueTipo, quantidade: 0 })

function openNew() {
  editing.value = null
  form.value = { nome: '', preco: '', estoqueTipo: 'ilimitado', quantidade: 0 }
  showModal.value = true
}
function openEdit(p: Produto) {
  editing.value = p
  form.value = {
    nome: p.nome,
    preco: centsToInput(p.preco),
    estoqueTipo: p.estoqueTipo,
    quantidade: p.quantidade
  }
  showModal.value = true
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
    produtos.value = await api().ListProdutos(eventoId.value)
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  const nome = form.value.nome.trim()
  if (!nome) return
  const preco = inputToCents(form.value.preco)
  error.value = ''
  try {
    if (editing.value) {
      await api().UpdateProduto(editing.value.id, nome, preco, form.value.estoqueTipo, form.value.quantidade)
    } else {
      await api().CreateProduto(eventoId.value, nome, preco, form.value.estoqueTipo, form.value.quantidade)
    }
    showModal.value = false
    await loadProdutos()
  } catch (e) {
    error.value = errMsg(e)
  }
}

async function toggleAtivo(p: Produto) {
  try {
    await api().SetProdutoAtivo(p.id, !p.ativo)
    await loadProdutos()
  } catch (e) {
    error.value = errMsg(e)
  }
}

async function del(p: Produto) {
  if (!confirm(`Excluir o produto "${p.nome}"?`)) return
  try {
    await api().DeleteProduto(p.id)
    await loadProdutos()
  } catch (e) {
    error.value = errMsg(e)
  }
}

watch(eventoId, () => {
  loadEvento()
  loadProdutos()
})
onMounted(() => {
  loadEvento()
  loadProdutos()
})
</script>

<template>
  <div class="mx-auto max-w-5xl p-6">
    <div class="mb-5 flex items-center justify-between">
      <div>
        <button class="text-sm text-neutral-500 hover:text-emerald-600" type="button" @click="router.push('/eventos')">
          ← Eventos
        </button>
        <h1 class="text-xl font-semibold">{{ evento?.nome ?? 'Produtos' }}</h1>
        <p class="text-sm text-neutral-500">
          Produtos à venda neste evento. Preço em reais; o estoque pode ser infinito ou limitado.
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="primary" variant="soft" icon="i-lucide-shopping-bag" @click="router.push(`/pdv/${eventoId}`)">
          Ir para venda
        </UButton>
        <UButton color="primary" icon="i-lucide-plus" @click="openNew">Novo produto</UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <div v-else-if="produtos.length === 0" class="rounded-xl border border-dashed border-neutral-300 p-10 text-center text-neutral-500 dark:border-neutral-700">
      Nenhum produto neste evento. Adicione o primeiro.
    </div>

    <div v-else class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="p in produtos"
        :key="p.id"
        class="flex flex-col justify-between rounded-lg border border-neutral-200 bg-white p-4 dark:border-neutral-800 dark:bg-neutral-900"
        :class="!p.ativo && 'opacity-60'"
      >
        <div>
          <div class="flex items-start justify-between gap-2">
            <h3 class="font-medium text-neutral-900 dark:text-neutral-100">{{ p.nome }}</h3>
            <span
              class="rounded-full px-2 py-0.5 text-xs font-medium"
              :class="p.ativo
                ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                : 'bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400'"
            >
              {{ p.ativo ? 'Ativo' : 'Inativo' }}
            </span>
          </div>
          <div class="mt-1 text-xl font-semibold text-neutral-900 dark:text-neutral-100">{{ money(p.preco) }}</div>
          <div class="mt-0.5 text-sm text-neutral-500 dark:text-neutral-400">
            <template v-if="p.estoqueTipo === 'ilimitado'">Estoque infinito</template>
            <template v-else>{{ p.quantidade }} em estoque</template>
          </div>
        </div>
        <div class="mt-3 flex justify-end gap-1">
          <UButton color="neutral" variant="ghost" :icon="p.ativo ? 'i-lucide-eye-off' : 'i-lucide-eye'" size="sm" @click="toggleAtivo(p)" />
          <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="sm" @click="openEdit(p)" />
          <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" @click="del(p)" />
        </div>
      </div>
    </div>

    <UModal v-model:open="showModal" :title="editing ? 'Editar produto' : 'Novo produto'" :ui="{ content: 'max-w-lg' }">
      <template #body>
        <form id="produto-form" class="space-y-4" @submit.prevent="save">
          <UFormField label="Nome do produto">
            <UInput v-model="form.nome" placeholder="Ex.: Pastel de carne" size="lg" autofocus />
          </UFormField>
          <div class="grid grid-cols-2 gap-3">
            <UFormField label="Preço (R$)">
              <UInput v-model="form.preco" inputmode="decimal" placeholder="0,00" size="lg" />
            </UFormField>
            <UFormField label="Tipo de estoque">
              <USelect
                v-model="form.estoqueTipo"
                :items="[
                  { label: 'Infinito', value: 'ilimitado' },
                  { label: 'Limitado', value: 'limitado' }
                ]"
                value-key="value"
                size="lg"
              />
            </UFormField>
          </div>
          <UFormField v-if="form.estoqueTipo === 'limitado'" label="Quantidade em estoque">
            <UInput v-model.number="form.quantidade" type="number" min="0" size="lg" />
          </UFormField>
        </form>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="outline" @click="showModal = false">Cancelar</UButton>
          <UButton color="primary" type="submit" form="produto-form" :disabled="!form.nome.trim()">
            {{ editing ? 'Salvar' : 'Adicionar' }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
