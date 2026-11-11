<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '@/lib/api'

const loading = ref(true)
const error = ref('')
const stats = ref(null)
const lastRefreshedAt = ref(new Date())

const statusConfig = {
  open: { label: 'Terbuka', tag: 'Open', colorHex: '#3b82f6', bgClass: 'bg-blue-500', badgeClass: 'bg-blue-50 text-blue-700 border-blue-200' },
  in_progress: { label: 'Diproses', tag: 'In Progress', colorHex: '#f59e0b', bgClass: 'bg-amber-500', badgeClass: 'bg-amber-50 text-amber-700 border-amber-200' },
  resolved: { label: 'Selesai', tag: 'Resolved', colorHex: '#10b981', bgClass: 'bg-emerald-500', badgeClass: 'bg-emerald-50 text-emerald-700 border-emerald-200' },
  closed: { label: 'Ditutup', tag: 'Closed', colorHex: '#71717a', bgClass: 'bg-zinc-500', badgeClass: 'bg-zinc-100 text-zinc-700 border-zinc-200' }
}

const priorityConfig = {
  urgent: { label: 'Mendesak', tag: 'Urgent', bgClass: 'bg-red-500', badgeClass: 'bg-red-50 text-red-700 border-red-200', fillClass: 'bg-red-500/10' },
  high: { label: 'Tinggi', tag: 'High', bgClass: 'bg-amber-500', badgeClass: 'bg-amber-50 text-amber-700 border-amber-200', fillClass: 'bg-amber-500/10' },
  medium: { label: 'Sedang', tag: 'Medium', bgClass: 'bg-blue-500', badgeClass: 'bg-blue-50 text-blue-700 border-blue-200', fillClass: 'bg-blue-500/10' },
  low: { label: 'Rendah', tag: 'Low', bgClass: 'bg-zinc-400', badgeClass: 'bg-zinc-100 text-zinc-700 border-zinc-200', fillClass: 'bg-zinc-400/10' }
}

async function fetchStats() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get('/admin/stats')
    stats.value = res.data
    lastRefreshedAt.value = new Date()
  } catch (err) {
    error.value = err.response?.data?.message || 'Gagal memuat statistik sistem.'
  } finally {
    loading.value = false
  }
}

const lastRefreshedText = computed(() => {
  return lastRefreshedAt.value.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
})

const activeTicketsCount = computed(() => {
  if (!stats.value) return 0
  const open = stats.value.status_breakdown?.find(s => s.status === 'open')?.count || 0
  const inProgress = stats.value.status_breakdown?.find(s => s.status === 'in_progress')?.count || 0
  return open + inProgress
})

const completedTicketsCount = computed(() => {
  if (!stats.value) return 0
  const resolved = stats.value.status_breakdown?.find(s => s.status === 'resolved')?.count || 0
  const closed = stats.value.status_breakdown?.find(s => s.status === 'closed')?.count || 0
  return resolved + closed
})

const resolutionRate = computed(() => {
  if (!stats.value || !stats.value.total_tickets) return 0
  return Math.round((completedTicketsCount.value / stats.value.total_tickets) * 100)
})

function formatResolutionTime(hours) {
  if (hours === undefined || hours === null || hours <= 0) return '0 menit'
  const totalMinutes = Math.round(hours * 60)
  if (totalMinutes < 60) {
    return `${totalMinutes} menit`
  }
  const h = Math.floor(totalMinutes / 60)
  const m = totalMinutes % 60
  if (m === 0) {
    return `${h} jam`
  }
  return `${h} jam ${m} mnt`
}

function getPercentage(count, total) {
  if (!total || total === 0) return 0
  return Math.round((count / total) * 100)
}

const donutSegments = computed(() => {
  if (!stats.value || !stats.value.total_tickets) return []
  const total = stats.value.total_tickets
  let accumulated = 0

  const order = ['open', 'in_progress', 'resolved', 'closed']
  return order.map((statusKey) => {
    const item = stats.value.status_breakdown?.find(s => s.status === statusKey)
    const count = item ? item.count : 0
    const pct = total > 0 ? (count / total) * 100 : 0
    const offset = accumulated
    accumulated += pct

    return {
      status: statusKey,
      count,
      pct,
      offset,
      config: statusConfig[statusKey]
    }
  })
})

const priorityRows = computed(() => {
  if (!stats.value) return []
  const total = stats.value.total_tickets || 0
  const order = ['urgent', 'high', 'medium', 'low']

  return order.map((pKey) => {
    const item = stats.value.priority_breakdown?.find(p => p.priority === pKey)
    const count = item ? item.count : 0
    const pct = total > 0 ? Math.round((count / total) * 100) : 0
    return {
      priority: pKey,
      count,
      pct,
      config: priorityConfig[pKey]
    }
  })
})

onMounted(() => {
  fetchStats()
})
</script>

