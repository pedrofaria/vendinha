<script lang="ts" setup>
import { useColorMode } from './lib/colorMode'
import { RouterLink } from 'vue-router'

const { isDark, toggle } = useColorMode()
</script>

<template>
  <div class="flex h-full flex-col">
    <header class="flex items-center justify-between border-b border-neutral-200 px-4 py-2.5 dark:border-neutral-800">
      <div class="flex items-center gap-4">
        <RouterLink to="/eventos" class="text-lg font-semibold tracking-tight text-emerald-700 dark:text-emerald-400">
          Vendinha
        </RouterLink>
        <nav class="flex items-center gap-1 text-sm">
          <RouterLink
            to="/eventos"
            class="rounded-md px-2.5 py-1.5 text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800"
            active-class="bg-emerald-50 font-medium text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
          >
            Eventos
          </RouterLink>
          <RouterLink
            to="/config"
            class="rounded-md px-2.5 py-1.5 text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-neutral-800"
            active-class="bg-emerald-50 font-medium text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300"
          >
            Configurações
          </RouterLink>
        </nav>
      </div>
      <button
        class="rounded-lg border border-neutral-200 p-2 text-neutral-600 hover:bg-neutral-100 dark:border-neutral-800 dark:text-neutral-300 dark:hover:bg-neutral-800"
        type="button"
        :aria-label="isDark ? 'Modo claro' : 'Modo escuro'"
        @click="toggle"
      >
        <UIcon :name="isDark ? 'i-lucide-sun' : 'i-lucide-moon'" class="size-5" />
      </button>
    </header>

    <main class="min-h-0 flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
