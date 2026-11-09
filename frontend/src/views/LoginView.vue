<script setup>
import { ref } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const errorMessage = ref('')
const validationErrors = ref({})

function validate() {
  const errors = {}
  if (!email.value.trim()) {
    errors.email = 'Email wajib diisi'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    errors.email = 'Format email tidak valid'
  }

  if (!password.value) {
    errors.password = 'Password wajib diisi'
  }

  validationErrors.value = errors
  return Object.keys(errors).length === 0
}

async function handleSubmit() {
  errorMessage.value = ''
  if (!validate()) return

  try {
    await auth.login(email.value.trim(), password.value)
    const redirect = route.query.redirect || '/tickets'
    router.push(redirect)
  } catch (err) {
    if (err.response?.data?.error) {
      errorMessage.value = err.response.data.error
    } else {
      errorMessage.value = 'Terjadi kesalahan koneksi ke server'
    }
  }
}

function fillDemo(demoEmail) {
  email.value = demoEmail
  password.value = 'Password123!'
  validationErrors.value = {}
  errorMessage.value = ''
}
</script>

<template>
  <div class="min-h-screen flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <div class="flex items-center justify-center space-x-2">
        <span class="w-3 h-3 rounded-full bg-zinc-900"></span>
        <span class="text-base font-semibold tracking-tight text-zinc-900">Simple Journey Helpdesk</span>
      </div>
      <h2 class="mt-4 text-center text-xl font-medium tracking-tight text-zinc-900">
        Masuk ke akun Anda
      </h2>
      <p class="mt-1 text-center text-xs text-zinc-500">
        Sistem internal tiket dan bantuan teknis operasional
      </p>
    </div>

    <div class="mt-6 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white py-6 px-4 border border-zinc-200 rounded-lg sm:px-8">
        <div
          v-if="errorMessage"
          class="mb-4 p-3 rounded-md bg-rose-50 border border-rose-200 text-xs text-rose-700"
        >
          {{ errorMessage }}
        </div>

        <form class="space-y-4" @submit.prevent="handleSubmit">
          <div>
            <label for="email" class="block text-xs font-medium text-zinc-700">Email</label>
            <div class="mt-1">
              <input
                id="email"
                v-model="email"
                type="email"
                autocomplete="email"
                placeholder="nama@simplejourney.co.id"
                class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.email ? 'border-rose-300' : 'border-zinc-200'"
              />
            </div>
            <p v-if="validationErrors.email" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.email }}
            </p>
          </div>

          <div>
            <label for="password" class="block text-xs font-medium text-zinc-700">Password</label>
            <div class="mt-1">
              <input
                id="password"
                v-model="password"
                type="password"
                autocomplete="current-password"
                class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.password ? 'border-rose-300' : 'border-zinc-200'"
              />
            </div>
            <p v-if="validationErrors.password" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.password }}
            </p>
          </div>

          <div>
            <button
              type="submit"
              :disabled="auth.loading"
              class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md text-xs font-medium text-white bg-zinc-900 hover:bg-zinc-800 disabled:opacity-50 transition-colors"
            >
              {{ auth.loading ? 'Memproses...' : 'Masuk' }}
            </button>
          </div>
        </form>

        <div class="mt-5 border-t border-zinc-100 pt-4 text-center">
          <p class="text-xs text-zinc-600">
            Belum memiliki akun?
            <RouterLink to="/register" class="font-medium text-zinc-900 hover:underline">
              Daftar sebagai requester
            </RouterLink>
          </p>
        </div>

        <div class="mt-6 border-t border-zinc-100 pt-4">
          <p class="text-[11px] font-medium text-zinc-500 uppercase tracking-wider mb-2">
            Akun Percobaan (Seed Data)
          </p>
          <div class="grid grid-cols-3 gap-2">
            <button
              type="button"
              @click="fillDemo('admin@simplejourney.co.id')"
              class="px-2 py-1.5 text-left border border-zinc-200 rounded text-[11px] hover:bg-zinc-50"
            >
              <div class="font-medium text-zinc-800">Admin</div>
              <div class="text-zinc-500 truncate">Budi S.</div>
            </button>
            <button
              type="button"
              @click="fillDemo('agent@simplejourney.co.id')"
              class="px-2 py-1.5 text-left border border-zinc-200 rounded text-[11px] hover:bg-zinc-50"
            >
              <div class="font-medium text-zinc-800">Agent</div>
              <div class="text-zinc-500 truncate">Nadia P.</div>
            </button>
            <button
              type="button"
              @click="fillDemo('requester@simplejourney.co.id')"
              class="px-2 py-1.5 text-left border border-zinc-200 rounded text-[11px] hover:bg-zinc-50"
            >
              <div class="font-medium text-zinc-800">Requester</div>
              <div class="text-zinc-500 truncate">Siti R.</div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
