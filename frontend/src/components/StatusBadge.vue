<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: {
    type: String,
    required: true
  }
})

const badgeConfig = computed(() => {
  switch (props.status) {
    case 'open':
      return {
        label: 'Open',
        classes: 'bg-amber-50 text-amber-700 border-amber-200'
      }
    case 'in_progress':
      return {
        label: 'In Progress',
        classes: 'bg-blue-50 text-blue-700 border-blue-200'
      }
    case 'resolved':
      return {
        label: 'Resolved',
        classes: 'bg-emerald-50 text-emerald-700 border-emerald-200'
      }
    case 'closed':
      return {
        label: 'Closed',
        classes: 'bg-zinc-100 text-zinc-600 border-zinc-200'
      }
    default:
      return {
        label: props.status || '-',
        classes: 'bg-zinc-100 text-zinc-600 border-zinc-200'
      }
  }
})
</script>

<template>
  <span
    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border"
    :class="badgeConfig.classes"
  >
    {{ badgeConfig.label }}
  </span>
</template>
