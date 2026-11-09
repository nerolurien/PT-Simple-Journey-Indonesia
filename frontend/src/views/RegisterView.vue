<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const errorMessage = ref('')
const validationErrors = ref({})

function validate() {
  const errors = {}

  if (!name.value.trim()) {
    errors.name = 'Nama lengkap wajib diisi'
  }

  if (!email.value.trim()) {
    errors.email = 'Email wajib diisi'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    errors.email = 'Format email tidak valid'
  }

  if (!password.value) {
    errors.password = 'Password wajib diisi'
  } else if (password.value.length < 8) {
    errors.password = 'Password minimal 8 karakter'
  }

  if (password.value !== confirmPassword.value) {
    errors.confirmPassword = 'Konfirmasi password tidak cocok'
  }

  validationErrors.value = errors
  return Object.keys(errors).length === 0
}

async function handleSubmit() {
  errorMessage.value = ''
  if (!validate()) return

  try {
    await auth.register(name.value.trim(), email.value.trim(), password.value)
    router.push('/tickets')
  } catch (err) {
    if (err.response?.data?.error) {
      errorMessage.value = err.response.data.error
    } else {
      errorMessage.value = 'Terjadi kesalahan saat pendaftaran'
    }
  }
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
        Pendaftaran Akun Baru
      </h2>
      <p class="mt-1 text-center text-xs text-zinc-500">
        Akun baru otomatis didaftarkan dengan hak akses Requester
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
            <label for="name" class="block text-xs font-medium text-zinc-700">Nama Lengkap</label>
            <div class="mt-1">
              <input
                id="name"
                v-model="name"
                type="text"
                autocomplete="name"
                placeholder="Rizky Ramadhan"
                class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.name ? 'border-rose-300' : 'border-zinc-200'"
              />
            </div>
            <p v-if="validationErrors.name" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.name }}
            </p>
          </div>

          <div>
            <label for="email" class="block text-xs font-medium text-zinc-700">Email Kantor</label>
            <div class="mt-1">
              <input
                id="email"
                v-model="email"
                type="email"
                autocomplete="email"
                placeholder="rizky@simplejourney.co.id"
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
                autocomplete="new-password"
                placeholder="Minimal 8 karakter"
                class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.password ? 'border-rose-300' : 'border-zinc-200'"
              />
            </div>
            <p v-if="validationErrors.password" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.password }}
            </p>
          </div>

          <div>
            <label for="confirmPassword" class="block text-xs font-medium text-zinc-700">Konfirmasi Password</label>
            <div class="mt-1">
              <input
                id="confirmPassword"
                v-model="confirmPassword"
                type="password"
                autocomplete="new-password"
                placeholder="Ulangi password"
                class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.confirmPassword ? 'border-rose-300' : 'border-zinc-200'"
              />
            </div>
            <p v-if="validationErrors.confirmPassword" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.confirmPassword }}
            </p>
          </div>

          <div>
            <button
              type="submit"
              :disabled="auth.loading"
              class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md text-xs font-medium text-white bg-zinc-900 hover:bg-zinc-800 disabled:opacity-50 transition-colors"
            >
              {{ auth.loading ? 'Mendaftarkan...' : 'Daftar Sekarang' }}
            </button>
          </div>
        </form>

        <div class="mt-5 border-t border-zinc-100 pt-4 text-center">
          <p class="text-xs text-zinc-600">
            Sudah memiliki akun?
            <RouterLink to="/login" class="font-medium text-zinc-900 hover:underline">
              Masuk ke sistem
            </RouterLink>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
