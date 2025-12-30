<template>
  <div class="create-questionnaire">
    <div class="page-header">
      <h1>Create New Questionnaire</h1>
      <p>Design a questionnaire to help your group make decisions</p>
    </div>

    <div class="create-form">
      <form @submit.prevent="handleSubmit" class="questionnaire-form">
        <!-- Basic Information -->
        <div class="form-section">
          <h2>Basic Information</h2>
          
          <div class="form-group">
            <label for="title">Title*</label>
            <input
              id="title"
              v-model="form.title"
              type="text"
              placeholder="Enter questionnaire title"
              :class="{ 'error': errors.title }"
              maxlength="200"
              required
            />
            <span v-if="errors.title" class="error-message">{{ errors.title }}</span>
            <small class="field-hint">Maximum 200 characters</small>
          </div>

          <div class="form-group">
            <label for="description">Description</label>
            <textarea
              id="description"
              v-model="form.description"
              placeholder="Describe what this questionnaire is about (optional)"
              rows="4"
              maxlength="1000"
              :class="{ 'error': errors.description }"
            ></textarea>
            <span v-if="errors.description" class="error-message">{{ errors.description }}</span>
            <small class="field-hint">Maximum 1000 characters</small>
          </div>
        </div>

        <!-- Information about next steps -->
        <div class="form-section info-section">
          <div class="info-card">
            <Icon name="info" />
            <div class="info-content">
              <h3>What happens next?</h3>
              <p>After creating your questionnaire, you'll be able to:</p>
              <ul>
                <li>Add questions with different types (multiple choice, text, ratings)</li>
                <li>Configure questionnaire settings</li>
                <li>Invite members to participate</li>
                <li>View and analyze responses</li>
              </ul>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="form-actions">
          <button 
            type="button" 
            @click="goBack" 
            class="action-btn secondary"
            :disabled="loading"
          >
            <Icon name="arrow-left" />
            Back to Dashboard
          </button>

          <button 
            type="submit" 
            class="action-btn primary"
            :disabled="loading || !isFormValid"
          >
            <Icon v-if="loading" name="loading" class="spin" />
            <Icon v-else name="plus" />
            {{ loading ? 'Creating...' : 'Create Questionnaire' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '../components/Icon.vue'
import { questionnaireAPI } from '../services/api.js'

const router = useRouter()

const loading = ref(false)

const form = reactive({
  title: '',
  description: ''
})

const errors = reactive({
  title: '',
  description: ''
})

const isFormValid = computed(() => {
  return form.title.trim().length > 0 && 
         form.title.length <= 200 &&
         (form.description.length <= 1000)
})

const validateForm = () => {
  errors.title = ''
  errors.description = ''
  
  let isValid = true
  
  if (!form.title.trim()) {
    errors.title = 'Title is required'
    isValid = false
  } else if (form.title.length > 200) {
    errors.title = 'Title must be 200 characters or less'
    isValid = false
  }
  
  if (form.description && form.description.length > 1000) {
    errors.description = 'Description must be 1000 characters or less'
    isValid = false
  }
  
  return isValid
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  loading.value = true
  
  try {
    const requestBody = {
      title: form.title,
      description: form.description || undefined
    }
    
    console.log('Creating questionnaire:', requestBody)
    
    // Real API call to create questionnaire
    const response = await questionnaireAPI.create(requestBody)
    console.log('Questionnaire created successfully:', response.data)
    
    // Navigate back to questionnaires list
    await router.push('/questionnaires')
  } catch (error) {
    console.error('Failed to create questionnaire:', error)
    
    // Show user-friendly error message
    let errorMessage = 'Failed to create questionnaire. Please try again.'
    
    if (error.response?.status === 401) {
      errorMessage = 'Please log in again to continue.'
      await router.push('/login')
    } else if (error.response?.status === 400) {
      errorMessage = 'Invalid questionnaire data. Please check your input.'
    } else if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    }
    
    alert(errorMessage) // TODO: Replace with better error UI
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/questionnaires')
}
</script>

<style scoped>
.create-questionnaire {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 3rem;
}

.page-header h1 {
  font-size: 2.5rem;
  font-weight: bold;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.page-header p {
  font-size: 1.1rem;
  color: #6b7280;
}

.create-form {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.questionnaire-form {
  padding: 2rem;
}

.form-section {
  margin-bottom: 2rem;
}

.form-section:last-of-type {
  margin-bottom: 0;
}

.form-section h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 1.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #e5e7eb;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group input.error,
.form-group textarea.error {
  border-color: #ef4444;
}

.form-group textarea {
  resize: vertical;
  min-height: 100px;
}

.field-hint {
  display: block;
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.error-message {
  display: block;
  font-size: 0.875rem;
  color: #ef4444;
  margin-top: 0.25rem;
}

.info-section {
  background: #f8fafc;
  border-radius: 0.5rem;
  padding: 1.5rem;
}

.info-card {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.info-content h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.info-content p {
  color: #4b5563;
  margin-bottom: 0.75rem;
}

.info-content ul {
  color: #4b5563;
  padding-left: 1.25rem;
  margin: 0;
}

.info-content li {
  margin-bottom: 0.25rem;
}

.form-actions {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding-top: 2rem;
  border-top: 1px solid #e5e7eb;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border-radius: 0.375rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s;
  cursor: pointer;
  border: none;
  font-size: 1rem;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.primary {
  background-color: #3b82f6;
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.action-btn.secondary {
  background-color: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.action-btn.secondary:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>