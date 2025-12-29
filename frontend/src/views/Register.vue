<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="register-header">
        <h1>Create Account</h1>
        <p>Join Radgifa to start making group decisions</p>
      </div>

      <form @submit.prevent="handleRegister" class="register-form">
        <div class="form-row">
          <div class="form-group">
            <label for="name">Full Name</label>
            <input
              id="name"
              v-model="form.name"
              type="text"
              placeholder="Enter your full name"
              :class="{ 'error': errors.name }"
              :disabled="loading"
              @input="onInput"
              required
            />
            <span v-if="errors.name" class="error-message">{{ errors.name }}</span>
          </div>

          <div class="form-group">
            <label for="display_name">Display Name (Optional)</label>
            <input
              id="display_name"
              v-model="form.display_name"
              type="text"
              placeholder="How others will see you"
              :class="{ 'error': errors.display_name }"
              :disabled="loading"
              @input="onInput"
            />
            <span v-if="errors.display_name" class="error-message">{{ errors.display_name }}</span>
          </div>
        </div>

        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            placeholder="Choose a unique username"
            :class="{ 'error': errors.username, 'checking': checkingUsername }"
            :disabled="loading"
            @input="onUsernameInput"
            @blur="checkUsernameAvailability"
            required
          />
          <div class="username-feedback">
            <span v-if="checkingUsername" class="checking-message">Checking availability...</span>
            <span v-else-if="usernameAvailable === true" class="success-message">✓ Username is available</span>
            <span v-else-if="usernameAvailable === false" class="error-message">✗ Username is not available</span>
            <span v-if="errors.username" class="error-message">{{ errors.username }}</span>
          </div>
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <div class="password-input">
            <input
              id="password"
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="Create a strong password"
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
          
          <!-- Validación de contraseña en tiempo real -->
          <div v-if="form.password" class="password-validation">
            <div class="validation-item" :class="{ 'valid': passwordChecks.length }">
              <Icon class="check-icon" :name="passwordChecks.length ? 'check' : 'x'" />
              At least 8 characters
            </div>
            <div class="validation-item" :class="{ 'valid': passwordChecks.hasUpper }">
              <Icon class="check-icon" :name="passwordChecks.hasUpper ? 'check' : 'x'" />
              One uppercase letter
            </div>
            <div class="validation-item" :class="{ 'valid': passwordChecks.hasLower }">
              <Icon class="check-icon" :name="passwordChecks.hasLower ? 'check' : 'x'" />
              One lowercase letter
            </div>
            <div class="validation-item" :class="{ 'valid': passwordChecks.hasNumber }">
              <Icon class="check-icon" :name="passwordChecks.hasNumber ? 'check' : 'x'" />
              One number
            </div>
            <div class="validation-item" :class="{ 'valid': passwordChecks.hasSpecial }">
              <Icon class="check-icon" :name="passwordChecks.hasSpecial ? 'check' : 'x'" />
              One special character
            </div>
          </div>
          
          <div class="password-strength">
            <div class="strength-bar">
              <div 
                class="strength-fill" 
                :class="passwordStrength.class"
                :style="{ width: passwordStrength.width }"
              ></div>
            </div>
            <span class="strength-text">{{ passwordStrength.text }}</span>
          </div>
          <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
        </div>

        <div class="form-group">
          <label for="confirm_password">Confirm Password</label>
          <div class="password-input">
            <input
              id="confirm_password"
              v-model="form.confirm_password"
              :type="showConfirmPassword ? 'text' : 'password'"
              placeholder="Confirm your password"
              :class="{ 'error': errors.confirm_password }"
              :disabled="loading"
              @input="onInput"
              required
            />
            <button
              type="button"
              class="password-toggle"
              @click="toggleConfirmPassword"
              :disabled="loading"
              :title="showConfirmPassword ? 'Hide password' : 'Show password'"
            >
              <Icon :name="showConfirmPassword ? 'eye-off' : 'eye'" />
            </button>
          </div>
          <span v-if="errors.confirm_password" class="error-message">{{ errors.confirm_password }}</span>
        </div>

        <div v-if="errors.general" class="error-message general-error">
          {{ errors.general }}
        </div>

        <button 
          type="submit" 
          class="register-btn"
          :disabled="loading || !isFormValid"
        >
          {{ loading ? 'Creating Account...' : 'Create Account' }}
        </button>
      </form>

      <div class="register-footer">
        <p>
          Already have an account? 
          <router-link to="/login" class="login-link">Sign in</router-link>
        </p>
        <router-link to="/" class="back-home">← Back to Home</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '../services/api'
