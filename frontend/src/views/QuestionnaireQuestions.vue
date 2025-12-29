<template>
  <div class="questionnaire-questions">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="breadcrumb">
          <button @click="goBack" class="breadcrumb-link">
            <Icon name="arrow-left" />
            Questionnaires
          </button>
          <span class="breadcrumb-separator">/</span>
          <span class="breadcrumb-current">Questions</span>
        </div>
        <h1 v-if="questionnaire">{{ questionnaire.title }}</h1>
        <h1 v-else>Loading...</h1>
        <p v-if="questionnaire">{{ questionnaire.description || 'No description provided' }}</p>
      </div>
      
      <div class="header-actions">
        <button @click="showAddQuestionForm = true" class="action-btn primary">
          <Icon name="plus" />
          Add Question
        </button>
        
        <button @click="refreshData" class="action-btn secondary">
          <Icon name="refresh" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <Icon name="loading" />
      Loading questions...
    </div>

    <!-- Add Question Form -->
    <div v-if="showAddQuestionForm" class="add-question-form">
      <div class="form-overlay" @click="cancelAddQuestion"></div>
      <div class="form-content">
        <div class="form-header">
          <h3>Add New Question</h3>
          <button @click="cancelAddQuestion" class="close-btn">
            <Icon name="x" />
          </button>
        </div>
        
        <form @submit.prevent="submitQuestion" class="question-form">
          <div class="form-group">
            <label for="questionText">Question Text *</label>
            <textarea
              id="questionText"
              v-model="newQuestion.text"
              placeholder="Enter your question here..."
              rows="3"
              required
              :class="{ 'error': errors.text }"
            ></textarea>
            <div v-if="errors.text" class="error-message">{{ errors.text }}</div>
            <div class="field-info">
              {{ newQuestion.text.length }} characters
            </div>
          </div>
          
          <div class="form-group">
            <label for="questionTheme">Theme (Optional)</label>
            <input
              id="questionTheme"
              v-model="newQuestion.theme"
              type="text"
              placeholder="e.g., Food Preferences, Technology, etc."
              maxlength="255"
              :class="{ 'error': errors.theme }"
            />
            <div v-if="errors.theme" class="error-message">{{ errors.theme }}</div>
            <div class="field-info">
              {{ newQuestion.theme.length }}/255 characters
            </div>
          </div>
          
          <div class="form-actions">
            <button type="button" @click="cancelAddQuestion" class="btn-secondary">
              Cancel
            </button>
            <button type="submit" :disabled="submitting" class="btn-primary">
              <Icon v-if="submitting" name="loading" />
              {{ submitting ? 'Adding...' : 'Add Question' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Questions List -->
    <div v-else-if="!loading" class="questions-container">
      <!-- Empty State -->
      <div v-if="questions.length === 0" class="empty-state">
        <Icon name="help" />
        <h3>No questions yet</h3>
        <p>Start by adding your first question to this questionnaire.</p>
        <button @click="showAddQuestionForm = true" class="create-btn">
          <Icon name="plus" />
          Add Your First Question
        </button>
      </div>

      <!-- Questions List -->
      <div v-else class="questions-list">
        <div class="list-header">
          <h3>Questions ({{ questions.length }})</h3>
          <div class="list-info">
            <span>Manage the questions for this questionnaire</span>
          </div>
        </div>
        
        <div class="questions-grid">
          <div
            v-for="(question, index) in questions"
            :key="question.id"
            class="question-card"
          >
            <div class="question-header">
              <div class="question-number">Q{{ index + 1 }}</div>
              <div v-if="question.theme" class="question-theme">
                <Icon name="tag" />
                {{ question.theme }}
              </div>
            </div>
            
            <div class="question-content">
              <p class="question-text">{{ question.text }}</p>
            </div>
            
            <div class="question-meta">
              <div class="meta-item">
                <Icon name="calendar" />
                <span>{{ formatDate(question.createdAt) }}</span>
              </div>
            </div>
            
            <div class="question-actions">
              <button @click="editQuestion(question)" class="action-btn small">
                <Icon name="edit" />
                Edit
              </button>
              <button @click="deleteQuestion(question.id)" class="action-btn small danger">
                <Icon name="trash" />
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="success-toast">
      <Icon name="check" />
      {{ successMessage }}
    </div>

    <!-- Error Message -->
    <div v-if="errorMessage" class="error-toast">
      <Icon name="x" />
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { questionnaireAPI } from '../services/api.js'
import Icon from '../components/Icon.vue'

const route = useRoute()
const router = useRouter()

// State
const loading = ref(false)
const submitting = ref(false)
const questionnaire = ref(null)
const questions = ref([])
const showAddQuestionForm = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

// Form data
const newQuestion = reactive({
  text: '',
  theme: ''
})

const errors = reactive({
  text: '',
  theme: ''
})

// Get questionnaire ID from route
const questionnaireId = route.params.id

// Load questionnaire details and questions
const loadData = async () => {
  loading.value = true
  try {
    // Load questionnaire details
    const questionnaireResponse = await questionnaireAPI.getDetails(questionnaireId)
    questionnaire.value = questionnaireResponse.data
    
    // Load questions
    const questionsResponse = await questionnaireAPI.getQuestions(questionnaireId)
    questions.value = questionsResponse.data || []
  } catch (error) {
    console.error('Error loading data:', error)
    showError('Failed to load questionnaire data')
  } finally {
    loading.value = false
  }
}

// Refresh data
const refreshData = () => {
  loadData()
}

// Go back to questionnaires list
const goBack = () => {
  router.push('/questionnaires')
}

// Validate question form
const validateQuestion = () => {
  errors.text = ''
  errors.theme = ''
  
  if (!newQuestion.text.trim()) {
    errors.text = 'Question text is required'
    return false
  }
  
  if (newQuestion.text.length < 1) {
    errors.text = 'Question text must be at least 1 character'
    return false
  }
  
  if (newQuestion.theme.length > 255) {
    errors.theme = 'Theme must be 255 characters or less'
    return false
  }
  
  return true
}

// Submit new question
const submitQuestion = async () => {
  if (!validateQuestion()) {
    return
  }
  
  submitting.value = true
  try {
    const questionData = {
      text: newQuestion.text.trim(),
      theme: newQuestion.theme.trim() || undefined
    }
    
    await questionnaireAPI.createQuestion(questionnaireId, questionData)
    
    // Reset form
    newQuestion.text = ''
    newQuestion.theme = ''
    showAddQuestionForm.value = false
    
    // Reload questions
    await loadData()
    
    showSuccess('Question added successfully!')
  } catch (error) {
    console.error('Error creating question:', error)
    if (error.response?.data?.error) {
      showError(error.response.data.error)
    } else {
      showError('Failed to create question')
    }
  } finally {
    submitting.value = false
  }
}

// Cancel adding question
const cancelAddQuestion = () => {
  newQuestion.text = ''
  newQuestion.theme = ''
  errors.text = ''
  errors.theme = ''
  showAddQuestionForm.value = false
}

// Edit question (placeholder)
const editQuestion = (question) => {
  alert(`Editing question "${question.text}" - functionality coming soon`)
}

// Delete question (placeholder)
const deleteQuestion = (questionId) => {
  if (confirm('Are you sure you want to delete this question? This action cannot be undone.')) {
    alert(`Deleting question ${questionId} - functionality coming soon`)
  }
}

// Format date
const formatDate = (date) => {
  if (!date) return 'No date'
  
  const dateObj = new Date(date)
  if (isNaN(dateObj.getTime())) {
    return 'Invalid date'
  }
  
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(dateObj)
}

// Show success message
const showSuccess = (message) => {
  successMessage.value = message
  setTimeout(() => {
    successMessage.value = ''
  }, 3000)
}

// Show error message
const showError = (message) => {
  errorMessage.value = message
  setTimeout(() => {
    errorMessage.value = ''
  }, 5000)
}

// Load data on mount
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.questionnaire-questions {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

.header-content {
  flex: 1;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 0.875rem;
  padding: 0.25rem 0;
}

.breadcrumb-link:hover {
  color: #374151;
}

.breadcrumb-separator {
  color: #9ca3af;
}

.breadcrumb-current {
  color: #111827;
  font-weight: 500;
}

.header-content h1 {
  font-size: 2.5rem;
  color: #111827;
  margin: 0 0 0.5rem 0;
}

.header-content p {
  color: #6b7280;
  margin: 0;
  font-size: 1.125rem;
}

.header-actions {
  display: flex;
  gap: 1rem;
  margin-left: 2rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 500;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn.primary {
  background: #3b82f6;
  color: white;
  border: none;
}

.action-btn.primary:hover {
  background: #2563eb;
}

.action-btn.secondary {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.action-btn.secondary:hover {
  background: #e5e7eb;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 4rem;
  color: #6b7280;
  font-size: 1.125rem;
}

/* Add Question Form */
.add-question-form {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.form-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
}

.form-content {
  position: relative;
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow: auto;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 1.5rem 0;
  border-bottom: 1px solid #e5e7eb;
  margin-bottom: 1.5rem;
}

.form-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #111827;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.25rem;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.question-form {
  padding: 0 1.5rem 1.5rem;
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
  border-radius: 0.5rem;
  font-size: 0.875rem;
  transition: border-color 0.2s;
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

.error-message {
  color: #ef4444;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.field-info {
  color: #6b7280;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.btn-secondary {
  padding: 0.75rem 1.5rem;
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e5e7eb;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Questions Container */
.questions-container {
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #6b7280;
}

.empty-state h3 {
  font-size: 1.5rem;
  color: #374151;
  margin: 1rem 0 0.5rem;
}

.empty-state p {
  margin: 0 0 2rem;
}

.create-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.create-btn:hover {
  background: #2563eb;
}

.list-header {
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.list-header h3 {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  color: #111827;
}

.list-info {
  color: #6b7280;
}

.questions-grid {
  display: grid;
  gap: 1rem;
  padding: 1.5rem;
}

.question-card {
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.2s;
}

.question-card:hover {
  border-color: #d1d5db;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.question-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.question-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  background: #3b82f6;
  color: white;
  border-radius: 50%;
  font-size: 0.75rem;
  font-weight: 600;
}

.question-theme {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: #f3f4f6;
  color: #374151;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.question-content {
  margin-bottom: 1rem;
}

.question-text {
  margin: 0;
  color: #111827;
  font-size: 1rem;
  line-height: 1.5;
}

.question-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.question-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn.small {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem 0.75rem;
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn.small:hover {
  background: #e5e7eb;
}

.action-btn.small.danger {
  background: #fef2f2;
  color: #dc2626;
  border-color: #fecaca;
}

.action-btn.small.danger:hover {
  background: #fee2e2;
}

/* Toast Messages */
.success-toast,
.error-toast {
  position: fixed;
  top: 2rem;
  right: 2rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 500;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  z-index: 60;
}

.success-toast {
  background: #f0fdf4;
  color: #166534;
  border: 1px solid #bbf7d0;
}

.error-toast {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}
</style>