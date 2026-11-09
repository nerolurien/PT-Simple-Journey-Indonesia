<script setup>
import { computed } from 'vue'

const props = defineProps({
  priority: {
    type: String,
    required: true
  }
})

const badgeConfig = computed(() => {
  switch (props.priority) {
    case 'low':
      return {
        label: 'Low',
        classes: 'bg-zinc-100 text-zinc-600 border-zinc-200'
      }
    case 'medium':
      return {
        label: 'Medium',
        classes: 'bg-sky-50 text-sky-700 border-sky-200'
      }
    case 'high':
      return {
        label: 'High',
        classes: 'bg-amber-50 text-amber-700 border-amber-200'
      }
    case 'urgent':
      return {
        label: 'Urgent',
        classes: 'bg-rose-50 text-rose-700 border-rose-200'
      }
    default:
      return {
        label: props.priority || '-',
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
