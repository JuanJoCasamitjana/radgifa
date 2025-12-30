<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="join-header">
        <h1>Únete al cuestionario</h1>
        <p v-if="questionnaireTitle">{{ questionnaireTitle }}</p>
        <p v-else>Completa los datos para participar</p>
      </div>

      <div v-if="loading" class="loading-container">
        <Icon name="loading" />
        <p>Verificando invitación...</p>
      </div>

      <div v-else-if="tokenError" class="error-container">
        <Icon name="alert-triangle" />
        <h3>Invitación inválida</h3>
        <p>{{ tokenError }}</p>
      </div>

      <form v-else @submit.prevent="handleSubmit" class="join-form">
        <div class="action-toggle">
          <button
            type="button"
            @click="joinAction = 'register'"
            :class="{ 'active': joinAction === 'register' }"
            class="toggle-btn left"
          >
            Nuevo participante
          </button>
          <button
            type="button"
            @click="joinAction = 'login'"
            :class="{ 'active': joinAction === 'login' }"
            class="toggle-btn right"
          >
            Ya tengo cuenta
          </button>
        </div>

        <div class="form-group">
          <label for="unique_identifier">Identificador único</label>
          <div class="input-with-icon">
            <input
              id="unique_identifier"
              v-model="formData.unique_identifier"
              type="text"
              placeholder="ej: participante123"
              :disabled="submitting"
              required
              minlength="3"
              maxlength="32"
              pattern="^[a-zA-Z0-9_-]+$"
              @input="checkIdentifierAvailability"
            />
            <div v-if="identifierStatus.checking" class="input-icon">
              <Icon name="loading" />
            </div>
            <div v-else-if="identifierStatus.available === false" class="input-icon error">
              <Icon name="x" />
            </div>
            <div v-else-if="identifierStatus.available === true" class="input-icon success">
              <Icon name="check" />
            </div>
          </div>
          <span v-if="identifierStatus.available === false" class="error-message">
            Este identificador ya está en uso
          </span>
          <span v-else class="hint-message">
            Solo letras, números, guiones y guiones bajos
          </span>
        </div>

        <div v-if="joinAction === 'register'" class="form-group">
          <label for="display_name">Nombre a mostrar (opcional)</label>
          <input
            id="display_name"
            v-model="formData.display_name"
            type="text"
            placeholder="ej: Juan Pérez"
            :disabled="submitting"
            maxlength="100"
          />
        </div>

        <div v-if="joinAction === 'login'" class="form-group">
          <label for="passcode">Código de acceso</label>
          <input
            id="passcode"
            v-model="formData.passcode"
            type="password"
            placeholder="Ingresa tu código de 8 caracteres"
            :disabled="submitting"
            required
            maxlength="8"
          />
          <span class="hint-message">
            Código de 8 caracteres que recibiste al unirte anteriormente
          </span>
        </div>

        <div v-if="error" class="error-banner">
          <Icon name="alert-triangle" />
          <div>
            <strong>Error</strong>
            <p>{{ error }}</p>
          </div>
        </div>

        <button
          type="submit"
          class="submit-btn"
          :disabled="submitting || (joinAction === 'register' && identifierStatus.available === false)"
        >
          <Icon v-if="submitting" name="loading" />
          {{ joinAction === 'register' ? 'Unirme al cuestionario' : 'Acceder' }}
        </button>
      </form>
    </div>

    <div v-if="showPasscodeModal" class="modal-overlay" @click="closeModalAndRedirect">
      <div class="passcode-modal" @click.stop>
        <div class="modal-header">
          <div class="icon-container success">
            <Icon name="check" />
          </div>
          <h3>¡Te has unido exitosamente!</h3>
        </div>
        
        <div class="modal-body">
          <p>Guarda este código de acceso para volver a entrar:</p>
          <div class="passcode-display">
            <code>{{ savedPasscode }}</code>
            <button @click="copyPasscode" class="copy-btn" :class="{ 'copied': passcodeCopied }">
              <Icon :name="passcodeCopied ? 'check' : 'copy'" />
              {{ passcodeCopied ? 'Copiado' : 'Copiar' }}
            </button>
          </div>
          <p class="warning-text">
            <Icon name="alert-triangle" />
            Importante: Lo necesitarás para acceder nuevamente a tus respuestas.
          </p>
        </div>
        
        <div class="modal-footer">
          <button @click="closeModalAndRedirect" class="confirm-btn">
            Continuar al cuestionario
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { participationAPI } from '../services/api'
import { getters } from '../store/auth'
import Icon from '../components/Icon.vue'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const submitting = ref(false)
const tokenError = ref(null)
const error = ref(null)
const questionnaireTitle = ref('')
const joinAction = ref('register')
const showPasscodeModal = ref(false)
const savedPasscode = ref('')
const passcodeCopied = ref(false)

