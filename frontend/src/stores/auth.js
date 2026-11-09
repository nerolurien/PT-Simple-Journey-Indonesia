import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/lib/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const loading = ref(false)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isAgent = computed(() => user.value?.role === 'agent')
  const isRequester = computed(() => user.value?.role === 'requester')

  async function login(email, password) {
    loading.value = true
    try {
      const res = await api.post('/auth/login', { email, password })
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('user', JSON.stringify(res.data.user))
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function register(name, email, password) {
    loading.value = true
    try {
      const res = await api.post('/auth/register', { name, email, password })
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('user', JSON.stringify(res.data.user))
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    if (!token.value) return null
    try {
      const res = await api.get('/me')
      user.value = res.data
      localStorage.setItem('user', JSON.stringify(res.data))
      return res.data
    } catch (err) {
      logout()
      throw err
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    user,
    loading,
    isAuthenticated,
    isAdmin,
    isAgent,
    isRequester,
    login,
    register,
    fetchProfile,
    logout
  }
})
