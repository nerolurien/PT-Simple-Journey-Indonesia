<script setup>
import { ref, reactive } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import api from '@/lib/api'

const router = useRouter()

const loading = ref(false)
const serverError = ref('')
const validationErrors = ref({})
const selectedFile = ref(null)
const fileError = ref('')

function handleFileChange(event) {
  fileError.value = ''
  const file = event.target.files[0]
  if (!file) {
    selectedFile.value = null
    return
  }

  const allowedExts = ['.png', '.jpg', '.jpeg', '.pdf', '.txt', '.log']
  const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase()
  if (!allowedExts.includes(ext)) {
    fileError.value = 'Tipe file tidak didukung. Format yang diizinkan: PNG, JPG, PDF, TXT, LOG.'
    event.target.value = ''
    selectedFile.value = null
    return
  }

  if (file.size > 5 * 1024 * 1024) {
    fileError.value = 'Ukuran file melebihi batas maksimal 5 MB.'
    event.target.value = ''
    selectedFile.value = null
    return
  }

  selectedFile.value = file
}

const form = reactive({
  title: '',
  category: 'Software',
  priority: 'medium',
  description: ''
})

const categories = [
  'Software',
  'Hardware',
  'Network',
  'Account & Access',
  'Email & Workspace',
  'Lainnya'
]

const priorities = [
  { value: 'low', label: 'Low', desc: 'Pertanyaan umum atau kendala minor' },
  { value: 'medium', label: 'Medium', desc: 'Kendala standar dengan solusi sementara' },
  { value: 'high', label: 'High', desc: 'Kendala yang menghambat pekerjaan utama' },
  { value: 'urgent', label: 'Urgent', desc: 'Kendala kritis darurat yang menghentikan operasional' }
]

function validate() {
  const errors = {}

  if (!form.title.trim()) {
    errors.title = 'Judul tiket wajib diisi'
  } else if (form.title.trim().length < 5) {
    errors.title = 'Judul tiket minimal 5 karakter'
  }

  if (!form.category) {
    errors.category = 'Kategori wajib dipilih'
  }

  if (!form.priority) {
    errors.priority = 'Prioritas wajib dipilih'
  }

  if (!form.description.trim()) {
    errors.description = 'Deskripsi kendala wajib diisi'
  } else if (form.description.trim().length < 10) {
    errors.description = 'Deskripsi minimal 10 karakter agar teknisi memahami kendala'
  }

  validationErrors.value = errors
  return Object.keys(errors).length === 0
}

