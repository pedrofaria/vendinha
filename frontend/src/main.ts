import './assets/css/main.css' // deve vir primeiro
import { createApp } from 'vue'
import ui from '@nuxt/ui/vue-plugin'
import App from './App.vue'
import { router } from './router'
import { useColorMode } from './lib/colorMode'

const { init } = useColorMode()
init()

createApp(App).use(router).use(ui).mount('#app')
