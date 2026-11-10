import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/tickets'
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/tickets',
    name: 'tickets',
    component: () => import('@/views/TicketListView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tickets/new',
    name: 'ticket-create',
    component: () => import('@/views/TicketCreateView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/tickets/:id',
    name: 'ticket-detail',
    component: () => import('@/views/TicketDetailView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users',
    name: 'admin-users',
    component: () => import('@/views/AdminUsersView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/tickets'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore()

  if (auth.token && !auth.user) {
    try {
      await auth.fetchProfile()
    } catch {
      return next({ name: 'login' })
    }
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next({ name: 'login', query: { redirect: to.fullPath } })
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return next({ name: 'tickets' })
  }

  if (to.meta.requiresAdmin && !auth.isAdmin) {
    return next({ name: 'tickets' })
  }

  next()
})

export default router
