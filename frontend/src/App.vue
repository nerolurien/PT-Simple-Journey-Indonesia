<script setup>
import { useAuthStore } from '@/stores/auth'
import AppNavbar from '@/components/AppNavbar.vue'

const auth = useAuthStore()
</script>

<template>
  <div class="min-h-screen bg-zinc-50 text-zinc-900 flex flex-col font-sans">
    <AppNavbar v-if="auth.isAuthenticated" />
    <main class="flex-1">
      <router-view />
    </main>
  </div>
</template>
