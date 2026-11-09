<script setup>
import { computed } from 'vue'

const props = defineProps({
  currentPage: {
    type: Number,
    required: true,
    default: 1
  },
  totalPages: {
    type: Number,
    required: true,
    default: 1
  },
  totalCount: {
    type: Number,
    required: true,
    default: 0
  },
  limit: {
    type: Number,
    required: true,
    default: 10
  }
})

const emit = defineEmits(['page-change'])

const startItem = computed(() => {
  if (props.totalCount === 0) return 0
  return (props.currentPage - 1) * props.limit + 1
})

const endItem = computed(() => {
  const calculated = props.currentPage * props.limit
  return calculated > props.totalCount ? props.totalCount : calculated
})

function goToPage(page) {
  if (page >= 1 && page <= props.totalPages && page !== props.currentPage) {
    emit('page-change', page)
  }
}
</script>

<template>
  <div class="flex items-center justify-between px-4 py-3 bg-white border-t border-zinc-200 sm:px-6">
    <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
      <div>
        <p class="text-xs text-zinc-600">
          Menampilkan
          <span class="font-medium text-zinc-900">{{ startItem }}</span>
          sampai
          <span class="font-medium text-zinc-900">{{ endItem }}</span>
          dari
          <span class="font-medium text-zinc-900">{{ totalCount }}</span>
          data
        </p>
      </div>
      <div>
        <nav class="relative z-0 inline-flex rounded-md shadow-xs -space-x-px" aria-label="Pagination">
          <button
            type="button"
            :disabled="currentPage <= 1"
            @click="goToPage(currentPage - 1)"
            class="relative inline-flex items-center px-3 py-1.5 rounded-l-md border border-zinc-200 bg-white text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Sebelumnya
          </button>
          <span class="relative inline-flex items-center px-3 py-1.5 border border-zinc-200 bg-zinc-50 text-xs font-medium text-zinc-700">
            Halaman {{ currentPage }} dari {{ totalPages || 1 }}
          </span>
          <button
            type="button"
            :disabled="currentPage >= totalPages"
            @click="goToPage(currentPage + 1)"
            class="relative inline-flex items-center px-3 py-1.5 rounded-r-md border border-zinc-200 bg-white text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Berikutnya
          </button>
        </nav>
      </div>
    </div>
    <div class="flex items-center justify-between w-full sm:hidden">
      <button
        type="button"
        :disabled="currentPage <= 1"
        @click="goToPage(currentPage - 1)"
        class="inline-flex items-center px-3 py-1.5 rounded border border-zinc-200 bg-white text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Sebelumnya
      </button>
      <span class="text-xs text-zinc-600">
        {{ currentPage }} / {{ totalPages || 1 }}
      </span>
      <button
        type="button"
        :disabled="currentPage >= totalPages"
        @click="goToPage(currentPage + 1)"
        class="inline-flex items-center px-3 py-1.5 rounded border border-zinc-200 bg-white text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Berikutnya
      </button>
    </div>
  </div>
</template>
