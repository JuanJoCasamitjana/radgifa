<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="login-header">
        <h1>Welcome Back</h1>
        <p>Sign in to your Radgifa account</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            placeholder="Enter your username"
            :class="{ 'error': errors.username }"
            :disabled="loading"
            @input="onInput"
            required
          />
          <span v-if="errors.username" class="error-message">{{ errors.username }}</span>
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <div class="password-input">
            <input
              id="password"
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Enter your password"
              :class="{ 'error': errors.password }"
              :disabled="loading"
              @input="onInput"
              required
            />
            <button
              type="button"
              class="password-toggle"
              @click="togglePassword"
              :disabled="loading"
              :title="showPassword ? 'Hide password' : 'Show password'"
            >
              <Icon :name="showPassword ? 'eye-off' : 'eye'" />
            </button>
          </div>
          
          <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
        </div>

        <div v-if="errors.general" class="error-message general-error">
          {{ errors.general }}
        </div>

        <button 
          type="submit" 
          class="login-btn"
          :disabled="loading"
        >
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <div class="login-footer">
        <p>
          Don't have an account? 
          <router-link to="/register" class="register-link">Sign up</router-link>
        </p>
        <router-link to="/" class="back-home">← Back to Home</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../services/api'
import { actions } from '../store/auth'
import Icon from '../components/Icon.vue'

const router = useRouter()

// Estado del formulario
const form = reactive({
  username: '',
  password: ''
})

// Estado de errores
const errors = reactive({
  username: '',
  password: '',
  general: ''
})

const loading = ref(false)
const showPassword = ref(false)

// Toggle para mostrar/ocultar contraseña
const togglePassword = () => {
  showPassword.value = !showPassword.value
}

// Limpiar errores cuando el usuario escribe
const clearErrors = () => {
  errors.username = ''
  errors.password = ''
  errors.general = ''
}

// Validar formulario
const validateForm = () => {
  clearErrors()
  let isValid = true

  if (!form.username.trim()) {
    errors.username = 'Username is required'
    isValid = false
  } else if (form.username.length < 3) {
    errors.username = 'Username must be at least 3 characters'
    isValid = false
  }

  if (!form.password) {
    errors.password = 'Password is required'
    isValid = false
  }

  return isValid
}

// Manejar login
const handleLogin = async () => {
  if (!validateForm()) return

  loading.value = true
  clearErrors()

  try {
    const response = await authAPI.login({
      username: form.username.toLowerCase().trim(),
      password: form.password
    })

    // Extraer token de la respuesta
    const token = response.data.token

    if (!token) {
      errors.general = 'Invalid response from server'
      return
    }

    // Guardar en store (usuario básico basado en username)
    const user = {
      username: form.username.toLowerCase().trim(),
      // Otros datos del usuario se pueden obtener de otro endpoint si es necesario
    }

    actions.login(user, token)

    // Redirigir al dashboard
    router.push('/dashboard')

  } catch (error) {
    console.error('Login error:', error)
    
    if (error.response?.status === 401) {
      errors.general = 'Invalid username or password'
    } else if (error.response?.data?.error) {
      errors.general = error.response.data.error
    } else {
      errors.general = 'Login failed. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

// Limpiar errores al escribir
const onInput = () => {
  if (errors.general) {
    clearErrors()
  }
}
</script>

<style scoped>
.auth-container {
  min-height: calc(100vh - 4rem); /* Account for navbar */
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.auth-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  padding: 3rem;
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.login-header h1 {
  font-size: 2rem;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.login-header p {
  color: #6b7280;
  font-size: 0.95rem;
}

.login-form {
  margin-bottom: 2rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  color: #374151;
  font-weight: 500;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
}

.form-group input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 2px solid #d1d5db;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.password-input {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input input {
  padding-right: 3rem;
}

.password-toggle {
  position: absolute;
  right: 0.75rem;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
  color: #6b7280;
}

.password-toggle:hover:not(:disabled) {
  color: #4f46e5;
}

.password-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}



.form-group input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.form-group input.error {
  border-color: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}

.form-group input:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

.error-message {
  color: #ef4444;
  font-size: 0.85rem;
  margin-top: 0.25rem;
  display: block;
}

.general-error {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 0.75rem;
  margin-bottom: 1rem;
}

.login-btn {
  width: 100%;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 0.875rem;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s, transform 0.1s;
}

.login-btn:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-1px);
}

.login-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  transform: none;
}

.login-footer {
  text-align: center;
}

.login-footer p {
  color: #6b7280;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.register-link {
  color: #4f46e5;
  text-decoration: none;
  font-weight: 500;
}

.register-link:hover {
  text-decoration: underline;
}

.back-home {
  color: #6b7280;
  text-decoration: none;
  font-size: 0.85rem;
  transition: color 0.2s;
}

.back-home:hover {
  color: #4f46e5;
}
</style>