<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'
import StatusBadge from '@/components/StatusBadge.vue'
import PriorityBadge from '@/components/PriorityBadge.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const ticketId = route.params.id

const loading = ref(true)
const error = ref('')
const ticket = ref(null)

const actionLoading = ref(false)
const actionError = ref('')

const commentText = ref('')
const isInternalNote = ref(false)
const commentLoading = ref(false)
const commentError = ref('')

const uploading = ref(false)
const uploadError = ref('')
const uploadSuccess = ref('')
const downloadingId = ref(null)

const assignableAgents = ref([])

const isOwner = computed(() => {
  return ticket.value?.requester_id === auth.user?.id
})

const isAdminOrAgent = computed(() => {
  return auth.isAdmin || auth.isAgent
})

function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

async function handleUploadAttachment(event) {
  uploadError.value = ''
  uploadSuccess.value = ''
  const file = event.target.files?.[0]
  if (!file) return

  const allowedExts = ['.png', '.jpg', '.jpeg', '.pdf', '.txt', '.log']
  const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!allowedExts.includes(ext)) {
    uploadError.value = 'Tipe file tidak didukung. Format yang diizinkan: PNG, JPG, PDF, TXT, LOG.'
    event.target.value = ''
    return
  }

  if (file.size > 5 * 1024 * 1024) {
    uploadError.value = 'Ukuran file melebihi batas maksimal 5 MB.'
    event.target.value = ''
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const res = await api.post(`/tickets/${ticket.value.id}/attachments`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if (!ticket.value.attachments) {
      ticket.value.attachments = []
    }
    ticket.value.attachments.push(res.data)
    uploadSuccess.value = 'Berkas berhasil diunggah'
    event.target.value = ''
  } catch (err) {
    uploadError.value = err.response?.data?.error || 'Gagal mengunggah lampiran'
  } finally {
    uploading.value = false
  }
}

async function downloadAttachment(att) {
  downloadingId.value = att.id
  try {
    const res = await api.get(`/tickets/${ticket.value.id}/attachments/${att.id}`, {
      responseType: 'blob'
    })
    const blobUrl = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = blobUrl
    link.setAttribute('download', att.original_name)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(blobUrl)
  } catch {
    uploadError.value = 'Gagal mengunduh lampiran'
  } finally {
    downloadingId.value = null
  }
}

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

async function fetchTicket() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get(`/tickets/${ticketId}`)
    ticket.value = res.data
  } catch (err) {
    if (err.response?.status === 403) {
      error.value = 'Anda tidak memiliki hak akses untuk melihat tiket ini.'
    } else if (err.response?.status === 404) {
      error.value = 'Tiket tidak ditemukan.'
    } else {
      error.value = 'Gagal memuat data tiket. Silakan periksa koneksi server.'
    }
  } finally {
    loading.value = false
  }
}

async function loadAgents() {
  if (auth.isAdmin) {
    try {
      const res = await api.get('/users')
      assignableAgents.value = (res.data || []).filter(
        (u) => u.role === 'agent' || u.role === 'admin'
      )
    } catch {
      // Ignored if failed
    }
  }
}

const availableTransitions = computed(() => {
  if (!ticket.value) return []
  const status = ticket.value.status
  const role = auth.user?.role
  const owner = isOwner.value
  const buttons = []

  if (status === 'open') {
    if (role === 'agent' || role === 'admin') {
      buttons.push({
        status: 'in_progress',
        label: 'Proses Tiket',
        classes: 'bg-zinc-900 text-white hover:bg-zinc-800'
      })
    }
    if (role === 'admin') {
      buttons.push({
        status: 'closed',
        label: 'Tutup Tiket',
        classes: 'bg-white text-zinc-700 border border-zinc-200 hover:bg-zinc-50'
      })
    } else if (role === 'requester' && owner) {
      buttons.push({
        status: 'closed',
        label: 'Batalkan Tiket',
        classes: 'bg-white text-rose-700 border border-rose-200 hover:bg-rose-50'
      })
    }
  } else if (status === 'in_progress') {
    if (role === 'agent' || role === 'admin') {
      buttons.push({
        status: 'resolved',
        label: 'Tandai Selesai (Resolved)',
        classes: 'bg-emerald-600 text-white hover:bg-emerald-700'
      })
    }
  } else if (status === 'resolved') {
    if (role === 'agent' || role === 'admin' || (role === 'requester' && owner)) {
      buttons.push({
        status: 'closed',
        label: 'Tutup Tiket',
        classes: 'bg-zinc-900 text-white hover:bg-zinc-800'
      })
    }
    if (role === 'requester' && owner) {
      buttons.push({
        status: 'open',
        label: 'Buka Kembali (Reopen)',
        classes: 'bg-white text-amber-700 border border-amber-200 hover:bg-amber-50'
      })
    }
  }

  return buttons
})

