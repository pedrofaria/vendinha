<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import draggable from 'vuedraggable'
import { api, errMsg } from '../lib/api'
import { centsToInput, inputToCents, money } from '../lib/format'
import { SemGrupo, type EstoqueTipo, type Grupo, type GrupoProdutos, type Produto } from '../lib/types'

// Paleta sugerida ao criar/editar um grupo.
const PALETA = ['#ef4444', '#f97316', '#f59e0b', '#84cc16', '#10b981', '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899', '#64748b']

const route = useRoute()
const eventoId = computed(() => Number(route.params.eventoId))

// Grupos reais (ordenados) + o bucket virtual "Sem grupo" separado (fica sempre por último).
const reaisArr = ref<GrupoProdutos[]>([])
const semBucket = ref<GrupoProdutos | null>(null)
const loading = ref(false)
const error = ref('')

function split(grupos: GrupoProdutos[]) {
  reaisArr.value = grupos.filter((g) => g.id !== SemGrupo)
  semBucket.value = grupos.find((g) => g.id === SemGrupo) ?? null
}

// ---- modal grupo ----
const showGrupoModal = ref(false)
const editingGrupo = ref<Grupo | null>(null)
const grupoForm = ref({ nome: '', cor: PALETA[0] })

function novoGrupo() {
  editingGrupo.value = null
  grupoForm.value = { nome: '', cor: PALETA[reaisArr.value.length % PALETA.length] }
  showGrupoModal.value = true
}
function editarGrupo(g: Grupo) {
  editingGrupo.value = g
  grupoForm.value = { nome: g.nome, cor: g.cor }
  showGrupoModal.value = true
}
async function saveGrupo() {
  const nome = grupoForm.value.nome.trim()
  if (!nome) return
  error.value = ''
  try {
    if (editingGrupo.value) {
      await api().UpdateGrupo(editingGrupo.value.id, nome, grupoForm.value.cor)
    } else {
      await api().CreateGrupo(eventoId.value, nome, grupoForm.value.cor)
    }
    showGrupoModal.value = false
    await loadProdutos()
  } catch (e) {
    error.value = errMsg(e)
  }
}
async function delGrupo(g: Grupo) {
  if (!confirm(`Excluir o grupo "${g.nome}"? Os produtos dele vão para "Sem grupo".`)) return
  try {
    await api().DeleteGrupo(g.id)
    await loadProdutos()
  } catch (e) {
    error.value = errMsg(e)
  }
}

// ---- modal produto ----
const showProdutoModal = ref(false)
const editing = ref<Produto | null>(null)
const form = ref({ nome: '', preco: '', estoqueTipo: 'ilimitado' as EstoqueTipo, quantidade: 0, grupoId: SemGrupo, atalho: 0 })

const opcoesGrupo = computed(() => {
  const opcoes: { label: string; value: number }[] = [{ label: 'Sem grupo', value: SemGrupo }]
  for (const g of reaisArr.value) opcoes.push({ label: g.nome, value: g.id })
  return opcoes
})

// Teclas (1-9) já em uso por produtos do evento, para barrar conflito no cadastro.
const teclaOcupada = computed<Record<number, Produto>>(() => {
  const mapa: Record<number, Produto> = {}
  const todos: Produto[] = reaisArr.value.flatMap((g) => g.produtos)
  if (semBucket.value) todos.push(...semBucket.value.produtos)
  for (const p of todos) if (p.atalho > 0 && !(p.atalho in mapa)) mapa[p.atalho] = p
  return mapa
})

const TECLAS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

// True quando a tecla pode ser escolhida: livre ou já pertence ao produto em edição.
function teclaDisponivel(n: number): boolean {
  const dono = teclaOcupada.value[n]
  if (!dono) return true
  return !!editing.value && dono.id === editing.value.id
}

