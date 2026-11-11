import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useNotificationStore = defineStore('notifications', () => {
  const items = ref([])
  const activeToast = ref(null)
  let eventSource = null
  let toastTimer = null

  const unreadCount = computed(() => {
    return items.value.filter((n) => !n.read).length
  })

  function connect(token) {
    if (!token || eventSource) return

    const baseURL = import.meta.env.VITE_API_URL || '/api'
    const streamURL = `${baseURL}/notifications/stream?token=${encodeURIComponent(token)}`

    eventSource = new EventSource(streamURL)

    const handleEvent = (type, event) => {
      try {
        const payload = JSON.parse(event.data)
        const notification = {
          id: Date.now() + Math.random(),
          type,
          ticketId: payload.ticket_id,
          title: payload.title,
          message: payload.message,
          createdAt: payload.created_at || new Date().toISOString(),
          read: false
        }

        items.value.unshift(notification)
        if (items.value.length > 30) {
          items.value.pop()
        }

        showToast(notification)
      } catch (err) {
        console.error('Failed to parse SSE event payload:', err)
      }
    }

    eventSource.addEventListener('ticket_created', (e) => handleEvent('ticket_created', e))
    eventSource.addEventListener('ticket_status_updated', (e) => handleEvent('ticket_status_updated', e))
    eventSource.addEventListener('ticket_comment_added', (e) => handleEvent('ticket_comment_added', e))

    eventSource.onerror = () => {}
  }

  function disconnect() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    if (toastTimer) {
      clearTimeout(toastTimer)
      toastTimer = null
    }
    activeToast.value = null
  }

  function showToast(notification) {
    if (toastTimer) {
      clearTimeout(toastTimer)
    }
    activeToast.value = notification
    toastTimer = setTimeout(() => {
      activeToast.value = null
    }, 4500)
  }

  function dismissToast() {
    if (toastTimer) {
      clearTimeout(toastTimer)
      toastTimer = null
    }
    activeToast.value = null
  }

  function markAsRead(id) {
    const item = items.value.find((n) => n.id === id)
    if (item) {
      item.read = true
    }
  }

  function markAllAsRead() {
    items.value.forEach((n) => {
      n.read = true
    })
  }

  function clearAll() {
    items.value = []
  }

  return {
    items,
    activeToast,
    unreadCount,
    connect,
    disconnect,
    dismissToast,
    markAsRead,
    markAllAsRead,
    clearAll
  }
})
