<script setup>
import { watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'
import AppNavbar from '@/components/AppNavbar.vue'

const router = useRouter()
const auth = useAuthStore()
const notifications = useNotificationStore()

watch(
  () => auth.token,
  (newToken) => {
    if (newToken) {
      notifications.connect(newToken)
    } else {
      notifications.disconnect()
    }
  },
  { immediate: true }
)

function navigateFromToast(ticketId) {
  notifications.dismissToast()
  if (ticketId) {
    router.push(`/tickets/${ticketId}`)
  }
}
</script>

<template>
  <div class="min-h-screen bg-zinc-50 text-zinc-900 flex flex-col font-sans">
    <AppNavbar v-if="auth.isAuthenticated" />
    <main class="flex-1">
      <router-view />
    </main>

    <div
      v-if="notifications.activeToast"
      class="fixed bottom-4 right-4 z-50 max-w-sm w-full bg-zinc-900 text-white rounded-lg shadow-lg border border-zinc-700 p-3.5 transition-all duration-200 cursor-pointer"
      @click="navigateFromToast(notifications.activeToast.ticketId)"
    >
      <div class="flex items-start justify-between">
        <div class="space-y-1 pr-2">
          <div class="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
            {{ notifications.activeToast.type === 'ticket_created' ? 'Tiket Baru' : notifications.activeToast.type === 'ticket_status_updated' ? 'Status Diperbarui' : 'Komentar Baru' }}
          </div>
          <div class="text-xs font-medium text-zinc-100">
            {{ notifications.activeToast.title }}
          </div>
          <div class="text-xs text-zinc-400 line-clamp-2">
            {{ notifications.activeToast.message }}
          </div>
        </div>
        <button
          type="button"
          @click.stop="notifications.dismissToast"
          class="text-zinc-500 hover:text-zinc-300 text-sm p-0.5 leading-none"
        >
          &times;
        </button>
      </div>
    </div>
  </div>
</template>
