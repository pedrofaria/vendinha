import { ref } from 'vue'

const STORAGE = 'vendinha-color-mode'

// Dark mode manual (o useColorMode do Nuxt UI v4 é um stub fora do Nuxt).
// A classe `dark` em <html> é o seletor que o CSS do Nuxt UI v4 usa.
function initialDark(): boolean {
  const saved = localStorage.getItem(STORAGE)
  if (saved === 'dark') return true
  if (saved === 'light') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

const isDark = ref(initialDark())

function apply() {
  const root = document.documentElement
  root.classList.toggle('dark', isDark.value)
  root.style.colorScheme = isDark.value ? 'dark' : 'light'
  localStorage.setItem(STORAGE, isDark.value ? 'dark' : 'light')
}

export function useColorMode() {
  return {
    isDark,
    toggle() {
      isDark.value = !isDark.value
      apply()
    },
    init() {
      apply()
    }
  }
}