import { actions } from '../store/auth'
import Icon from '../components/Icon.vue'

const router = useRouter()

// Estado del formulario
const form = reactive({
  name: '',
  display_name: '',
  username: '',
  password: '',
  confirm_password: ''
})

// Estado de errores
const errors = reactive({
  name: '',
  display_name: '',
  username: '',
  password: '',
  confirm_password: '',
  general: ''
})

const loading = ref(false)
const checkingUsername = ref(false)
const usernameAvailable = ref(null)
const showPassword = ref(false)
const showConfirmPassword = ref(false)

// Toggle para mostrar/ocultar contraseñas
const togglePassword = () => {
  showPassword.value = !showPassword.value
}

const toggleConfirmPassword = () => {
  showConfirmPassword.value = !showConfirmPassword.value
}

// Validación específica de contraseña
const passwordChecks = computed(() => {
  const password = form.password
  return {
    length: password.length >= 8,
    hasUpper: /[A-Z]/.test(password),
    hasLower: /[a-z]/.test(password),
    hasNumber: /\d/.test(password),
    hasSpecial: /[^A-Za-z0-9]/.test(password)
  }
})

// Verificar si la contraseña cumple todos los requisitos
const isPasswordValid = computed(() => {
  const checks = passwordChecks.value
  return checks.length && checks.hasUpper && checks.hasLower && checks.hasNumber && checks.hasSpecial
})

// Validación de fortaleza de contraseña
const passwordStrength = computed(() => {
  const password = form.password
  if (!password) return { width: '0%', class: '', text: '' }

  let score = 0
  let requirements = []

  if (password.length >= 8) score += 1
  else requirements.push('8+ characters')

  if (/[A-Z]/.test(password)) score += 1
  else requirements.push('uppercase letter')

  if (/[a-z]/.test(password)) score += 1
  else requirements.push('lowercase letter')

  if (/\d/.test(password)) score += 1
  else requirements.push('number')

  if (/[^A-Za-z0-9]/.test(password)) score += 1
  else requirements.push('special character')

  const strength = {
    0: { width: '20%', class: 'very-weak', text: 'Very Weak' },
    1: { width: '40%', class: 'weak', text: 'Weak' },
    2: { width: '60%', class: 'fair', text: 'Fair' },
    3: { width: '80%', class: 'good', text: 'Good' },
    4: { width: '90%', class: 'strong', text: 'Strong' },
    5: { width: '100%', class: 'very-strong', text: 'Very Strong' }
  }

  return strength[score] || strength[0]
})

// Validar si el formulario es válido
const isFormValid = computed(() => {
  return form.name.trim() && 
         form.username.trim().length >= 3 && 
         isPasswordValid.value &&
         form.password === form.confirm_password &&
         usernameAvailable.value === true
})

// Limpiar errores
const clearErrors = () => {
  Object.keys(errors).forEach(key => errors[key] = '')
}

// Verificar disponibilidad del username
const checkUsernameAvailability = async () => {
  const username = form.username.trim().toLowerCase()
  
  if (username.length < 3) {
    usernameAvailable.value = null
    return
  }

  checkingUsername.value = true
  usernameAvailable.value = null

  try {
    await authAPI.checkUsername(username)
    usernameAvailable.value = true
  } catch (error) {
    if (error.response?.status === 409) {
      usernameAvailable.value = false
    } else {
      usernameAvailable.value = null
    }
  } finally {
    checkingUsername.value = false
  }
}

// Validar formulario completo
const validateForm = () => {
  clearErrors()
  let isValid = true

  // Validar nombre
  if (!form.name.trim()) {
    errors.name = 'Full name is required'
    isValid = false
  } else if (form.name.trim().length < 2) {
    errors.name = 'Name must be at least 2 characters'
    isValid = false
  }

  // Validar username
  if (!form.username.trim()) {
    errors.username = 'Username is required'
    isValid = false
  } else if (form.username.length < 3) {
    errors.username = 'Username must be at least 3 characters'
    isValid = false
  } else if (!/^[a-zA-Z0-9_-]+$/.test(form.username)) {
    errors.username = 'Username can only contain letters, numbers, _ and -'
    isValid = false
  } else if (usernameAvailable.value !== true) {
    errors.username = 'Username is not available'
    isValid = false
  }

  // Validar contraseña
  if (!form.password) {
    errors.password = 'Password is required'
    isValid = false
  } else if (form.password.length < 8) {
    errors.password = 'Password must be at least 8 characters'
    isValid = false
  }

  // Validar confirmación de contraseña
  if (!form.confirm_password) {
    errors.confirm_password = 'Please confirm your password'
    isValid = false
  } else if (form.password !== form.confirm_password) {
    errors.confirm_password = 'Passwords do not match'
    isValid = false
  }

  return isValid
}

