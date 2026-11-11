<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '@/lib/api'

const loading = ref(true)
const error = ref('')
const stats = ref(null)

const statusLabels = {
  open: { label: 'Terbuka (Open)', barClass: 'bg-blue-500' },
  in_progress: { label: 'Diproses (In Progress)', barClass: 'bg-amber-500' },
  resolved: { label: 'Selesai (Resolved)', barClass: 'bg-emerald-500' },
  closed: { label: 'Ditutup (Closed)', barClass: 'bg-zinc-400' }
}

const priorityLabels = {
  low: { label: 'Rendah (Low)', barClass: 'bg-zinc-400' },
  medium: { label: 'Sedang (Medium)', barClass: 'bg-blue-400' },
  high: { label: 'Tinggi (High)', barClass: 'bg-amber-500' },
  urgent: { label: 'Mendesak (Urgent)', barClass: 'bg-red-500' }
}

async function fetchStats() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get('/admin/stats')
    stats.value = res.data
  } catch (err) {
    error.value = err.response?.data?.message || 'Gagal memuat statistik sistem.'
  } finally {
    loading.value = false
  }
}

function getPercentage(count, total) {
  if (!total || total === 0) return 0
  return Math.round((count / total) * 100)
}

function formatHours(hours) {
  if (hours === undefined || hours === null) return '0 jam'
  if (hours < 1) {
    const mins = Math.round(hours * 60)
    return `${mins} menit`
  }
  return `${hours.toFixed(1)} jam`
}

