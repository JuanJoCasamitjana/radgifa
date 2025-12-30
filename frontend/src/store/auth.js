import { reactive, computed } from 'vue'

const state = reactive({
  user: JSON.parse(localStorage.getItem('user')) || null,
  token: localStorage.getItem('token') || null,
  loading: false
})

export const getters = {
  isAuthenticated: computed(() => !!state.token),
  currentUser: computed(() => state.user),
  isLoading: computed(() => state.loading)
}

export const actions = {
  setUser(user) {
    state.user = user
    localStorage.setItem('user', JSON.stringify(user))
  },

  setToken(token) {
    state.token = token
    localStorage.setItem('token', token)
  },

  login(user, token) {
    actions.setUser(user)
    actions.setToken(token)
  },

  logout() {
    state.user = null
    state.token = null
    localStorage.removeItem('user')
    localStorage.removeItem('token')
  },

  setLoading(loading) {
    state.loading = loading
  },

  initializeAuth() {
    // This method is called to initialize auth state from localStorage
    // The state is already initialized in the reactive declaration above
    // But we can add additional logic here if needed
    const token = localStorage.getItem('token')
    const user = localStorage.getItem('user')
    
    if (token && user) {
      try {
        state.token = token
        state.user = JSON.parse(user)
      } catch (error) {
        console.error('Error parsing stored user data:', error)
        actions.logout()
      }
    }
  }
}

// Store completo
export const store = {
  state,
  getters,
  actions
}