// Manejar registro
const handleRegister = async () => {
  if (!validateForm()) return

  loading.value = true
  clearErrors()

  try {
    const userData = {
      name: form.name.trim(),
      display_name: form.display_name.trim() || form.name.trim(),
      username: form.username.toLowerCase().trim(),
      password: form.password
    }

    const response = await authAPI.register(userData)

    // Registro exitoso, ahora hacer login automático
    const loginResponse = await authAPI.login({
      username: userData.username,
      password: userData.password
    })

    const token = loginResponse.data.token
    const user = {
      username: userData.username,
      name: userData.name,
      display_name: userData.display_name
    }

    actions.login(user, token)
    router.push('/dashboard')

  } catch (error) {
    console.error('Registration error:', error)
    
    if (error.response?.status === 409) {
      errors.username = 'Username is already taken'
    } else if (error.response?.data?.error) {
      errors.general = error.response.data.error
    } else {
      errors.general = 'Registration failed. Please try again.'
    }
  } finally {
    loading.value = false
  }
}

// Handlers de input
const onInput = () => {
  if (errors.general) {
    clearErrors()
  }
}

const onUsernameInput = () => {
  usernameAvailable.value = null
  onInput()
}

// Watch para verificar username automáticamente
let usernameTimeout
watch(() => form.username, (newVal) => {
  if (usernameTimeout) clearTimeout(usernameTimeout)
  if (newVal.trim().length >= 3) {
    usernameTimeout = setTimeout(checkUsernameAvailability, 500)
  } else {
    usernameAvailable.value = null
  }
})
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
  max-width: 500px;
}

.register-header {
  text-align: center;
  margin-bottom: 2rem;
}

.register-header h1 {
  font-size: 2rem;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.register-header p {
  color: #6b7280;
  font-size: 0.95rem;
}

.register-form {
  margin-bottom: 2rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
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
}

.form-group input.checking {
  border-color: #f59e0b;
}

.form-group input:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

.username-feedback {
  margin-top: 0.25rem;
  min-height: 1.2rem;
}

.checking-message {
  color: #f59e0b;
  font-size: 0.85rem;
}

.success-message {
  color: #10b981;
  font-size: 0.85rem;
}

.error-message {
  color: #ef4444;
  font-size: 0.85rem;
  display: block;
}

.general-error {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 0.75rem;
  margin-bottom: 1rem;
}

.password-strength {
  margin-top: 0.5rem;
}

.strength-bar {
  width: 100%;
  height: 4px;
  background-color: #e5e7eb;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 0.25rem;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s ease;
  border-radius: 2px;
}

.strength-fill.very-weak { background-color: #ef4444; }
.strength-fill.weak { background-color: #f59e0b; }
.strength-fill.fair { background-color: #eab308; }
.strength-fill.good { background-color: #22c55e; }
.strength-fill.strong { background-color: #10b981; }
.strength-fill.very-strong { background-color: #059669; }

.strength-text {
  font-size: 0.8rem;
  color: #6b7280;
}

.password-validation {
  margin-top: 0.75rem;
  padding: 0.75rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
}

.validation-item {
  display: flex;
  align-items: center;
  margin-bottom: 0.25rem;
  font-size: 0.85rem;
  transition: color 0.2s;
  color: #ef4444;
}

.validation-item:last-child {
  margin-bottom: 0;
}

.validation-item.valid {
  color: #10b981;
}

.validation-item .check-icon {
  margin-right: 0.5rem;
  width: 1rem;
  height: 1rem;
}

.validation-item.valid .check-icon {
  color: #10b981;
}

.validation-item:not(.valid) .check-icon {
  color: #ef4444;
}

.register-btn {
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

.register-btn:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-1px);
}

.register-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  transform: none;
}

.register-footer {
  text-align: center;
}

.register-footer p {
  color: #6b7280;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.login-link {
  color: #4f46e5;
  text-decoration: none;
  font-weight: 500;
}

.login-link:hover {
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