async function handleStatusChange(targetStatus) {
  actionLoading.value = true
  actionError.value = ''
  try {
    const res = await api.patch(`/tickets/${ticket.value.id}`, { status: targetStatus })
    ticket.value = res.data
  } catch (err) {
    actionError.value = err.response?.data?.error || 'Gagal mengubah status tiket'
  } finally {
    actionLoading.value = false
  }
}

async function handlePriorityChange(event) {
  const newPriority = event.target.value
  if (!newPriority || newPriority === ticket.value.priority) return

  actionLoading.value = true
  actionError.value = ''
  try {
    const res = await api.patch(`/tickets/${ticket.value.id}`, { priority: newPriority })
    ticket.value = res.data
  } catch (err) {
    actionError.value = err.response?.data?.error || 'Gagal mengubah prioritas'
  } finally {
    actionLoading.value = false
  }
}

async function handleAssigneeChange(event) {
  const newAssigneeId = parseInt(event.target.value, 10)
  actionLoading.value = true
  actionError.value = ''
  try {
    const res = await api.patch(`/tickets/${ticket.value.id}`, { assignee_id: newAssigneeId })
    ticket.value = res.data
  } catch (err) {
    actionError.value = err.response?.data?.error || 'Gagal mengubah teknisi penanggung jawab'
  } finally {
    actionLoading.value = false
  }
}

async function handleAssignToMe() {
  actionLoading.value = true
  actionError.value = ''
  try {
    const res = await api.patch(`/tickets/${ticket.value.id}`, { assignee_id: auth.user.id })
    ticket.value = res.data
  } catch (err) {
    actionError.value = err.response?.data?.error || 'Gagal menugaskan tiket ke diri sendiri'
  } finally {
    actionLoading.value = false
  }
}

async function handleUnassign() {
  actionLoading.value = true
  actionError.value = ''
  try {
    const res = await api.patch(`/tickets/${ticket.value.id}`, { assignee_id: 0 })
    ticket.value = res.data
  } catch (err) {
    actionError.value = err.response?.data?.error || 'Gagal melepas penugasan'
  } finally {
    actionLoading.value = false
  }
}

async function handleAddComment() {
  commentError.value = ''
  if (!commentText.value.trim()) {
    commentError.value = 'Isi komentar tidak boleh kosong'
    return
  }

  commentLoading.value = true
  try {
    const payload = {
      body: commentText.value.trim(),
      is_internal: isAdminOrAgent.value ? isInternalNote.value : false
    }

    const res = await api.post(`/tickets/${ticket.value.id}/comments`, payload)
    if (!ticket.value.comments) {
      ticket.value.comments = []
    }
    ticket.value.comments.push(res.data)
    commentText.value = ''
    isInternalNote.value = false
  } catch (err) {
    commentError.value = err.response?.data?.error || 'Gagal menambahkan komentar'
  } finally {
    commentLoading.value = false
  }
}