const formData = reactive({
  unique_identifier: '',
  display_name: '',
  passcode: ''
})

const identifierStatus = reactive({
  checking: false,
  available: null
})

const token = computed(() => route.params.token)

let identifierTimeout = null

const checkIdentifierAvailability = async () => {
  if (!formData.unique_identifier || formData.unique_identifier.length < 3 || joinAction.value === 'login') {
    identifierStatus.available = null
    return
  }

  clearTimeout(identifierTimeout)
  identifierTimeout = setTimeout(async () => {
    identifierStatus.checking = true
    try {
      const response = await participationAPI.checkMemberIdentifier(token.value, formData.unique_identifier)
      identifierStatus.available = response.data.available
    } catch (err) {
      identifierStatus.available = null
    } finally {
      identifierStatus.checking = false
    }
  }, 500)
}

const handleSubmit = async () => {
  error.value = null
  submitting.value = true

  try {
    const memberData = {
      action: joinAction.value,
      unique_identifier: formData.unique_identifier.toLowerCase().trim(),
      display_name: formData.display_name.trim() || null,
      passcode: formData.passcode || null
    }

    const response = await participationAPI.joinQuestionnaire(token.value, memberData)

    // Guardar token JWT en localStorage
    if (response.data.token) {
      localStorage.setItem('member_token', response.data.token)
      localStorage.setItem('member_id', response.data.member_id)
      localStorage.setItem('member_type', joinAction.value === 'login' ? 'returning' : 'anonymous')
    }

    if (joinAction.value === 'register' && response.data.passcode) {
      savedPasscode.value = response.data.passcode
      showPasscodeModal.value = true
      submitting.value = false
    } else {
      router.push(`/answer/${token.value}`)
    }

  } catch (err) {
    if (err.response?.status === 409) {
      error.value = 'Este identificador ya está en uso. Prueba con otro diferente.'
    } else if (err.response?.status === 401 && joinAction.value === 'login') {
      error.value = 'Código de acceso incorrecto. Verifica e intenta nuevamente.'
    } else if (err.response?.status === 400) {
      error.value = 'La invitación ha expirado o es inválida.'
    } else {
      error.value = 'Error al unirse al cuestionario. Intenta nuevamente.'
    }
  } finally {
    submitting.value = false
  }
}

const validateToken = async () => {
  if (!token.value) {
    tokenError.value = 'No se proporcionó un token de invitación válido.'
    loading.value = false
    return
  }

  try {
    await participationAPI.checkMemberIdentifier(token.value, 'test-validation')
    
    if (getters.isAuthenticated.value) {
      await tryAutoJoin()
    } else {
      loading.value = false
    }
  } catch (err) {
    if (err.response?.status === 400 || err.response?.status === 401) {
      tokenError.value = 'La invitación ha expirado o es inválida.'
    } else {
      tokenError.value = 'Error al verificar la invitación.'
    }
    loading.value = false
  }
}

const tryAutoJoin = async () => {
  try {
    const response = await participationAPI.joinQuestionnaire(token.value, {
      action: 'register',
      unique_identifier: `user_${Date.now()}`,
      display_name: getters.currentUser.value?.display_name || getters.currentUser.value?.name || null,
      passcode: null
    })

    if (response.data.already_member) {
      localStorage.setItem('member_id', response.data.member_id)
      localStorage.setItem('member_type', 'authenticated')
      router.push(`/answer/${token.value}`)
    } else if (response.data.member_id) {
      localStorage.setItem('member_id', response.data.member_id)
      localStorage.setItem('member_type', 'authenticated')
      router.push(`/answer/${token.value}`)
    }
  } catch (err) {
    loading.value = false
  }
}