onMounted(() => {
  fetchStats()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-zinc-900">Statistik Helpdesk</h1>
        <p class="text-xs text-zinc-500 mt-1">Ringkasan performa tiket, beban kerja agen, dan kepatuhan SLA</p>
      </div>

      <button
        type="button"
        @click="fetchStats"
        :disabled="loading"
        class="inline-flex items-center justify-center px-3 py-1.5 border border-zinc-200 text-xs font-medium rounded-md text-zinc-700 bg-white hover:bg-zinc-50 focus:outline-hidden disabled:opacity-50"
      >
        <svg class="w-3.5 h-3.5 mr-1.5 text-zinc-500" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Segarkan
      </button>
    </div>

    <div v-if="error" class="p-4 bg-red-50 border border-red-200 rounded-lg text-xs text-red-700">
      {{ error }}
    </div>

    <div v-if="loading && !stats" class="grid grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-24 bg-zinc-100 rounded-lg border border-zinc-200"></div>
    </div>

    <div v-else-if="stats" class="space-y-6">
      <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
        <div class="p-4 bg-white border border-zinc-200 rounded-lg">
          <p class="text-xs font-medium text-zinc-500">Total Tiket</p>
          <p class="text-2xl font-bold text-zinc-900 mt-2">{{ stats.total_tickets }}</p>
          <p class="text-[11px] text-zinc-400 mt-1">Seluruh tiket terdaftar</p>
        </div>

        <div class="p-4 bg-white border border-zinc-200 rounded-lg">
          <p class="text-xs font-medium text-zinc-500">Tiket Terbuka</p>
          <p class="text-2xl font-bold text-blue-600 mt-2">
            {{ stats.status_breakdown.find(s => s.status === 'open')?.count || 0 }}
          </p>
          <p class="text-[11px] text-zinc-400 mt-1">Menunggu penanganan</p>
        </div>

        <div class="p-4 bg-white border border-zinc-200 rounded-lg">
          <p class="text-xs font-medium text-zinc-500">Sedang Diproses</p>
          <p class="text-2xl font-bold text-amber-600 mt-2">
            {{ stats.status_breakdown.find(s => s.status === 'in_progress')?.count || 0 }}
          </p>
          <p class="text-[11px] text-zinc-400 mt-1">Dalam pengerjaan agen</p>
        </div>

        <div class="p-4 bg-white border border-zinc-200 rounded-lg" :class="{ 'border-red-300 bg-red-50/30': stats.sla_breached_count > 0 }">
          <p class="text-xs font-medium" :class="stats.sla_breached_count > 0 ? 'text-red-700 font-semibold' : 'text-zinc-500'">
            SLA Terlambat
          </p>
          <p class="text-2xl font-bold mt-2" :class="stats.sla_breached_count > 0 ? 'text-red-600' : 'text-zinc-900'">
            {{ stats.sla_breached_count }}
          </p>
          <p class="text-[11px] text-zinc-400 mt-1">Urgent &gt; 2 jam belum diproses</p>
        </div>

        <div class="p-4 bg-white border border-zinc-200 rounded-lg col-span-2 lg:col-span-1">
          <p class="text-xs font-medium text-zinc-500">Rata-rata Resolusi</p>
          <p class="text-2xl font-bold text-zinc-900 mt-2">{{ formatHours(stats.avg_resolution_hours) }}</p>
          <p class="text-[11px] text-zinc-400 mt-1">Waktu hingga diselesaikan</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div class="p-5 bg-white border border-zinc-200 rounded-lg">
          <h2 class="text-sm font-semibold text-zinc-900 mb-4">Distribusi Status Tiket</h2>
          <div class="space-y-3">
            <div v-for="item in stats.status_breakdown" :key="item.status">
              <div class="flex items-center justify-between text-xs mb-1">
                <span class="font-medium text-zinc-700">{{ statusLabels[item.status]?.label || item.status }}</span>
                <span class="text-zinc-500">{{ item.count }} ({{ getPercentage(item.count, stats.total_tickets) }}%)</span>
              </div>
              <div class="w-full bg-zinc-100 rounded-full h-2 overflow-hidden">
                <div
                  class="h-2 rounded-full transition-all duration-300"
                  :class="statusLabels[item.status]?.barClass || 'bg-zinc-400'"
                  :style="{ width: `${getPercentage(item.count, stats.total_tickets)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="p-5 bg-white border border-zinc-200 rounded-lg">
          <h2 class="text-sm font-semibold text-zinc-900 mb-4">Distribusi Prioritas</h2>
          <div class="space-y-3">
            <div v-for="item in stats.priority_breakdown" :key="item.priority">
              <div class="flex items-center justify-between text-xs mb-1">
                <span class="font-medium text-zinc-700">{{ priorityLabels[item.priority]?.label || item.priority }}</span>
                <span class="text-zinc-500">{{ item.count }} ({{ getPercentage(item.count, stats.total_tickets) }}%)</span>
              </div>
              <div class="w-full bg-zinc-100 rounded-full h-2 overflow-hidden">
                <div
                  class="h-2 rounded-full transition-all duration-300"
                  :class="priorityLabels[item.priority]?.barClass || 'bg-zinc-400'"
                  :style="{ width: `${getPercentage(item.count, stats.total_tickets)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white border border-zinc-200 rounded-lg overflow-hidden">
        <div class="px-5 py-4 border-b border-zinc-200">
          <h2 class="text-sm font-semibold text-zinc-900">Performa Penyelesaian Agen</h2>
          <p class="text-xs text-zinc-500 mt-0.5">Jumlah tiket yang berhasil diselesaikan oleh masing-masing agen</p>
        </div>

        <div v-if="!stats.top_agents || stats.top_agents.length === 0" class="p-6 text-center text-xs text-zinc-400">
          Belum ada data penyelesaian tiket oleh staf agen.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-zinc-50 border-b border-zinc-200 text-zinc-500">
              <tr>
                <th class="px-5 py-3 font-medium">Agen</th>
                <th class="px-5 py-3 font-medium">Email</th>
                <th class="px-5 py-3 font-medium text-right">Tiket Diselesaikan</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-zinc-100">
              <tr v-for="agent in stats.top_agents" :key="agent.id" class="hover:bg-zinc-50/60">
                <td class="px-5 py-3 font-medium text-zinc-900">{{ agent.name }}</td>
                <td class="px-5 py-3 text-zinc-500">{{ agent.email }}</td>
                <td class="px-5 py-3 text-right">
                  <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                    {{ agent.resolved_count }} tiket
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
