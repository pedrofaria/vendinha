<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, errMsg } from '../lib/api'

const router = useRouter()

// ================= Banco de dados =================
const dbPath = ref('')
const dbError = ref('')
const abrindoPasta = ref(false)
const copiado = ref(false)

async function loadDb() {
  dbError.value = ''
  try {
    dbPath.value = await api().GetDBPath()
  } catch (e) {
    dbError.value = errMsg(e)
  }
}

async function abrirPasta() {
  abrindoPasta.value = true
  dbError.value = ''
  try {
    await api().OpenDBFolder()
  } catch (e) {
    dbError.value = errMsg(e)
  } finally {
    abrindoPasta.value = false
  }
}

async function copiarPath() {
  try {
    await navigator.clipboard.writeText(dbPath.value)
    copiado.value = true
    setTimeout(() => (copiado.value = false), 2000)
  } catch {
    /* clipboard indisponível no webview: sem ação */
  }
}

// ---- Zerar banco (exige digitar "confirmo") ----
const showZerar = ref(false)
const confirmo = ref('')
const zerando = ref(false)
const zerarError = ref('')

const podeZerar = computed(() => confirmo.value.trim().toLowerCase() === 'confirmo')

function abrirModalZerar() {
  confirmo.value = ''
  zerarError.value = ''
  showZerar.value = true
}

async function zerar() {
  zerando.value = true
  zerarError.value = ''
  try {
    await api().ZerarBanco()
    showZerar.value = false
    // A lista de eventos agora está vazia — mostra a "tela limpa".
    router.push('/eventos')
  } catch (e) {
    zerarError.value = errMsg(e)
  } finally {
    zerando.value = false
  }
}

// ================= Impressora =================
const impNames = ref<string[]>([])
const impressora = ref('')
const impPadrao = ref('')
const carregandoImp = ref(false)
const carregadoImp = ref(false)
const salvandoImp = ref(false)
const impError = ref('')
const impSalvo = ref('')
let pausaWatcher = false

async function loadImpressoras() {
  carregandoImp.value = true
  impError.value = ''
  carregadoImp.value = false
  try {
    const info = await api().ListImpressoras()
    impNames.value = info.nomes ?? []
    impPadrao.value = info.padrao ?? ''
    impressora.value = info.selecionada ?? ''
  } catch (e) {
    impError.value = errMsg(e)
  } finally {
    carregadoImp.value = true
    carregandoImp.value = false
  }
}

// Ao trocar a impressora no select, grava a escolha (config.json).
watch(impressora, (nova, velha) => {
  if (pausaWatcher || !carregadoImp.value || nova === velha) return
  salvarImpressora(nova, velha)
})

async function salvarImpressora(nome: string, velha: string) {
  salvandoImp.value = true
  impSalvo.value = ''
  impError.value = ''
  try {
    await api().SetImpressora(nome)
    impSalvo.value = nome ? `Impressora definida: ${nome}` : 'Usando a impressora padrão do sistema.'
  } catch (e) {
    impError.value = errMsg(e)
    // Volta à seleção anterior sem religar o watcher (evita loop em erro).
    pausaWatcher = true
    impressora.value = velha
    pausaWatcher = false
  } finally {
    salvandoImp.value = false
  }
}

onMounted(() => {
  loadDb()
  loadImpressoras()
  loadLargura()
  loadDebug()
})

// ================= Modo debug (não imprime; mostra recibo na tela) =================
const debug = ref(false)
const carregadoDebug = ref(false)
const salvandoDebug = ref(false)
const debugError = ref('')
let pausaDebug = false

async function loadDebug() {
  debugError.value = ''
  try {
    debug.value = await api().GetModoDebug()
  } catch (e) {
    debugError.value = errMsg(e)
  } finally {
    carregadoDebug.value = true
  }
}

// Grava a escolha ao alternar (global, tabela config). Reverte em erro sem religar o watcher.
watch(debug, (nova, velha) => {
  if (pausaDebug || !carregadoDebug.value || nova === velha) return
  gravarDebug(nova, velha)
})

async function gravarDebug(nova: boolean, velha: boolean) {
  salvandoDebug.value = true
  debugError.value = ''
  try {
    await api().SetModoDebug(nova)
  } catch (e) {
    debugError.value = errMsg(e)
    pausaDebug = true
    debug.value = velha
    pausaDebug = false
  } finally {
    salvandoDebug.value = false
  }
}

// ================= Largura da linha da impressão =================
const larguraTxt = ref('')
const carregadoLargura = ref(false)
const salvandoLargura = ref(false)
const larguraError = ref('')
const larguraSalvo = ref('')
let larguraInicial = 0

async function loadLargura() {
  larguraError.value = ''
  larguraSalvo.value = ''
  carregadoLargura.value = false
  try {
    const v = await api().GetLarguraLinha()
    larguraInicial = v
    larguraTxt.value = String(v)
  } catch (e) {
    larguraError.value = errMsg(e)
  } finally {
    carregadoLargura.value = true
  }
}