async function handleSubmit() {
  serverError.value = ''
  if (!validate()) return

  loading.value = true
  try {
    const payload = {
      title: form.title.trim(),
      category: form.category,
      priority: form.priority,
      description: form.description.trim()
    }

    const res = await api.post('/tickets', payload)

    if (selectedFile.value) {
      const formData = new FormData()
      formData.append('file', selectedFile.value)
      try {
        await api.post(`/tickets/${res.data.id}/attachments`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' }
        })
      } catch (uploadErr) {
        console.error('Lampiran gagal diunggah:', uploadErr)
      }
    }

    router.push(`/tickets/${res.data.id}`)
  } catch (err) {
    if (err.response?.data?.error) {
      serverError.value = err.response.data.error
    } else {
      serverError.value = 'Gagal membuat tiket. Silakan periksa kembali isian form Anda.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
    <div class="mb-6">
      <RouterLink
        to="/tickets"
        class="inline-flex items-center text-xs text-zinc-500 hover:text-zinc-900 transition-colors"
      >
        <span class="mr-1">&larr;</span> Kembali ke Daftar Tiket
      </RouterLink>
      <h1 class="mt-2 text-lg font-semibold tracking-tight text-zinc-900">Buat Tiket Bantuan Baru</h1>
      <p class="mt-1 text-xs text-zinc-500">
        Jelaskan kendala teknis atau kebutuhan operasional Anda secara rinci untuk penanganan tim IT
      </p>
    </div>

    <div class="bg-white border border-zinc-200 rounded-lg shadow-2xs overflow-hidden">
      <div
        v-if="serverError"
        class="m-6 mb-0 p-3 rounded-md bg-rose-50 border border-rose-200 text-xs text-rose-700"
      >
        {{ serverError }}
      </div>

      <form @submit.prevent="handleSubmit" class="p-6 space-y-5">
        <div>
          <label for="title" class="block text-xs font-medium text-zinc-700">
            Judul Tiket <span class="text-rose-500">*</span>
          </label>
          <div class="mt-1">
            <input
              id="title"
              v-model="form.title"
              type="text"
              placeholder="Contoh: Akses VPN kantor terputus setelah update OS"
              class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
              :class="validationErrors.title ? 'border-rose-300' : 'border-zinc-200'"
            />
          </div>
          <p v-if="validationErrors.title" class="mt-1 text-xs text-rose-600">
            {{ validationErrors.title }}
          </p>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label for="category" class="block text-xs font-medium text-zinc-700">
              Kategori <span class="text-rose-500">*</span>
            </label>
            <div class="mt-1">
              <select
                id="category"
                v-model="form.category"
                class="block w-full px-3 py-2 text-xs border rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.category ? 'border-rose-300' : 'border-zinc-200'"
              >
                <option v-for="cat in categories" :key="cat" :value="cat">
                  {{ cat }}
                </option>
              </select>
            </div>
            <p v-if="validationErrors.category" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.category }}
            </p>
          </div>

          <div>
            <label for="priority" class="block text-xs font-medium text-zinc-700">
              Tingkat Prioritas <span class="text-rose-500">*</span>
            </label>
            <div class="mt-1">
              <select
                id="priority"
                v-model="form.priority"
                class="block w-full px-3 py-2 text-xs border rounded-md bg-white text-zinc-800 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
                :class="validationErrors.priority ? 'border-rose-300' : 'border-zinc-200'"
              >
                <option v-for="p in priorities" :key="p.value" :value="p.value">
                  {{ p.label }} — {{ p.desc }}
                </option>
              </select>
            </div>
            <p v-if="validationErrors.priority" class="mt-1 text-xs text-rose-600">
              {{ validationErrors.priority }}
            </p>
          </div>
        </div>

        <div>
          <label for="description" class="block text-xs font-medium text-zinc-700">
            Deskripsi Rinci Kendala <span class="text-rose-500">*</span>
          </label>
          <div class="mt-1">
            <textarea
              id="description"
              v-model="form.description"
              rows="6"
              placeholder="Jelaskan pesan error yang muncul, perangkat yang digunakan, serta langkah yang sudah dicoba..."
              class="block w-full px-3 py-2 text-xs border rounded-md placeholder-zinc-400 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 focus:border-zinc-900"
              :class="validationErrors.description ? 'border-rose-300' : 'border-zinc-200'"
            ></textarea>
          </div>
          <p v-if="validationErrors.description" class="mt-1 text-xs text-rose-600">
            {{ validationErrors.description }}
          </p>
        </div>

        <div>
          <label for="attachment" class="block text-xs font-medium text-zinc-700">
            Lampiran File (Opsional)
          </label>
          <div class="mt-1">
            <input
              id="attachment"
              type="file"
              accept=".png,.jpg,.jpeg,.pdf,.txt,.log"
              @change="handleFileChange"
              class="block w-full text-xs text-zinc-500 file:mr-3 file:py-1.5 file:px-3 file:rounded-md file:border-0 file:text-xs file:font-medium file:bg-zinc-100 file:text-zinc-700 hover:file:bg-zinc-200 cursor-pointer"
            />
          </div>
          <p class="mt-1 text-[11px] text-zinc-400">
            Format diizinkan: PNG, JPG, PDF, TXT, LOG. Maksimal 5 MB.
          </p>
          <p v-if="fileError" class="mt-1 text-xs text-rose-600">
            {{ fileError }}
          </p>
        </div>

        <div class="pt-4 border-t border-zinc-100 flex items-center justify-end space-x-3">
          <RouterLink
            to="/tickets"
            class="px-3 py-2 border border-zinc-200 rounded-md text-xs font-medium text-zinc-700 bg-white hover:bg-zinc-50"
          >
            Batal
          </RouterLink>
          <button
            type="submit"
            :disabled="loading"
            class="px-4 py-2 border border-transparent rounded-md text-xs font-medium text-white bg-zinc-900 hover:bg-zinc-800 disabled:opacity-50 transition-colors"
          >
            {{ loading ? 'Menyimpan Tiket...' : 'Kirim Tiket' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