const copyPasscode = async () => {
  try {
    await navigator.clipboard.writeText(savedPasscode.value)
    passcodeCopied.value = true
    setTimeout(() => {
      passcodeCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Error copying to clipboard:', err)
  }
}

const closeModalAndRedirect = () => {
  showPasscodeModal.value = false
  router.push(`/answer/${token.value}`)
}

onMounted(() => {
  validateToken()
})
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 2rem 1rem;
}

.auth-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 2.5rem;
  width: 100%;
  max-width: 480px;
  animation: slideUp 0.4s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.join-header {
  text-align: center;
  margin-bottom: 2rem;
}

.join-header h1 {
  font-size: 1.875rem;
  font-weight: 700;
  color: #111827;
  margin-bottom: 0.5rem;
}

.join-header p {
  color: #6b7280;
  font-size: 0.95rem;
}

.loading-container {
  text-align: center;
  padding: 3rem 1rem;
}

.loading-container p {
  margin-top: 1rem;
  color: #6b7280;
  font-size: 0.95rem;
}

.error-container {
  text-align: center;
  padding: 2rem 1rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
}

.error-container h3 {
  color: #991b1b;
  font-size: 1.125rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem;
}

.error-container p {
  color: #dc2626;
  font-size: 0.875rem;
}

.join-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.action-toggle {
  display: flex;
  gap: 0;
  background: #f3f4f6;
  border-radius: 8px;
  padding: 0.25rem;
}

.toggle-btn {
  flex: 1;
  padding: 0.75rem 1rem;
  border: none;
  background: transparent;
  color: #6b7280;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn.left {
  border-radius: 6px 0 0 6px;
}

.toggle-btn.right {
  border-radius: 0 6px 6px 0;
}

.toggle-btn.active {
  background: #4f46e5;
  color: white;
  box-shadow: 0 2px 4px rgba(79, 70, 229, 0.2);
}

.toggle-btn:hover:not(.active) {
  background: #e5e7eb;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #374151;
}

.form-group input {
  padding: 0.875rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
  width: 100%;
}

.form-group input:focus {
  outline: none;
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.form-group input:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

.input-with-icon {
  position: relative;
}

.input-icon {
  position: absolute;
  right: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.input-icon.error {
  color: #ef4444;
}

.input-icon.success {
  color: #10b981;
}

.error-message {
  color: #ef4444;
  font-size: 0.85rem;
}

.hint-message {
  color: #6b7280;
  font-size: 0.8rem;
}

.error-banner {
  display: flex;
  gap: 0.75rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  padding: 0.875rem;
}

.error-banner svg {
  flex-shrink: 0;
  color: #ef4444;
}

.error-banner strong {
  display: block;
  color: #991b1b;
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
}

.error-banner p {
  color: #dc2626;
  font-size: 0.85rem;
  margin: 0;
}

.submit-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 0.875rem;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.submit-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
  transform: none;
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.passcode-modal {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 2rem;
  width: 90%;
  max-width: 500px;
  animation: slideUp 0.3s ease-out;
}

.passcode-modal .modal-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.passcode-modal .icon-container {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
}

.passcode-modal .icon-container.success {
  background: #dcfce7;
  color: #16a34a;
}

.passcode-modal .icon-container :deep(svg) {
  width: 32px;
  height: 32px;
}

.passcode-modal .modal-header h3 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  margin: 0;
}

.passcode-modal .modal-body {
  margin-bottom: 1.5rem;
}

.passcode-modal .modal-body > p {
  color: #6b7280;
  margin-bottom: 1rem;
  text-align: center;
}

.passcode-display {
  background: #f9fafb;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.passcode-display code {
  font-family: 'Courier New', monospace;
  font-size: 1.25rem;
  font-weight: 700;
  color: #4f46e5;
  letter-spacing: 0.1em;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.copy-btn:hover {
  background: #4338ca;
}

.copy-btn.copied {
  background: #16a34a;
}

.copy-btn :deep(svg) {
  width: 16px;
  height: 16px;
}

.warning-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 6px;
  padding: 0.75rem;
  color: #92400e;
  font-size: 0.875rem;
}

.warning-text :deep(svg) {
  width: 18px;
  height: 18px;
  color: #f59e0b;
  flex-shrink: 0;
}

.passcode-modal .modal-footer {
  display: flex;
  justify-content: center;
}

.confirm-btn {
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 0.875rem 2rem;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.confirm-btn:hover {
  background: #4338ca;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

@media (max-width: 640px) {
  .auth-card {
    padding: 1.5rem;
  }

  .join-header h1 {
    font-size: 1.5rem;
  }

  .passcode-modal {
    padding: 1.5rem;
    width: 95%;
  }

  .passcode-display {
    flex-direction: column;
    text-align: center;
  }

  .copy-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
