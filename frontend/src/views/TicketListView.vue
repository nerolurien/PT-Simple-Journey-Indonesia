<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/lib/api'
import StatusBadge from '@/components/StatusBadge.vue'
import PriorityBadge from '@/components/PriorityBadge.vue'
import PaginationBar from '@/components/PaginationBar.vue'

const router = useRouter()
const auth = useAuthStore()

const loading = ref(false)
const error = ref('')
const tickets = ref([])
const pagination = reactive({
  page: 1,
  limit: 10,
  total_count: 0,
  total_pages: 1
})

const filters = reactive({
  q: '',
  status: '',
  priority: '',
  sort: 'created_at:desc'
})

let searchDebounceTimer = null

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

async function fetchTickets() {
  loading.value = true
  error.value = ''

  try {
    const [sortBy, order] = filters.sort.split(':')
    const params = {
      page: pagination.page,
      limit: pagination.limit,
      sort_by: sortBy,
      order: order
    }

    if (filters.q.trim()) {
      params.q = filters.q.trim()
    }
    if (filters.status) {
      params.status = filters.status
    }
    if (filters.priority) {
      params.priority = filters.priority
    }

    const res = await api.get('/tickets', { params })
    tickets.value = res.data.tickets || []
    pagination.page = res.data.page || 1
    pagination.limit = res.data.limit || 10
    pagination.total_count = res.data.total_count || 0
    pagination.total_pages = res.data.total_pages || 1
  } catch (err) {
    if (err.response?.data?.error) {
      error.value = err.response.data.error
    } else {
      error.value = 'Gagal memuat daftar tiket. Pastikan server backend sedang aktif.'
    }
  } finally {
    loading.value = false
  }
}

function handlePageChange(newPage) {
  pagination.page = newPage
  fetchTickets()
}

function handleSearchInput() {
  clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    pagination.page = 1
    fetchTickets()
  }, 350)
}

function handleFilterChange() {
  pagination.page = 1
  fetchTickets()
}

function resetFilters() {
  filters.q = ''
  filters.status = ''
  filters.priority = ''
  filters.sort = 'created_at:desc'
  pagination.page = 1
  fetchTickets()
}

function navigateToDetail(ticketId) {
  router.push(`/tickets/${ticketId}`)
}

