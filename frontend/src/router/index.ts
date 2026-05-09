import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '系统概览' }
  },
  {
    path: '/ticket-machine',
    name: 'TicketMachine',
    component: () => import('@/views/TicketMachine.vue'),
    meta: { title: '取号机' }
  },
  {
    path: '/window-management',
    name: 'WindowManagement',
    component: () => import('@/views/WindowManagement.vue'),
    meta: { title: '窗口管理' }
  },
  {
    path: '/call-control',
    name: 'CallControl',
    component: () => import('@/views/CallControl.vue'),
    meta: { title: '叫号控制' }
  },
  {
    path: '/statistics',
    name: 'Statistics',
    component: () => import('@/views/Statistics.vue'),
    meta: { title: '业务统计' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title || '银行排队叫号系统'} - 银行网点排队叫号系统`
  next()
})

export default router