<template>
  <div class="min-h-screen bg-zinc-50/70 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
      <!-- Header Bar -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-zinc-200/80 pb-5">
        <div>
          <h1 class="text-xl font-bold tracking-tight text-zinc-900">Statistik & Analitik Helpdesk</h1>
          <div class="flex items-center gap-2 mt-1">
            <p class="text-xs text-zinc-500">Ringkasan performa tiket, beban kerja staf agen, dan kepatuhan SLA</p>
            <span class="text-zinc-300">·</span>
            <div class="flex items-center text-[11px] text-zinc-400">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 inline-block mr-1.5"></span>
              Diperbarui {{ lastRefreshedText }}
            </div>
          </div>
        </div>

        <div>
          <button
            type="button"
            @click="fetchStats"
            :disabled="loading"
            class="inline-flex items-center justify-center px-3.5 py-1.5 border border-zinc-200/90 text-xs font-medium rounded-md text-zinc-700 bg-white hover:bg-zinc-50 focus:outline-hidden shadow-2xs transition-colors disabled:opacity-50"
          >
            <svg
              class="w-3.5 h-3.5 mr-1.5 text-zinc-500"
              :class="{ 'animate-spin': loading }"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Segarkan Data
          </button>
        </div>
      </div>

      <!-- Error Alert -->
      <div v-if="error" class="p-4 bg-red-50 border border-red-200 rounded-lg text-xs text-red-700">
        {{ error }}
      </div>

      <!-- Loading Skeleton -->
      <div v-if="loading && !stats" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
        <div v-for="i in 4" :key="i" class="h-28 bg-zinc-100 rounded-lg border border-zinc-200/80"></div>
      </div>

      <div v-else-if="stats" class="space-y-6">
        <!-- 4 Top KPI Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <!-- Card 1: Total Tiket -->
          <div class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs flex flex-col justify-between">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium uppercase tracking-wider text-zinc-500">Total Tiket</span>
              <div class="w-8 h-8 rounded-md bg-zinc-100 flex items-center justify-center text-zinc-600">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
              </div>
            </div>
            <div class="my-2">
              <span class="text-3xl font-semibold tracking-tight text-zinc-900 font-mono">{{ stats.total_tickets }}</span>
            </div>
            <div class="pt-1">
              <span class="inline-flex items-center text-[11px] font-medium text-zinc-600 bg-zinc-100 px-2 py-0.5 rounded-full">
                {{ activeTicketsCount }} aktif · {{ completedTicketsCount }} selesai
              </span>
            </div>
          </div>

          <!-- Card 2: Tingkat Penyelesaian -->
          <div class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs flex flex-col justify-between">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium uppercase tracking-wider text-zinc-500">Tingkat Penyelesaian</span>
              <div class="w-8 h-8 rounded-md bg-zinc-100 flex items-center justify-center text-zinc-600">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
            <div class="my-2">
              <span class="text-3xl font-semibold tracking-tight text-zinc-900 font-mono">{{ resolutionRate }}%</span>
            </div>
            <div class="pt-1 space-y-1">
              <div class="w-full bg-zinc-100 rounded-full h-1.5 overflow-hidden">
                <div
                  class="bg-emerald-500 h-1.5 rounded-full transition-all duration-500"
                  :style="{ width: `${resolutionRate}%` }"
                ></div>
              </div>
              <p class="text-[11px] text-zinc-400">{{ completedTicketsCount }} dari {{ stats.total_tickets }} tiket diselesaikan</p>
            </div>
          </div>

          <!-- Card 3: Kepatuhan SLA Urgent -->
          <div
            class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs flex flex-col justify-between"
            :class="stats.sla_breached_count > 0 ? 'border-l-4 border-l-red-500' : 'border-l-4 border-l-emerald-500'"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium uppercase tracking-wider text-zinc-500">SLA Terlambat (Urgent)</span>
              <div class="w-8 h-8 rounded-md bg-zinc-100 flex items-center justify-center text-zinc-600">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
            </div>
            <div class="my-2 flex items-baseline gap-2">
              <span
                class="text-3xl font-semibold tracking-tight font-mono"
                :class="stats.sla_breached_count > 0 ? 'text-red-600' : 'text-zinc-900'"
              >
                {{ stats.sla_breached_count }}
              </span>
              <span class="text-xs text-zinc-400 font-normal">tiket</span>
            </div>
            <div class="pt-1 flex items-center justify-between">
              <span
                v-if="stats.sla_breached_count > 0"
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-red-50 text-red-700 border border-red-200"
              >
                Perlu Tindakan Segera
              </span>
              <span
                v-else
                class="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-50 text-emerald-700 border border-emerald-200"
              >
                Semua Terkendali
              </span>
              <span class="text-[11px] text-zinc-400">Target &le; 2 jam</span>
            </div>
          </div>

          <!-- Card 4: Rata-rata Resolusi -->
          <div class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs flex flex-col justify-between">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium uppercase tracking-wider text-zinc-500">Rata-rata Resolusi</span>
              <div class="w-8 h-8 rounded-md bg-zinc-100 flex items-center justify-center text-zinc-600">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
            <div class="my-2">
              <span class="text-3xl font-semibold tracking-tight text-zinc-900 font-mono">
                {{ formatResolutionTime(stats.avg_resolution_hours) }}
              </span>
            </div>
            <div class="pt-1">
              <p class="text-[11px] text-zinc-400">Dari tiket dibuat hingga status selesai</p>
            </div>
          </div>
        </div>

        <!-- 2 Analytical Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Left: Status Distribution with SVG Donut Chart -->
          <div class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs">
            <div class="border-b border-zinc-100 pb-3 mb-4">
              <h2 class="text-sm font-semibold text-zinc-900">Distribusi Status Tiket</h2>
              <p class="text-xs text-zinc-500 mt-0.5">Proporsi tiket berdasarkan tahapan penanganan saat ini</p>
            </div>

            <div class="flex flex-col sm:flex-row items-center gap-6 pt-2">
              <!-- SVG Donut Chart -->
              <div class="relative w-36 h-36 flex-shrink-0">
                <svg viewBox="0 0 36 36" class="w-full h-full -rotate-90 transform">
                  <!-- Base background track -->
                  <circle
                    cx="18"
                    cy="18"
                    r="15.91549430918954"
                    fill="none"
                    stroke="#f4f4f5"
                    stroke-width="3.5"
                  />
                  <!-- Colored segments -->
                  <circle
                    v-for="seg in donutSegments"
                    :key="seg.status"
                    cx="18"
                    cy="18"
                    r="15.91549430918954"
                    fill="none"
                    :stroke="seg.config.colorHex"
                    stroke-width="3.6"
                    :stroke-dasharray="`${seg.pct} ${100 - seg.pct}`"
                    :stroke-dashoffset="`${-seg.offset}`"
                    class="transition-all duration-500"
                  />
                </svg>
                <!-- Central metric text -->
                <div class="absolute inset-0 flex flex-col items-center justify-center text-center pointer-events-none">
                  <span class="text-2xl font-bold font-mono text-zinc-900 leading-tight">{{ stats.total_tickets }}</span>
                  <span class="text-[10px] font-medium uppercase tracking-wider text-zinc-400">Tiket</span>
                </div>
              </div>

              <!-- Status Legend Table -->
              <div class="flex-1 w-full space-y-2.5">
                <div
                  v-for="seg in donutSegments"
                  :key="seg.status"
                  class="space-y-1"
                >
                  <div class="flex items-center justify-between text-xs">
                    <div class="flex items-center gap-2">
                      <span
                        class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                        :style="{ backgroundColor: seg.config.colorHex }"
                      ></span>
                      <span class="font-medium text-zinc-700">{{ seg.config.label }}</span>
                      <span class="text-[11px] text-zinc-400">({{ seg.config.tag }})</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="font-mono font-semibold text-zinc-900">{{ seg.count }}</span>
                      <span class="text-[11px] font-mono text-zinc-500 bg-zinc-100 px-1.5 py-0.5 rounded">
                        {{ Math.round(seg.pct) }}%
                      </span>
                    </div>
                  </div>
                  <div class="w-full bg-zinc-100 rounded-full h-1.5 overflow-hidden">
                    <div
                      class="h-1.5 rounded-full transition-all duration-300"
                      :class="seg.config.bgClass"
                      :style="{ width: `${seg.pct}%` }"
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Right: Priority Breakdown with Segmented Stacked Bar -->
          <div class="bg-white border border-zinc-200/80 rounded-lg p-5 shadow-2xs">
            <div class="border-b border-zinc-100 pb-3 mb-4">
              <h2 class="text-sm font-semibold text-zinc-900">Distribusi & Prioritas Tiket</h2>
              <p class="text-xs text-zinc-500 mt-0.5">Komposisi tingkat urgensi penanganan tiket</p>
            </div>

            <!-- Segmented Horizontal Stacked Bar -->
            <div class="mb-4">
              <div class="h-3 w-full rounded-full flex gap-1 overflow-hidden bg-zinc-100 p-0.5">
                <div
                  v-for="p in priorityRows"
                  :key="p.priority"
                  class="h-full rounded-full transition-all duration-500"
                  :class="p.config.bgClass"
                  :style="{ width: `${p.pct}%` }"
                  :title="`${p.config.label}: ${p.count} tiket (${p.pct}%)`"
                ></div>
              </div>
            </div>

            <!-- Structured Priority Rows -->
            <div class="space-y-2">
              <div
                v-for="p in priorityRows"
                :key="p.priority"
                class="relative overflow-hidden p-2.5 rounded-md border border-zinc-100 bg-zinc-50/50 hover:bg-zinc-50 transition-colors flex items-center justify-between"
              >
                <!-- Subtle background fill indicator -->
                <div
                  class="absolute inset-y-0 left-0 transition-all duration-300 pointer-events-none"
                  :class="p.config.fillClass"
                  :style="{ width: `${p.pct}%` }"
                ></div>

                <div class="relative z-10 flex items-center gap-2">
                  <span
                    class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border"
                    :class="p.config.badgeClass"
                  >
                    {{ p.config.label }}
                  </span>
                  <span class="text-[11px] text-zinc-400">({{ p.config.tag }})</span>
                </div>

                <div class="relative z-10 flex items-center gap-2">
                  <span class="font-mono text-xs font-semibold text-zinc-900">{{ p.count }} tiket</span>
                  <span class="text-[11px] font-mono text-zinc-500 bg-white border border-zinc-200/70 px-1.5 py-0.2 rounded">
                    {{ p.pct }}%
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Agent Workload & Capacity Table -->
        <div class="bg-white border border-zinc-200/80 rounded-lg overflow-hidden shadow-2xs">
          <div class="px-5 py-4 border-b border-zinc-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
            <div>
              <h2 class="text-sm font-semibold text-zinc-900">Performa & Beban Kerja Staf Agen</h2>
              <p class="text-xs text-zinc-500 mt-0.5">Distribusi penanganan tiket dan rasio efisiensi penyelesaian per staf</p>
            </div>
            <span
              v-if="stats.top_agents && stats.top_agents.length > 0"
              class="text-xs font-medium text-zinc-500 bg-zinc-100 px-2.5 py-1 rounded-md self-start sm:self-auto"
            >
              {{ stats.top_agents.length }} Staf Terdaftar
            </span>
          </div>

          <div v-if="!stats.top_agents || stats.top_agents.length === 0" class="p-8 text-center text-xs text-zinc-400">
            Belum ada data staf agen terdaftar dalam sistem.
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-zinc-50/80 border-b border-zinc-200/80 text-zinc-500 uppercase tracking-wider font-medium text-[11px]">
                <tr>
                  <th class="px-5 py-3">Agen</th>
                  <th class="px-5 py-3">Peran</th>
                  <th class="px-5 py-3 text-center">Sedang Ditangani</th>
                  <th class="px-5 py-3 text-center">Tiket Selesai</th>
                  <th class="px-5 py-3 text-right">Rasio Penyelesaian</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-zinc-100">
                <tr v-for="agent in stats.top_agents" :key="agent.id" class="hover:bg-zinc-50/70 transition-colors">
                  <!-- Name & Email with Initial Avatar -->
                  <td class="px-5 py-3.5">
                    <div class="flex items-center gap-3">
                      <div class="w-8 h-8 rounded-full bg-zinc-100 border border-zinc-200 text-zinc-700 font-semibold flex items-center justify-center text-xs flex-shrink-0">
                        {{ agent.name.charAt(0).toUpperCase() }}
                      </div>
                      <div>
                        <div class="font-medium text-zinc-900">{{ agent.name }}</div>
                        <div class="text-[11px] text-zinc-400 font-mono">{{ agent.email }}</div>
                      </div>
                    </div>
                  </td>

                  <!-- Role Badge -->
                  <td class="px-5 py-3.5">
                    <span
                      class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium border"
                      :class="agent.role === 'admin' ? 'bg-purple-50 text-purple-700 border-purple-200' : 'bg-blue-50 text-blue-700 border-blue-200'"
                    >
                      {{ agent.role === 'admin' ? 'Admin' : 'Agent' }}
                    </span>
                  </td>

                  <!-- In Progress -->
                  <td class="px-5 py-3.5 text-center">
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-amber-50 text-amber-700 border border-amber-200">
                      {{ agent.in_progress_count }} tiket
                    </span>
                  </td>

                  <!-- Resolved -->
                  <td class="px-5 py-3.5 text-center">
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                      {{ agent.resolved_count }} tiket
                    </span>
                  </td>

                  <!-- Completion Ratio with Progress Bar -->
                  <td class="px-5 py-3.5 text-right whitespace-nowrap">
                    <div class="inline-flex items-center justify-end gap-2">
                      <div class="w-24 sm:w-28 bg-zinc-100 rounded-full h-2 overflow-hidden inline-block">
                        <div
                          class="bg-emerald-500 h-2 rounded-full transition-all duration-300"
                          :style="{ width: `${Math.min(100, Math.max(0, agent.completion_rate))}%` }"
                        ></div>
                      </div>
                      <span class="font-mono text-xs font-semibold text-zinc-900 w-12 text-right">
                        {{ agent.completion_rate.toFixed(1) }}%
                      </span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
