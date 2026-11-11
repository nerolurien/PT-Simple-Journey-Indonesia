<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'

const auth = useAuthStore()
const notifications = useNotificationStore()
const router = useRouter()
const route = useRoute()

const showNotifications = ref(false)
const notificationRef = ref(null)

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

function formatTime(dateStr) {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}

function handleNotificationClick(n) {
  notifications.markAsRead(n.id)
  showNotifications.value = false
  if (n.ticketId) {
    router.push(`/tickets/${n.ticketId}`)
  }
}

function handleClickOutside(event) {
  if (notificationRef.value && !notificationRef.value.contains(event.target)) {
    showNotifications.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

function handleLogout() {
  notifications.disconnect()
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
              :class="route.path === '/admin/users' ? 'bg-zinc-100 text-zinc-900' : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'"
            >
              Pengguna
            </RouterLink>
            <RouterLink
              v-if="auth.isAdmin"
              to="/admin/stats"
              class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
              :class="route.path === '/admin/stats' ? 'bg-zinc-100 text-zinc-900' : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-50'"
            >
              Statistik
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center space-x-3">
          <!-- Notification Bell & Dropdown -->
          <div class="relative" ref="notificationRef">
            <button
              type="button"
              @click="showNotifications = !showNotifications"
              class="relative p-1.5 rounded-md text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 focus:outline-hidden"
              title="Notifikasi Real-time"
            >
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
              <span
                v-if="notifications.unreadCount > 0"
                class="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 rounded-full bg-red-600 text-white text-[10px] font-bold flex items-center justify-center"
              >
                {{ notifications.unreadCount > 9 ? '9+' : notifications.unreadCount }}
              </span>
            </button>

            <!-- Dropdown Menu -->
            <div
              v-if="showNotifications"
              class="absolute right-0 mt-2 w-80 sm:w-96 bg-white border border-zinc-200 rounded-lg shadow-lg z-50 overflow-hidden"
            >
              <div class="p-3 bg-zinc-50 border-b border-zinc-200 flex items-center justify-between">
                <span class="text-xs font-semibold text-zinc-800">Notifikasi Real-time</span>
                <button
                  v-if="notifications.unreadCount > 0"
                  type="button"
                  @click="notifications.markAllAsRead"
                  class="text-[11px] text-zinc-600 hover:text-zinc-900 font-medium"
                >
                  Tandai dibaca
                </button>
              </div>

              <div class="max-h-80 overflow-y-auto divide-y divide-zinc-100">
                <div
                  v-if="notifications.items.length === 0"
                  class="p-6 text-center text-xs text-zinc-400"
                >
                  Belum ada notifikasi baru
                </div>
                <div
                  v-for="n in notifications.items"
                  :key="n.id"
                  @click="handleNotificationClick(n)"
                  class="p-3 hover:bg-zinc-50 cursor-pointer transition-colors flex items-start space-x-2.5"
                  :class="!n.read ? 'bg-zinc-50/60' : ''"
                >
                  <span
                    class="w-1.5 h-1.5 rounded-full mt-1.5 shrink-0"
                    :class="!n.read ? 'bg-blue-600' : 'bg-transparent'"
                  ></span>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between">
                      <span class="text-[11px] font-medium text-zinc-900 truncate">{{ n.title }}</span>
                      <span class="text-[10px] text-zinc-400 shrink-0 ml-2">{{ formatTime(n.createdAt) }}</span>
                    </div>
                    <p class="text-xs text-zinc-600 mt-0.5 line-clamp-2 leading-snug">{{ n.message }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

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
