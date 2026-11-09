<script setup>
import { computed } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const roleBadge = computed(() => {
  switch (auth.user?.role) {
    case 'admin':
      return { label: 'Admin', classes: 'bg-purple-50 text-purple-700 border-purple-200' }
    case 'agent':
      return { label: 'Agent', classes: 'bg-blue-50 text-blue-700 border-blue-200' }
    case 'requester':
      return { label: 'Requester', classes: 'bg-zinc-100 text-zinc-700 border-zinc-200' }
    default:
      return { label: '-', classes: 'bg-zinc-100 text-zinc-600 border-zinc-200' }
  }
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <header class="bg-white border-b border-zinc-200 sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-14">
        <div class="flex items-center space-x-6">
          <RouterLink to="/tickets" class="flex items-center space-x-2 text-zinc-900 font-semibold text-sm tracking-tight">
            <span class="inline-block w-2.5 h-2.5 rounded-full bg-zinc-900"></span>
            <span>Simple Journey Helpdesk</span>
          </RouterLink>

          <nav class="hidden md:flex items-center space-x-1">
            <RouterLink
              to="/tickets"
              class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
              :class="route.path === '/tickets' ? 'bg-zinc-100 text-zinc-900' : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'"
            >
              Tiket
            </RouterLink>
            <RouterLink
              to="/tickets/new"
              class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
              :class="route.path === '/tickets/new' ? 'bg-zinc-100 text-zinc-900' : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'"
            >
              Tiket Baru
            </RouterLink>
            <RouterLink
              v-if="auth.isAdmin"
              to="/admin/users"
              class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
              :class="route.path.startsWith('/admin') ? 'bg-zinc-100 text-zinc-900' : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'"
            >
              Pengguna
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center space-x-3">
          <div class="hidden sm:flex items-center space-x-2 text-xs">
            <span class="font-medium text-zinc-900">{{ auth.user?.name }}</span>
            <span
              class="inline-flex items-center px-1.5 py-0.5 rounded text-[11px] font-medium border"
              :class="roleBadge.classes"
            >
              {{ roleBadge.label }}
            </span>
          </div>

          <button
            type="button"
            @click="handleLogout"
            class="inline-flex items-center px-2.5 py-1.5 border border-zinc-200 text-xs font-medium rounded-md text-zinc-700 bg-white hover:bg-zinc-50 focus:outline-hidden"
          >
            Keluar
          </button>
        </div>
      </div>
    </div>
  </header>
</template>