function larguraValida(): number | null {
  // LarguraTxt pode vir do UInput como string OU número (type="number"): normaliza
  // antes de parsear, senão um TypeError em .replace estoura o handler sem feedback.
  const n = parseInt(String(larguraTxt.value ?? '').replace(/\D/g, ''), 10)
  if (isNaN(n) || n < 1) return null
  return n
}

function aoDigitarLargura() {
  larguraError.value = ''
  larguraSalvo.value = ''
}

async function salvarLargura() {
  const n = larguraValida()
  if (n === null) {
    larguraError.value = 'Informe um número de caracteres por linha maior que zero (ex.: 32).'
    return
  }
  salvandoLargura.value = true
  larguraError.value = ''
  larguraSalvo.value = ''
  try {
    await api().SetLarguraLinha(n)
    larguraInicial = n
    larguraSalvo.value = `${n} caracteres por linha.`
  } catch (e) {
    larguraError.value = errMsg(e)
    larguraTxt.value = String(larguraInicial)
  } finally {
    salvandoLargura.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-6 p-6">
    <div>
      <h1 class="text-xl font-semibold">Configurações</h1>
      <p class="text-sm text-neutral-500">Banco de dados e a impressora usada para imprimir.</p>
    </div>

    <!-- ============ Banco de dados ============ -->
    <section class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
      <div class="mb-4 flex items-start gap-3">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-emerald-100 dark:bg-emerald-950">
          <UIcon name="i-lucide-database" class="size-5 text-emerald-700 dark:text-emerald-400" />
        </div>
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Banco de dados</h2>
          <p class="text-sm text-neutral-500">
            Onde o Vendinha guarda todos os dados. Faça backup copiando este arquivo.
          </p>
        </div>
      </div>

      <UAlert v-if="dbError" color="error" icon="i-lucide-alert-triangle" :title="dbError" class="mb-4" />

      <div class="mb-4">
        <p class="mb-1.5 text-xs font-medium text-neutral-500">Arquivo de dados</p>
        <div class="flex items-center gap-2">
          <code
            class="min-w-0 flex-1 break-all rounded-lg border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs leading-relaxed text-neutral-700 select-all dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-300"
            :title="dbPath"
          >
            {{ dbPath || 'Carregando…' }}
          </code>
          <UButton
            color="neutral"
            variant="soft"
            size="sm"
            icon="i-lucide-copy"
            :title="copiado ? 'Copiado!' : 'Copiar caminho'"
            :disabled="!dbPath"
            @click="copiarPath"
          />
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <UButton
          color="primary"
          variant="soft"
          icon="i-lucide-folder-open"
          :loading="abrindoPasta"
          @click="abrirPasta"
        >
          Abrir pasta do banco
        </UButton>
        <UButton
          color="error"
          variant="outline"
          icon="i-lucide-trash-2"
          @click="abrirModalZerar"
        >
          Zerar banco
        </UButton>
      </div>
      <p class="mt-3 text-xs text-neutral-400">
        <span class="font-medium text-red-500 dark:text-red-400">Zerar</span> apaga todos os eventos,
        pedidos, produtos, grupos e contas do "Anota aí" e recomeça a numeração do 1. Essa ação é
        permanente e não tem desfazer.
      </p>
    </section>

    <!-- ============ Impressora ============ -->
    <section class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
      <div class="mb-4 flex items-start gap-3">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-emerald-100 dark:bg-emerald-950">
          <UIcon name="i-lucide-printer" class="size-5 text-emerald-700 dark:text-emerald-400" />
        </div>
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Impressora</h2>
          <p class="text-sm text-neutral-500">Escolha a impressora em que os recibos e cartelas serão impressos.</p>
        </div>
      </div>

      <UAlert v-if="impError" color="error" icon="i-lucide-alert-triangle" :title="impError" class="mb-4" />

      <div v-if="!carregandoImp && impNames.length === 0 && !impError" class="mb-4">
        <UAlert
          color="warning"
          icon="i-lucide-printer-off"
          title="Nenhuma impressora encontrada"
          description="Conecte ou instale uma impressora no sistema e clique em Atualizar."
        />
      </div>

      <div v-if="impNames.length" class="flex flex-wrap items-end gap-3">
        <UFormField label="Impressora para impressão" class="min-w-64 flex-1">
          <USelect
            v-model="impressora"
            :items="impNames"
            :loading="carregandoImp"
            :disabled="salvandoImp"
            placeholder="Escolha a impressora…"
            size="lg"
          />
        </UFormField>
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-refresh-cw"
          :loading="carregandoImp"
          title="Atualizar lista de impressoras"
          @click="loadImpressoras"
        >
          Atualizar
        </UButton>
      </div>

      <div v-else class="flex items-center gap-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-refresh-cw"
          :loading="carregandoImp"
          @click="loadImpressoras"
        >
          Atualizar
        </UButton>
      </div>

      <!-- Largura da linha do papel térmico -->
      <div class="mt-5 border-t border-neutral-100 pt-4 dark:border-neutral-800">
        <div class="flex flex-wrap items-end gap-3">
          <UFormField label="Caracteres por linha" class="min-w-44">
            <UInput
              v-model="larguraTxt"
              type="number"
              min="1"
              max="200"
              step="1"
              inputmode="numeric"
              class="w-36"
              size="lg"
              aria-label="Caracteres por linha da impressão"
              @input="aoDigitarLargura"
              @keydown.enter="salvarLargura"
            />
          </UFormField>
          <UButton
            color="primary"
            icon="i-lucide-check"
            :loading="salvandoLargura"
            :disabled="!carregadoLargura || salvandoLargura"
            @click="salvarLargura"
          >
            Salvar
          </UButton>
          <p v-if="larguraSalvo" class="pb-3 text-xs font-medium text-emerald-600 dark:text-emerald-400">
            {{ larguraSalvo }}
          </p>
        </div>
        <UAlert v-if="larguraError" color="error" icon="i-lucide-alert-triangle" :title="larguraError" class="mt-3" />
        <p class="mt-1.5 text-xs text-neutral-400">
          Largura (em caracteres) de uma linha do recibo. Térmica de 58&nbsp;mm (MTP&nbsp;II) costuma usar
          <strong>32</strong>; papel de 80&nbsp;mm ≈ 48.
        </p>
      </div>

      <p v-if="impPadrao" class="mt-3 text-xs text-neutral-400">
        Impressora padrão do sistema: <span class="font-medium text-neutral-600 dark:text-neutral-300">{{ impPadrao }}</span>
      </p>
      <p v-if="impSalvo" class="mt-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">{{ impSalvo }}</p>
    </section>

    <!-- ============ Modo debug ============ -->
    <section class="rounded-xl border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-900">
      <div class="mb-4 flex items-start gap-3">
        <div class="flex size-9 shrink-0 items-center justify-center rounded-lg bg-amber-100 dark:bg-amber-950">
          <UIcon name="i-lucide-bug" class="size-5 text-amber-700 dark:text-amber-400" />
        </div>
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Modo debug</h2>
          <p class="text-sm text-neutral-500">Ao fechar um pedido na tela de venda.</p>
        </div>
      </div>

      <UAlert v-if="debugError" color="error" icon="i-lucide-alert-triangle" :title="debugError" class="mb-4" />

      <div class="flex items-center justify-between gap-3 rounded-lg border border-neutral-200 p-3 dark:border-neutral-800">
        <div class="min-w-0">
          <p class="text-sm font-medium text-neutral-900 dark:text-neutral-100">Não imprimir (mostrar recibo na tela)</p>
          <p class="text-xs text-neutral-500">
            Com o debug <strong>ligado</strong>, o recibo <strong>não</strong> vai à impressora: aparece formatado na tela
            (o mesmo conteúdo da impressão). Desligue para imprimir de verdade na impressora escolhida acima.
          </p>
        </div>
        <USwitch v-model="debug" aria-label="Modo debug — não imprimir" :disabled="!carregadoDebug || salvandoDebug" />
      </div>
    </section>

    <!-- Modal de confirmação do zerar -->
    <UModal
      v-model:open="showZerar"
      title="Zerar banco de dados"
      :ui="{ content: 'max-w-md' }"
    >
      <template #body>
        <div class="space-y-4">
          <div class="flex gap-3 rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-900 dark:bg-red-950/40">
            <UIcon name="i-lucide-alert-triangle" class="mt-0.5 size-5 shrink-0 text-red-600 dark:text-red-400" />
            <div class="text-sm text-red-800 dark:text-red-200">
              <p class="font-medium">Esta ação é permanente e não pode ser desfeita.</p>
              <p class="mt-1">
                Todos os <strong>eventos</strong>, <strong>pedidos</strong>, <strong>produtos</strong>,
                <strong>grupos</strong> e os dados do <strong>Anota aí</strong> (contas e pendências)
                serão removidos, e a numeração dos registros voltará a começar do 1.
              </p>
            </div>
          </div>
          <UAlert v-if="zerarError" color="error" icon="i-lucide-alert-circle" :title="zerarError" />

          <UFormField label="Digite “confirmo” para habilitar o botão">
            <UInput v-model="confirmo" placeholder="confirmo" size="lg" autofocus autocomplete="off" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="outline" :disabled="zerando" @click="showZerar = false">
            Cancelar
          </UButton>
          <UButton
            color="error"
            icon="i-lucide-trash-2"
            :disabled="!podeZerar || zerando"
            :loading="zerando"
            @click="zerar"
          >
            Zerar banco
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