function openNew(grupoId: number = SemGrupo) {
  editing.value = null
  form.value = { nome: '', preco: '', estoqueTipo: 'ilimitado', quantidade: 0, grupoId, atalho: 0 }
  showProdutoModal.value = true
}
function openEdit(p: Produto) {
  editing.value = p
  form.value = {
    nome: p.nome,
    preco: centsToInput(p.preco),
    estoqueTipo: p.estoqueTipo,
    quantidade: p.quantidade,
    grupoId: p.grupoId,
    atalho: p.atalho
  }
  showProdutoModal.value = true
}
async function saveProduto() {
  const nome = form.value.nome.trim()
  if (!nome) return
  const preco = inputToCents(form.value.preco)
  error.value = ''
  try {
    if (editing.value) {
      await api().UpdateProduto(editing.value.id, nome, preco, form.value.estoqueTipo, form.value.quantidade, form.value.grupoId, form.value.atalho)
    } else {
      await api().CreateProduto(eventoId.value, nome, preco, form.value.estoqueTipo, form.value.quantidade, form.value.grupoId, form.value.atalho)
    }
    showProdutoModal.value = false
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

// ---- drag & drop: persistir a ordem após soltar ----
// vuedraggable já reordena o array local; aqui gravamos a ordem final no banco.
async function onEndGrupos() {
  const ids = reaisArr.value.map((g) => g.id)
  try {
    await api().ReorderGrupos(eventoId.value, ids)
  } catch (e) {
    error.value = errMsg(e)
    await loadProdutos() // reverte para a ordem do banco
  }
}
async function onEndProdutos(produtos: Produto[]) {
  const ids = produtos.map((p) => p.id)
  try {
    await api().ReorderProdutos(ids)
  } catch (e) {
    error.value = errMsg(e)
    await loadProdutos() // reverte para a ordem do banco
  }
}

async function loadProdutos() {
  loading.value = true
  error.value = ''
  try {
    split(await api().ListProdutos(eventoId.value))
  } catch (e) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

watch(eventoId, loadProdutos)
onMounted(loadProdutos)
</script>

<template>
  <div class="mx-auto max-w-5xl p-6">
    <div class="mb-5 flex flex-wrap items-start justify-between gap-3">
      <div class="max-w-xl">
        <h2 class="text-xl font-semibold">Produtos</h2>
        <p class="text-sm text-neutral-500">
          Produtos organizados em grupos. A cor do grupo identifica os itens na tela de venda (PDV).
          Arraste a alça ⋮⋮ para reordenar grupos e produtos.
        </p>
      </div>
      <div class="flex flex-wrap justify-end gap-2">
        <UButton color="neutral" variant="soft" icon="i-lucide-folder-plus" @click="novoGrupo">
          Novo grupo
        </UButton>
        <UButton color="primary" icon="i-lucide-plus" @click="openNew()">
          Novo produto
        </UButton>
      </div>
    </div>

    <UAlert v-if="error" color="error" icon="i-lucide-alert-circle" :title="error" class="mb-4" />

    <div v-if="loading" class="flex justify-center py-10">
      <UIcon name="i-lucide-loader-circle" class="size-6 animate-spin text-neutral-400" />
    </div>

    <div v-else-if="reaisArr.length === 0 && !semBucket" class="rounded-xl border border-dashed border-neutral-300 p-10 text-center text-neutral-500 dark:border-neutral-700">
      <p class="text-lg font-medium">Nenhum produto neste evento.</p>
      <p class="mt-1 text-sm">Crie um grupo (ex.: Comida, Bebida, Sobremesa) e adicione produtos a ele.</p>
      <div class="mt-4 flex justify-center gap-2">
        <UButton color="neutral" variant="soft" icon="i-lucide-folder-plus" @click="novoGrupo">Novo grupo</UButton>
        <UButton color="primary" icon="i-lucide-plus" @click="openNew()">Novo produto</UButton>
      </div>
    </div>

    <!-- Lista real de grupos: arrastável (sem incluir o bucket "Sem grupo") -->
    <draggable
      v-else
      v-model="reaisArr"
      item-key="id"
      handle=".drag-grupo"
      :animation="150"
      class="space-y-6"
      @end="onEndGrupos"
    >
      <template #item="{ element: g }">
        <section class="overflow-hidden rounded-xl border border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900">
          <!-- cabeçalho do grupo -->
          <div class="flex items-center gap-2 border-b border-neutral-100 px-2 py-2 dark:border-neutral-800">
            <button type="button" class="drag-grupo cursor-grab rounded p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600 active:cursor-grabbing dark:hover:bg-neutral-800 dark:hover:text-neutral-300" title="Arrastar para reordenar o grupo">
              <UIcon name="i-lucide-grip-vertical" class="size-4" />
            </button>
            <span class="size-3 shrink-0 rounded-full ring-1 ring-black/5" :style="{ backgroundColor: g.cor }" :title="g.cor" />
            <h2 class="truncate text-sm font-semibold text-neutral-900 dark:text-neutral-100">{{ g.nome }}</h2>
            <span class="text-xs text-neutral-400">{{ g.produtos.length }} {{ g.produtos.length === 1 ? 'item' : 'itens' }}</span>

            <div class="ml-auto flex items-center gap-1">
              <UButton color="neutral" variant="ghost" icon="i-lucide-plus" size="xs" :title="`Adicionar produto em ${g.nome}`" @click="openNew(g.id)" />
              <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="xs" :title="`Editar grupo ${g.nome}`" @click="editarGrupo(g)" />
              <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="xs" :title="`Excluir grupo ${g.nome}`" @click="delGrupo(g)" />
            </div>
          </div>

          <!-- produtos do grupo (arrastáveis dentro do grupo) -->
          <draggable
            v-if="g.produtos.length"
            v-model="g.produtos"
            item-key="id"
            handle=".drag-produto"
            :animation="150"
            class="grid grid-cols-1 gap-2 p-3 sm:grid-cols-2 lg:grid-cols-3"
            @end="onEndProdutos(g.produtos)"
          >
            <template #item="{ element: p }">
              <div class="relative flex flex-col justify-between overflow-hidden rounded-lg border border-neutral-200 bg-white pl-5 dark:border-neutral-800 dark:bg-neutral-900" :class="!p.ativo && 'opacity-60'">
                <span class="absolute inset-y-0 left-0 w-1.5" :style="{ backgroundColor: g.cor }" />
                <div class="p-3 pl-4">
                  <div class="flex items-start justify-between gap-2">
                    <h3 class="font-medium text-neutral-900 dark:text-neutral-100">{{ p.nome }}</h3>
                    <div class="flex items-center gap-1.5">
                      <span
                        v-if="p.atalho > 0"
                        class="flex h-5 min-w-5 items-center justify-center rounded-md bg-neutral-900 px-1.5 text-xs font-bold text-white dark:bg-neutral-200 dark:text-neutral-900"
                        :title="`Tecla ${p.atalho}: aperte no PDV para adicionar`"
                      >{{ p.atalho }}</span>
                      <span
                        class="rounded-full px-2 py-0.5 text-xs font-medium"
                        :class="p.ativo
                          ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                          : 'bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400'"
                      >
                        {{ p.ativo ? 'Ativo' : 'Inativo' }}
                      </span>
                    </div>
                  </div>
                  <div class="mt-1 text-xl font-semibold text-neutral-900 dark:text-neutral-100">{{ money(p.preco) }}</div>
                  <div class="mt-0.5 text-sm text-neutral-500 dark:text-neutral-400">
                    <template v-if="p.estoqueTipo === 'ilimitado'">Estoque infinito</template>
                    <template v-else>{{ p.quantidade }} em estoque</template>
                  </div>
                </div>
                <div class="flex items-center justify-between gap-1 border-t border-neutral-100 px-2 py-1 dark:border-neutral-800">
                  <button type="button" class="drag-produto cursor-grab rounded p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600 active:cursor-grabbing dark:hover:bg-neutral-800 dark:hover:text-neutral-300" title="Arrastar para reordenar">
                    <UIcon name="i-lucide-grip-vertical" class="size-4" />
                  </button>
                  <div class="flex gap-1">
                    <UButton color="neutral" variant="ghost" :icon="p.ativo ? 'i-lucide-eye-off' : 'i-lucide-eye'" size="sm" title="Ativar/desativar" @click="toggleAtivo(p)" />
                    <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="sm" title="Editar" @click="openEdit(p)" />
                    <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" title="Excluir" @click="del(p)" />
                  </div>
                </div>
              </div>
            </template>
          </draggable>
          <div v-else class="px-4 py-3 text-sm text-neutral-400">
            Nenhum produto neste grupo.
            <button type="button" class="font-medium text-emerald-600 hover:underline" @click="openNew(g.id)">Adicionar</button>
          </div>
        </section>
      </template>
    </draggable>

    <!-- Bucket virtual "Sem grupo" (fixo por último, sem reordenar grupos aqui) -->
    <section v-if="semBucket" class="mt-6 overflow-hidden rounded-xl border border-dashed border-neutral-300 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-900/40">
      <div class="flex items-center gap-2 border-b border-neutral-200/70 px-3 py-2 dark:border-neutral-800">
        <span class="size-3 shrink-0 rounded-full ring-1 ring-black/5" :style="{ backgroundColor: semBucket.cor }" />
        <h2 class="text-sm font-semibold text-neutral-600 dark:text-neutral-300">{{ semBucket.nome }}</h2>
        <span class="text-xs text-neutral-400">{{ semBucket.produtos.length }} {{ semBucket.produtos.length === 1 ? 'item' : 'itens' }}</span>
        <span class="ml-auto text-xs text-neutral-400">Produtos ainda não organizados em grupo</span>
      </div>

      <draggable
        v-if="semBucket.produtos.length"
        v-model="semBucket.produtos"
        item-key="id"
        handle=".drag-produto"
        :animation="150"
        class="grid grid-cols-1 gap-2 p-3 sm:grid-cols-2 lg:grid-cols-3"
        @end="onEndProdutos(semBucket.produtos)"
      >
        <template #item="{ element: p }">
          <div class="relative flex flex-col justify-between overflow-hidden rounded-lg border border-neutral-200 bg-white pl-5 dark:border-neutral-800 dark:bg-neutral-900" :class="!p.ativo && 'opacity-60'">
            <span class="absolute inset-y-0 left-0 w-1.5" :style="{ backgroundColor: semBucket.cor }" />
            <div class="p-3 pl-4">
              <div class="flex items-start justify-between gap-2">
                <h3 class="font-medium text-neutral-900 dark:text-neutral-100">{{ p.nome }}</h3>
                <div class="flex items-center gap-1.5">
                  <span
                    v-if="p.atalho > 0"
                    class="flex h-5 min-w-5 items-center justify-center rounded-md bg-neutral-900 px-1.5 text-xs font-bold text-white dark:bg-neutral-200 dark:text-neutral-900"
                    :title="`Tecla ${p.atalho}: aperte no PDV para adicionar`"
                  >{{ p.atalho }}</span>
                  <span
                    class="rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="p.ativo
                      ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                      : 'bg-neutral-100 text-neutral-500 dark:bg-neutral-800 dark:text-neutral-400'"
                  >
                    {{ p.ativo ? 'Ativo' : 'Inativo' }}
                  </span>
                </div>
              </div>
              <div class="mt-1 text-xl font-semibold text-neutral-900 dark:text-neutral-100">{{ money(p.preco) }}</div>
              <div class="mt-0.5 text-sm text-neutral-500 dark:text-neutral-400">
                <template v-if="p.estoqueTipo === 'ilimitado'">Estoque infinito</template>
                <template v-else>{{ p.quantidade }} em estoque</template>
              </div>
            </div>
            <div class="flex items-center justify-between gap-1 border-t border-neutral-100 px-2 py-1 dark:border-neutral-800">
              <button type="button" class="drag-produto cursor-grab rounded p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-600 active:cursor-grabbing dark:hover:bg-neutral-800 dark:hover:text-neutral-300" title="Arrastar para reordenar">
                <UIcon name="i-lucide-grip-vertical" class="size-4" />
              </button>
              <div class="flex gap-1">
                <UButton color="neutral" variant="ghost" :icon="p.ativo ? 'i-lucide-eye-off' : 'i-lucide-eye'" size="sm" title="Ativar/desativar" @click="toggleAtivo(p)" />
                <UButton color="neutral" variant="ghost" icon="i-lucide-pencil" size="sm" title="Editar" @click="openEdit(p)" />
                <UButton color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" title="Excluir" @click="del(p)" />
              </div>
            </div>
          </div>
        </template>
      </draggable>
      <div v-else class="px-4 py-3 text-sm text-neutral-400">
        Sem produtos por aqui.
        <button type="button" class="font-medium text-emerald-600 hover:underline" @click="openNew()">Adicionar</button>
      </div>
    </section>

    <!-- Modal grupo -->
    <UModal v-model:open="showGrupoModal" :title="editingGrupo ? 'Editar grupo' : 'Novo grupo'" :ui="{ content: 'max-w-sm' }">
      <template #body>
        <form id="grupo-form" class="space-y-4" @submit.prevent="saveGrupo">
          <UFormField label="Nome do grupo">
            <UInput v-model="grupoForm.nome" placeholder="Ex.: Comida" size="lg" autofocus />
          </UFormField>
          <UFormField label="Cor de identificação (usada no PDV)">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="c in PALETA"
                :key="c"
                type="button"
                class="size-8 rounded-full ring-2 ring-offset-2 transition"
                :class="grupoForm.cor === c ? 'ring-emerald-500' : 'ring-transparent hover:ring-neutral-300'"
                :style="{ backgroundColor: c }"
                :title="c"
                @click="grupoForm.cor = c"
              />
            </div>
          </UFormField>
        </form>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="outline" @click="showGrupoModal = false">Cancelar</UButton>
          <UButton color="primary" type="submit" form="grupo-form" :disabled="!grupoForm.nome.trim()">
            {{ editingGrupo ? 'Salvar' : 'Criar grupo' }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- Modal produto -->
    <UModal v-model:open="showProdutoModal" :title="editing ? 'Editar produto' : 'Novo produto'" :ui="{ content: 'max-w-lg' }">
      <template #body>
        <form id="produto-form" class="space-y-4" @submit.prevent="saveProduto">
          <UFormField label="Nome do produto">
            <UInput v-model="form.nome" placeholder="Ex.: Pastel de carne" size="lg" autofocus />
          </UFormField>
          <UFormField label="Grupo">
            <USelect
              v-model="form.grupoId"
              :items="opcoesGrupo"
              value-key="value"
              size="lg"
            />
          </UFormField>
          <UFormField label="Atalho (tecla 1-9) — opcional">
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="n in TECLAS"
                :key="n"
                type="button"
                :disabled="n !== 0 && !teclaDisponivel(n)"
                :title="n === 0
                  ? 'Sem atalho: o produto entra só pelo toque'
                  : teclaDisponivel(n)
                    ? `Tecla ${n}: aperte ${n} no PDV para adicionar`
                    : `Tecla ${n} já em uso por “${teclaOcupada[n]?.nome}”`"
                class="flex h-10 min-w-10 items-center justify-center gap-1 rounded-lg border px-2 text-sm font-medium transition
                  focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500"
                :class="form.atalho === n
                  ? 'border-emerald-500 bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-400'
                  : n !== 0 && !teclaDisponivel(n)
                    ? 'cursor-not-allowed border-neutral-200 bg-neutral-50 text-neutral-300 dark:border-neutral-800 dark:bg-neutral-900 dark:text-neutral-600'
                    : 'border-neutral-200 text-neutral-600 hover:border-neutral-300 dark:border-neutral-700 dark:text-neutral-300'"
                @click="form.atalho = n"
              >
                <template v-if="n === 0">Nenhum</template>
                <template v-else>{{ n }}</template>
              </button>
            </div>
            <p class="mt-1 text-xs text-neutral-400">
              Com atalho, a tecla adiciona o produto direto no pedido. Cada tecla só pode ser usada por um produto por evento.
            </p>
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
          <UButton color="neutral" variant="outline" @click="showProdutoModal = false">Cancelar</UButton>
          <UButton color="primary" type="submit" form="produto-form" :disabled="!form.nome.trim()">
            {{ editing ? 'Salvar' : 'Adicionar' }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