onMounted(() => {
  fetchTicket()
  loadAgents()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="mb-4">
      <RouterLink
        to="/tickets"
        class="inline-flex items-center text-xs text-zinc-500 hover:text-zinc-900 transition-colors"
      >
        <span class="mr-1">&larr;</span> Kembali ke Daftar Tiket
      </RouterLink>
    </div>

    <!-- Error State -->
    <div
      v-if="error"
      class="p-6 bg-white border border-rose-200 rounded-lg text-center"
    >
      <div class="text-xs text-rose-700 font-medium">{{ error }}</div>
      <div class="mt-4">
        <RouterLink
          to="/tickets"
          class="inline-flex items-center px-3 py-1.5 border border-zinc-200 rounded-md text-xs font-medium text-zinc-700 bg-white hover:bg-zinc-50"
        >
          Lihat Tiket Lainnya
        </RouterLink>
      </div>
    </div>

    <!-- Loading State -->
    <div v-else-if="loading" class="p-12 bg-white border border-zinc-200 rounded-lg text-center">
      <div class="inline-block w-5 h-5 border-2 border-zinc-300 border-t-zinc-900 rounded-full animate-spin"></div>
      <p class="mt-2 text-xs text-zinc-500">Memuat detail tiket...</p>
    </div>

    <!-- Main Detail View -->
    <div v-else-if="ticket" class="space-y-6">
      <!-- Header Banner -->
      <div class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs">
        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <div class="flex items-center space-x-2 flex-wrap gap-y-1">
              <span class="font-mono text-xs text-zinc-500">#{{ ticket.id }}</span>
              <StatusBadge :status="ticket.status" />
              <PriorityBadge :priority="ticket.priority" />
              <span
                v-if="ticket.is_sla_breached"
                class="px-2 py-0.5 rounded text-[10px] font-medium bg-red-50 text-red-700 border border-red-200"
              >
                Terlambat (SLA)
              </span>
              <span class="text-xs text-zinc-400">|</span>
              <span class="text-xs text-zinc-600">{{ ticket.category }}</span>
            </div>
            <h1 class="mt-2 text-base sm:text-lg font-semibold text-zinc-900">
              {{ ticket.title }}
            </h1>
            <p class="mt-1 text-xs text-zinc-500">
              Diajukan oleh <span class="font-medium text-zinc-800">{{ ticket.requester?.name }}</span> pada {{ formatDate(ticket.created_at) }}
            </p>
          </div>

          <!-- Dynamic Status Transition Buttons -->
          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="btn in availableTransitions"
              :key="btn.status"
              type="button"
              :disabled="actionLoading"
              @click="handleStatusChange(btn.status)"
              class="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-50"
              :class="btn.classes"
            >
              {{ btn.label }}
            </button>
          </div>
        </div>

        <div v-if="actionError" class="mt-3 p-2 rounded bg-rose-50 border border-rose-200 text-xs text-rose-700">
          {{ actionError }}
        </div>
      </div>

      <!-- Two-column Body -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left: Ticket Description & Comments -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Description Card -->
          <div class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs">
            <h2 class="text-xs font-semibold text-zinc-800 uppercase tracking-wider mb-3">
              Deskripsi Kendala
            </h2>
            <div class="text-xs text-zinc-700 whitespace-pre-line leading-relaxed">
              {{ ticket.description }}
            </div>
          </div>

          <!-- Attachments Section -->
          <div class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xs font-semibold text-zinc-800 uppercase tracking-wider">
                Lampiran Berkas ({{ ticket.attachments?.length || 0 }})
              </h2>
              <label
                v-if="isAdminOrAgent || isOwner"
                class="inline-flex items-center px-2.5 py-1 border border-zinc-200 rounded text-xs font-medium text-zinc-700 bg-white hover:bg-zinc-50 cursor-pointer transition-colors"
                :class="uploading ? 'opacity-50 cursor-not-allowed' : ''"
              >
                <span>{{ uploading ? 'Mengunggah...' : '+ Unggah Berkas' }}</span>
                <input
                  type="file"
                  :disabled="uploading"
                  accept=".png,.jpg,.jpeg,.pdf,.txt,.log"
                  @change="handleUploadAttachment"
                  class="sr-only"
                />
              </label>
            </div>

            <div v-if="uploadError" class="mb-3 p-2 rounded bg-rose-50 border border-rose-200 text-xs text-rose-700">
              {{ uploadError }}
            </div>
            <div v-if="uploadSuccess" class="mb-3 p-2 rounded bg-emerald-50 border border-emerald-200 text-xs text-emerald-700">
              {{ uploadSuccess }}
            </div>

            <div v-if="!ticket.attachments || ticket.attachments.length === 0" class="text-center py-5 text-xs text-zinc-400">
              Belum ada berkas terlampir pada tiket ini.
            </div>

            <div v-else class="divide-y divide-zinc-100">
              <div
                v-for="att in ticket.attachments"
                :key="att.id"
                class="py-2.5 flex items-center justify-between text-xs"
              >
                <div class="flex items-center space-x-2 truncate pr-2">
                  <span class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-zinc-100 text-zinc-600">
                    {{ att.mime_type.split('/')[1] || 'FILE' }}
                  </span>
                  <span class="font-medium text-zinc-800 truncate" :title="att.original_name">
                    {{ att.original_name }}
                  </span>
                  <span class="text-zinc-400 text-[11px] whitespace-nowrap">
                    ({{ formatFileSize(att.size_bytes) }})
                  </span>
                </div>
                <div class="flex items-center space-x-3 shrink-0">
                  <span class="text-[11px] text-zinc-400 hidden sm:inline">
                    {{ formatDate(att.created_at) }}
                  </span>
                  <button
                    type="button"
                    :disabled="downloadingId === att.id"
                    @click="downloadAttachment(att)"
                    class="px-2 py-1 border border-zinc-200 rounded text-xs text-zinc-700 hover:bg-zinc-50 disabled:opacity-50"
                  >
                    {{ downloadingId === att.id ? 'Mengunduh...' : 'Unduh' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Comments Section -->
          <div class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs">
            <h2 class="text-xs font-semibold text-zinc-800 uppercase tracking-wider mb-4">
              Aktivitas & Komentar ({{ ticket.comments?.length || 0 }})
            </h2>

            <!-- Comment List -->
            <div v-if="ticket.comments?.length === 0" class="text-center py-6 text-xs text-zinc-500">
              Belum ada komentar pada tiket ini.
            </div>

            <div v-else class="space-y-3 mb-6">
              <div
                v-for="comment in ticket.comments"
                :key="comment.id"
                class="p-3.5 rounded-md border text-xs"
                :class="comment.is_internal ? 'bg-amber-50/70 border-amber-200' : 'bg-white border-zinc-200'"
              >
                <div class="flex items-center justify-between pb-2 mb-2 border-b" :class="comment.is_internal ? 'border-amber-200/60' : 'border-zinc-100'">
                  <div class="flex items-center space-x-2">
                    <span class="font-medium text-zinc-900">{{ comment.author?.name || 'Pengguna' }}</span>
                    <span class="text-[11px] text-zinc-500 capitalize">({{ comment.author?.role }})</span>
                    <span
                      v-if="comment.is_internal"
                      class="px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-100 text-amber-800 border border-amber-200"
                    >
                      Catatan Internal
                    </span>
                  </div>
                  <span class="text-[11px] text-zinc-400">
                    {{ formatDate(comment.created_at) }}
                  </span>
                </div>
                <div class="text-zinc-800 whitespace-pre-line leading-relaxed">
                  {{ comment.body }}
                </div>
              </div>
            </div>

            <!-- Comment Form -->
            <div class="pt-4 border-t border-zinc-200">
              <h3 class="text-xs font-medium text-zinc-900 mb-2">Tulis Komentar</h3>

              <div v-if="commentError" class="mb-3 p-2 rounded bg-rose-50 border border-rose-200 text-xs text-rose-700">
                {{ commentError }}
              </div>

              <form @submit.prevent="handleAddComment" class="space-y-3">
                <div>
                  <textarea
                    v-model="commentText"
                    rows="3"
                    placeholder="Tulis respons, instruksi pengecekan, atau update penyelesaian tiket..."
                    class="block w-full px-3 py-2 text-xs border border-zinc-200 rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                  ></textarea>
                </div>

                <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div v-if="isAdminOrAgent" class="flex items-center space-x-2">
                    <input
                      id="internalNote"
                      v-model="isInternalNote"
                      type="checkbox"
                      class="rounded border-zinc-300 text-zinc-900 focus:ring-zinc-900"
                    />
                    <label for="internalNote" class="text-xs text-zinc-700">
                      Jadikan catatan internal (hanya dilihat Agent & Admin)
                    </label>
                  </div>
                  <div v-else></div>

                  <button
                    type="submit"
                    :disabled="commentLoading"
                    class="inline-flex justify-center px-3 py-1.5 border border-transparent rounded-md text-xs font-medium text-white bg-zinc-900 hover:bg-zinc-800 disabled:opacity-50 transition-colors"
                  >
                    {{ commentLoading ? 'Mengirim...' : 'Kirim Komentar' }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>

        <!-- Right: Meta & Agent Controls -->
        <div class="space-y-6">
          <!-- Meta Card -->
          <div class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs space-y-3.5">
            <h2 class="text-xs font-semibold text-zinc-800 uppercase tracking-wider pb-2 border-b border-zinc-100">
              Informasi Tiket
            </h2>

            <div>
              <div class="text-[11px] text-zinc-500">Kategori</div>
              <div class="text-xs font-medium text-zinc-900 mt-0.5">{{ ticket.category }}</div>
            </div>

            <div>
              <div class="text-[11px] text-zinc-500">Pemohon (Requester)</div>
              <div class="text-xs font-medium text-zinc-900 mt-0.5">{{ ticket.requester?.name }}</div>
              <div class="text-[11px] text-zinc-500">{{ ticket.requester?.email }}</div>
            </div>

            <div>
              <div class="text-[11px] text-zinc-500">Teknisi (Assignee)</div>
              <div class="text-xs font-medium text-zinc-900 mt-0.5">
                <span v-if="ticket.assignee">{{ ticket.assignee.name }}</span>
                <span v-else class="text-zinc-400 italic">Belum ditugaskan</span>
              </div>
            </div>

            <div>
              <div class="text-[11px] text-zinc-500">Waktu Dibuat</div>
              <div class="text-xs text-zinc-700 mt-0.5">{{ formatDate(ticket.created_at) }}</div>
            </div>

            <div>
              <div class="text-[11px] text-zinc-500">Terakhir Diperbarui</div>
              <div class="text-xs text-zinc-700 mt-0.5">{{ formatDate(ticket.updated_at) }}</div>
            </div>
          </div>

          <!-- Agent / Admin Quick Controls -->
          <div
            v-if="isAdminOrAgent"
            class="bg-white border border-zinc-200 rounded-lg p-5 shadow-2xs space-y-4"
          >
            <h2 class="text-xs font-semibold text-zinc-800 uppercase tracking-wider pb-2 border-b border-zinc-100">
              Kendali Teknisi / Admin
            </h2>

            <div>
              <label for="updatePriority" class="block text-xs font-medium text-zinc-700 mb-1">
                Ubah Prioritas
              </label>
              <select
                id="updatePriority"
                :value="ticket.priority"
                :disabled="actionLoading"
                @change="handlePriorityChange"
                class="w-full px-2.5 py-1.5 text-xs border border-zinc-200 rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="urgent">Urgent</option>
              </select>
            </div>

            <div>
              <label for="updateAssignee" class="block text-xs font-medium text-zinc-700 mb-1">
                Tugaskan Teknisi
              </label>

              <!-- Admin Select Dropdown -->
              <div v-if="auth.isAdmin && assignableAgents.length > 0" class="space-y-2">
                <select
                  id="updateAssignee"
                  :value="ticket.assignee_id || 0"
                  :disabled="actionLoading"
                  @change="handleAssigneeChange"
                  class="w-full px-2.5 py-1.5 text-xs border border-zinc-200 rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900"
                >
                  <option :value="0">Belum ditugaskan</option>
                  <option v-for="agent in assignableAgents" :key="agent.id" :value="agent.id">
                    {{ agent.name }} ({{ agent.role }})
                  </option>
                </select>
              </div>

              <!-- Quick Self-Assign & Unassign Buttons for Agents/Admins -->
              <div class="mt-2 flex flex-col space-y-1.5">
                <button
                  v-if="ticket.assignee_id !== auth.user.id"
                  type="button"
                  :disabled="actionLoading"
                  @click="handleAssignToMe"
                  class="w-full px-2.5 py-1.5 border border-zinc-200 rounded text-xs font-medium text-zinc-700 bg-white hover:bg-zinc-50 disabled:opacity-50 text-center"
                >
                  Tugaskan ke Saya
                </button>
                <button
                  v-if="ticket.assignee_id"
                  type="button"
                  :disabled="actionLoading"
                  @click="handleUnassign"
                  class="w-full px-2.5 py-1.5 border border-rose-200 rounded text-xs font-medium text-rose-700 bg-white hover:bg-rose-50 disabled:opacity-50 text-center"
                >
                  Lepas Penugasan
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
