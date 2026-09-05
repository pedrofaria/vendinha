import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import EventosView from '../views/EventosView.vue'
import ProdutosView from '../views/ProdutosView.vue'
import PdvView from '../views/PdvView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/eventos' },
  {
    path: '/eventos',
    name: 'eventos',
    component: EventosView,
    meta: { title: 'Eventos' }
  },
  {
    path: '/eventos/:eventoId',
    name: 'produtos',
    component: ProdutosView,
    meta: { title: 'Produtos' }
  },
  {
    path: '/pdv/:eventoId',
    name: 'pdv',
    component: PdvView,
    meta: { title: 'Venda' }
  }
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · Vendinha` : 'Vendinha'
})
