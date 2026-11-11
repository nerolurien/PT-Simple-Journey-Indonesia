<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'

const auth = useAuthStore()

const loading = ref(true)
const error = ref('')
const users = ref([])
const feedbackMessage = ref('')
const feedbackType = ref('success')
const updatingUserId = ref(null)
const searchQuery = ref('')

function formatDate(dateStr) {
  if (!dateStr) return '-'
  try {
    return new Intl.DateTimeFormat('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(new Date(dateStr))
  } catch {
    return dateStr
  }
}

function getRoleBadge(role) {
  switch (role) {
    case 'admin':
      return { label: 'Admin', classes: 'bg-purple-50 text-purple-700 border-purple-200' }
    case 'agent':
      return { label: 'Agent', classes: 'bg-blue-50 text-blue-700 border-blue-200' }
    case 'requester':
      return { label: 'Requester', classes: 'bg-zinc-100 text-zinc-700 border-zinc-200' }
    default:
      return { label: role, classes: 'bg-zinc-100 text-zinc-600 border-zinc-200' }
  }
}

async function fetchUsers() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get('/users')
    users.value = res.data || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Gagal memuat daftar pengguna'
  } finally {
    loading.value = false
  }
}

const filteredUsers = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return users.value
  return users.value.filter(
    (u) =>
      u.name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q) ||
      u.role.toLowerCase().includes(q)
  )
})

async function handleRoleChange(user, event) {
  const newRole = event.target.value
  if (newRole === user.role) return

  feedbackMessage.value = ''
  updatingUserId.value = user.id

  try {
    await api.patch(`/users/${user.id}/role`, { role: newRole })
    user.role = newRole
    feedbackType.value = 'success'
    feedbackMessage.value = `Role pengguna ${user.name} berhasil diperbarui menjadi ${newRole}`
  } catch (err) {
    event.target.value = user.role
    feedbackType.value = 'error'
    feedbackMessage.value = err.response?.data?.error || 'Gagal memperbarui role pengguna'
  } finally {
    updatingUserId.value = null
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="sm:flex sm:items-center sm:justify-between pb-5 border-b border-zinc-200">
      <div>
        <h1 class="text-lg font-semibold tracking-tight text-zinc-900">Manajemen Pengguna</h1>
        <p class="mt-1 text-xs text-zinc-500">
          Kelola hak akses dan peran seluruh pengguna dalam sistem internal helpdesk
        </p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button
          type="button"
          @click="fetchUsers"
          :disabled="loading"
          class="inline-flex items-center px-3 py-2 border border-zinc-200 rounded-md text-xs font-medium text-zinc-700 bg-white hover:bg-zinc-50 disabled:opacity-50 transition-colors"
        >
          Muat Ulang
        </button>
      </div>
    </div>

    <div
      v-if="feedbackMessage"
      class="mt-4 p-3 rounded-md text-xs flex items-center justify-between"
      :class="
        feedbackType === 'success'
          ? 'bg-emerald-50 border border-emerald-200 text-emerald-800'
          : 'bg-rose-50 border border-rose-200 text-rose-700'
      "
    >
      <span>{{ feedbackMessage }}</span>
      <button
        type="button"
        @click="feedbackMessage = ''"
        class="ml-3 font-medium hover:underline text-xs"
      >
        Tutup
      </button>
    </div>

    <div class="mt-4 p-3 bg-white border border-zinc-200 rounded-lg">
      <div class="max-w-md">
        <label for="searchUsers" class="sr-only">Cari pengguna</label>
        <input
          id="searchUsers"
          v-model="searchQuery"
          type="text"
          placeholder="Cari nama atau email pengguna..."
          class="w-full px-3 py-1.5 text-xs border border-zinc-200 rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
        />
      </div>
    </div>

    <div
      v-if="error"
      class="mt-4 p-4 rounded-lg bg-rose-50 border border-rose-200 flex items-center justify-between"
    >
      <div class="text-xs text-rose-700">{{ error }}</div>
      <button
        type="button"
        @click="fetchUsers"
        class="ml-3 px-2.5 py-1 rounded bg-rose-600 text-white text-xs font-medium hover:bg-rose-700"
      >
        Coba Lagi
      </button>
    </div>

    <div v-else-if="loading" class="mt-4 bg-white border border-zinc-200 rounded-lg p-8 text-center">
      <div class="inline-block w-5 h-5 border-2 border-zinc-300 border-t-zinc-900 rounded-full animate-spin"></div>
      <p class="mt-2 text-xs text-zinc-500">Memuat daftar pengguna...</p>
    </div>

    <div v-else class="mt-4 bg-white border border-zinc-200 rounded-lg overflow-hidden shadow-2xs">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-zinc-200 text-left text-xs">
          <thead class="bg-zinc-50 text-zinc-600 font-medium">
            <tr>
              <th scope="col" class="py-3 px-4 w-16">ID</th>
              <th scope="col" class="py-3 px-4">Nama Lengkap</th>
              <th scope="col" class="py-3 px-4">Email Kantor</th>
              <th scope="col" class="py-3 px-4 w-32">Role Saat Ini</th>
              <th scope="col" class="py-3 px-4 w-44">Ubah Role</th>
              <th scope="col" class="py-3 px-4 w-36 text-right">Terdaftar</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100 bg-white">
            <tr
              v-for="user in filteredUsers"
              :key="user.id"
              class="hover:bg-zinc-50 transition-colors"
            >
              <td class="py-3 px-4 font-mono text-zinc-500">
                #{{ user.id }}
              </td>
              <td class="py-3 px-4 font-medium text-zinc-900">
                {{ user.name }}
                <span
                  v-if="user.id === auth.user?.id"
                  class="ml-1 text-[10px] text-zinc-500 font-normal"
                >
                  (Akun Anda)
                </span>
              </td>
              <td class="py-3 px-4 text-zinc-600">
                {{ user.email }}
              </td>
              <td class="py-3 px-4">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border"
                  :class="getRoleBadge(user.role).classes"
                >
                  {{ getRoleBadge(user.role).label }}
                </span>
              </td>
              <td class="py-3 px-4">
                <div class="flex items-center space-x-2">
                  <select
                    :value="user.role"
                    :disabled="user.id === auth.user?.id || updatingUserId === user.id"
                    @change="handleRoleChange(user, $event)"
                    class="px-2 py-1 text-xs border border-zinc-200 rounded bg-white text-zinc-800 disabled:opacity-40 disabled:cursor-not-allowed focus:outline-hidden focus:ring-1 focus:ring-zinc-900"
                  >
                    <option value="requester">Requester</option>
                    <option value="agent">Agent</option>
                    <option value="admin">Admin</option>
                  </select>
                  <span
                    v-if="user.id === auth.user?.id"
                    class="text-[10px] text-zinc-400 italic whitespace-nowrap"
                  >
                    Terkunci
                  </span>
                </div>
              </td>
              <td class="py-3 px-4 text-right text-zinc-500 whitespace-nowrap">
                {{ formatDate(user.created_at) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="px-4 py-3 bg-zinc-50 border-t border-zinc-200 text-xs text-zinc-500">
        Total {{ filteredUsers.length }} pengguna terdaftar dalam sistem
      </div>
    </div>
  </div>
</template>