onMounted(() => {
  fetchTickets()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="sm:flex sm:items-center sm:justify-between pb-5 border-b border-zinc-200">
      <div>
        <h1 class="text-lg font-semibold tracking-tight text-zinc-900">Tiket Bantuan</h1>
        <p class="mt-1 text-xs text-zinc-500">
          {{ auth.isRequester ? 'Kelola dan pantau tiket bantuan yang Anda ajukan' : 'Kelola antrean tiket bantuan teknis operasional' }}
        </p>
      </div>
      <div class="mt-4 sm:mt-0">
        <RouterLink
          to="/tickets/new"
          class="inline-flex items-center px-3 py-2 border border-transparent rounded-md text-xs font-medium text-white bg-zinc-900 hover:bg-zinc-800 transition-colors"
        >
          Buat Tiket Baru
        </RouterLink>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="mt-4 p-3 bg-white border border-zinc-200 rounded-lg">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
        <div class="lg:col-span-2">
          <label for="search" class="sr-only">Cari judul tiket</label>
          <input
            id="search"
            v-model="filters.q"
            @input="handleSearchInput"
            type="text"
            placeholder="Cari berdasarkan judul tiket..."
            class="w-full px-3 py-1.5 text-xs border border-zinc-200 rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
          />
        </div>

        <div>
          <label for="status" class="sr-only">Status</label>
          <select
            id="status"
            v-model="filters.status"
            @change="handleFilterChange"
            class="w-full px-2.5 py-1.5 text-xs border border-zinc-200 rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
          >
            <option value="">Semua Status</option>
            <option value="open">Open</option>
            <option value="in_progress">In Progress</option>
            <option value="resolved">Resolved</option>
            <option value="closed">Closed</option>
          </select>
        </div>

        <div>
          <label for="priority" class="sr-only">Prioritas</label>
          <select
            id="priority"
            v-model="filters.priority"
            @change="handleFilterChange"
            class="w-full px-2.5 py-1.5 text-xs border border-zinc-200 rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
          >
            <option value="">Semua Prioritas</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>
        </div>

        <div class="flex items-center space-x-2">
          <select
            id="sort"
            v-model="filters.sort"
            @change="handleFilterChange"
            class="w-full px-2.5 py-1.5 text-xs border border-zinc-200 rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
          >
            <option value="created_at:desc">Terbaru</option>
            <option value="created_at:asc">Terlama</option>
            <option value="updated_at:desc">Terakhir Diperbarui</option>
            <option value="priority:desc">Prioritas</option>
          </select>
          <button
            type="button"
            @click="resetFilters"
            title="Reset filter"
            class="px-2.5 py-1.5 border border-zinc-200 rounded-md text-xs text-zinc-600 hover:bg-zinc-50 whitespace-nowrap"
          >
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div
      v-if="error"
      class="mt-4 p-4 rounded-lg bg-rose-50 border border-rose-200 flex items-center justify-between"
    >
      <div class="text-xs text-rose-700">{{ error }}</div>
      <button
        type="button"
        @click="fetchTickets"
        class="ml-3 px-2.5 py-1 rounded bg-rose-600 text-white text-xs font-medium hover:bg-rose-700"
      >
        Coba Lagi
      </button>
    </div>

    <!-- Loading State -->
    <div v-else-if="loading" class="mt-4 bg-white border border-zinc-200 rounded-lg p-8 text-center">
      <div class="inline-block w-5 h-5 border-2 border-zinc-300 border-t-zinc-900 rounded-full animate-spin"></div>
      <p class="mt-2 text-xs text-zinc-500">Memuat data tiket...</p>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="tickets.length === 0"
      class="mt-4 bg-white border border-zinc-200 rounded-lg p-12 text-center"
    >
      <div class="mx-auto w-10 h-10 rounded-full bg-zinc-100 flex items-center justify-center text-zinc-400">
        <span class="text-xs font-mono">0</span>
      </div>
      <h3 class="mt-3 text-xs font-medium text-zinc-900">Tidak ada tiket ditemukan</h3>
      <p class="mt-1 text-xs text-zinc-500">
        {{ filters.q || filters.status || filters.priority ? 'Coba ubah kata kunci atau reset filter pencarian Anda.' : 'Belum ada tiket yang diajukan saat ini.' }}
      </p>
      <div class="mt-4">
        <button
          v-if="filters.q || filters.status || filters.priority"
          type="button"
          @click="resetFilters"
          class="inline-flex items-center px-3 py-1.5 border border-zinc-200 text-xs font-medium rounded-md text-zinc-700 bg-white hover:bg-zinc-50"
        >
          Bersihkan Filter
        </button>
        <RouterLink
          v-else
          to="/tickets/new"
          class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md text-white bg-zinc-900 hover:bg-zinc-800"
        >
          Buat Tiket Pertama
        </RouterLink>
      </div>
    </div>

    <!-- Data Table & Cards -->
    <div v-else class="mt-4 bg-white border border-zinc-200 rounded-lg overflow-hidden shadow-2xs">
      <!-- Desktop Table -->
      <div class="hidden md:block overflow-x-auto">
        <table class="min-w-full divide-y divide-zinc-200 text-left text-xs">
          <thead class="bg-zinc-50 text-zinc-600 font-medium">
            <tr>
              <th scope="col" class="py-3 px-4 w-16">ID</th>
              <th scope="col" class="py-3 px-4">Judul & Kategori</th>
              <th scope="col" class="py-3 px-4 w-28">Status</th>
              <th scope="col" class="py-3 px-4 w-24">Prioritas</th>
              <th scope="col" class="py-3 px-4 w-40">Pemohon</th>
              <th scope="col" class="py-3 px-4 w-40">Teknisi</th>
              <th scope="col" class="py-3 px-4 w-36 text-right">Dibuat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100 bg-white">
            <tr
              v-for="ticket in tickets"
              :key="ticket.id"
              @click="navigateToDetail(ticket.id)"
              class="hover:bg-zinc-50 cursor-pointer transition-colors"
            >
              <td class="py-3 px-4 font-mono text-zinc-500">
                #{{ ticket.id }}
              </td>
              <td class="py-3 px-4">
                <div class="font-medium text-zinc-900">{{ ticket.title }}</div>
                <div class="text-[11px] text-zinc-500">{{ ticket.category }}</div>
              </td>
              <td class="py-3 px-4">
                <StatusBadge :status="ticket.status" />
              </td>
              <td class="py-3 px-4">
                <PriorityBadge :priority="ticket.priority" />
              </td>
              <td class="py-3 px-4 text-zinc-700">
                {{ ticket.requester?.name || '-' }}
              </td>
              <td class="py-3 px-4 text-zinc-600">
                <span v-if="ticket.assignee" class="text-zinc-800">
                  {{ ticket.assignee.name }}
                </span>
                <span v-else class="text-zinc-400 italic">
                  Belum ditugaskan
                </span>
              </td>
              <td class="py-3 px-4 text-right text-zinc-500 whitespace-nowrap">
                {{ formatDate(ticket.created_at) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Mobile Stacked Card List -->
      <div class="md:hidden divide-y divide-zinc-100">
        <div
          v-for="ticket in tickets"
          :key="ticket.id"
          @click="navigateToDetail(ticket.id)"
          class="p-4 hover:bg-zinc-50 cursor-pointer transition-colors space-y-2"
        >
          <div class="flex items-center justify-between text-xs">
            <span class="font-mono text-zinc-500">#{{ ticket.id }}</span>
            <div class="flex items-center space-x-1.5">
              <PriorityBadge :priority="ticket.priority" />
              <StatusBadge :status="ticket.status" />
            </div>
          </div>

          <div>
            <h4 class="text-xs font-medium text-zinc-900">{{ ticket.title }}</h4>
            <p class="text-[11px] text-zinc-500">{{ ticket.category }}</p>
          </div>

          <div class="flex items-center justify-between pt-1 border-t border-zinc-100 text-[11px] text-zinc-500">
            <span>Teknisi: {{ ticket.assignee?.name || 'Belum ditugaskan' }}</span>
            <span>{{ formatDate(ticket.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <PaginationBar
        :current-page="pagination.page"
        :total-pages="pagination.total_pages"
        :total-count="pagination.total_count"
        :limit="pagination.limit"
        @page-change="handlePageChange"
      />
    </div>
  </div>
</template>
