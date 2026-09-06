import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import EventosView from '../views/EventosView.vue'
import EventoShell from '../views/EventoShell.vue'
import DashboardView from '../views/DashboardView.vue'
import ProdutosView from '../views/ProdutosView.vue'
import ContasView from '../views/ContasView.vue'
import PedidosView from '../views/PedidosView.vue'
import PdvView from '../views/PdvView.vue'
import ConfigView from '../views/ConfigView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/eventos' },
  {
    path: '/config',
    name: 'config',
    component: ConfigView,
    meta: { title: 'Configurações' }
  },
  {
    path: '/eventos',
    name: 'eventos',
    component: EventosView,
    meta: { title: 'Eventos' }
  },
  {
    // Página do evento: hub com abas (Dashboard padrão / Produtos / Anota aí).
    path: '/eventos/:eventoId',
    name: 'evento',
    component: EventoShell,
    children: [
      {
        path: '',
        name: 'evento-dashboard',
        component: DashboardView,
        meta: { title: 'Dashboard' }
      },
      {
        path: 'produtos',
        name: 'evento-produtos',
        component: ProdutosView,
        meta: { title: 'Produtos' }
      },
      {
        path: 'contas',
        name: 'evento-contas',
        component: ContasView,
        meta: { title: 'Anota aí' }
      },
      {
        path: 'pedidos',
        name: 'evento-pedidos',
        component: PedidosView,
        meta: { title: 'Pedidos' }
      }
    ]
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
