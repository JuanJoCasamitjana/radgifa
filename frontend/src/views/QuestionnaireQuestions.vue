<template>
  <div class="questionnaire-questions">
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

    <div v-if="loading" class="loading-state">
      <Icon name="loading" />
      Loading questions...
    </div>

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

    <div v-if="showEditQuestionForm" class="add-question-form">
      <div class="form-overlay" @click="cancelEditQuestion"></div>
      <div class="form-content">
        <div class="form-header">
          <h3>Edit Question</h3>
          <button @click="cancelEditQuestion" class="close-btn">
            <Icon name="x" />
          </button>
        </div>
        
        <form @submit.prevent="submitEditQuestion" class="question-form">
          <div class="form-group">
            <label for="edit-question-text">Question Text *</label>
            <textarea
              id="edit-question-text"
              v-model="editQuestion.text"
              class="question-input"
              :class="{ error: errors.text }"
              placeholder="Enter your question here..."
              rows="3"
              required
            ></textarea>
            <span v-if="errors.text" class="error-message">{{ errors.text }}</span>
          </div>
          
          <div class="form-group">
            <label for="edit-question-theme">Theme (Optional)</label>
            <input
              id="edit-question-theme"
              v-model="editQuestion.theme"
              type="text"
              class="theme-input"
              :class="{ error: errors.theme }"
              placeholder="e.g., Demographics, Satisfaction, etc."
            />
            <span v-if="errors.theme" class="error-message">{{ errors.theme }}</span>
            <small class="help-text">Themes help organize questions by category</small>
          </div>
          
          <div class="form-actions">
            <button type="button" @click="cancelEditQuestion" class="btn-secondary">
              Cancel
            </button>
            <button type="submit" :disabled="editSubmitting" class="btn-primary">
              <Icon v-if="editSubmitting" name="loading" />
              {{ editSubmitting ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-else-if="!loading" class="questions-container">
      <div v-if="questions.length === 0" class="empty-state">
        <Icon name="help" />
        <h3>No questions yet</h3>
        <p>Start by adding your first question to this questionnaire.</p>
        <button @click="showAddQuestionForm = true" class="create-btn">
          <Icon name="plus" />
          Add Your First Question
        </button>
      </div>


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
              <div class="meta-item">
                <Icon name="users" />
                <span>{{ getTotalAnswers(question) }} responses</span>
              </div>
            </div>
            
            <div v-if="hasAnswers(question)" class="question-results">
              <div class="results-header">
                <Icon name="bar-chart" />
                <span>Results</span>
              </div>
              <div class="results-grid">
                <div class="result-item">
                  <div class="result-label">
                    <Icon name="check" />
                    <span>Yes</span>
                  </div>
                  <div class="result-bar">
                    <div class="result-fill yes" :style="{ width: getPercentage(question, 'Yes') + '%' }"></div>
                  </div>
                  <div class="result-count">{{ getAnswerCount(question, 'Yes') }} ({{ getPercentage(question, 'Yes') }}%)</div>
                </div>
                <div class="result-item">
                  <div class="result-label">
                    <Icon name="x" />
                    <span>No</span>
                  </div>
                  <div class="result-bar">
                    <div class="result-fill no" :style="{ width: getPercentage(question, 'No') + '%' }"></div>
                  </div>
                  <div class="result-count">{{ getAnswerCount(question, 'No') }} ({{ getPercentage(question, 'No') }}%)</div>
                </div>
                <div class="result-item">
                  <div class="result-label">
                    <Icon name="minus" />
                    <span>Pass</span>
                  </div>
                  <div class="result-bar">
                    <div class="result-fill pass" :style="{ width: getPercentage(question, 'Pass') + '%' }"></div>
                  </div>
                  <div class="result-count">{{ getAnswerCount(question, 'Pass') }} ({{ getPercentage(question, 'Pass') }}%)</div>
                </div>
              </div>
            </div>
            
            <div class="question-actions">
              <button 
                @click="editQuestionHandler(question)" 
                :disabled="questionnaire?.is_published"
                class="action-btn small"
              >
                <Icon name="edit" />
                Edit
              </button>
              <button 
                @click="deleteQuestionHandler(question.id)" 
                :disabled="questionnaire?.is_published"
                class="action-btn small danger"
              >
                <Icon name="trash" />
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="successMessage" class="success-toast">
      <Icon name="check" />
      {{ successMessage }}
    </div>

    <div v-if="errorMessage" class="error-toast">
      <Icon name="x" />
      {{ errorMessage }}
    </div>
  </div>

  <ConfirmModal
    :show="confirmModal.show"
    :title="confirmModal.title"
    :message="confirmModal.message"
    :confirm-text="confirmModal.confirmText"
    :type="confirmModal.type"
    :loading="confirmModal.loading"
    @confirm="confirmModal.onConfirm"
    @cancel="closeConfirmModal"
  />
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { questionnaireAPI } from '../services/api.js'
import Icon from '../components/Icon.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()


const loading = ref(false)
const submitting = ref(false)
const editSubmitting = ref(false)
const questionnaire = ref(null)
const questions = ref([])
const showAddQuestionForm = ref(false)
const showEditQuestionForm = ref(false)
const editingQuestion = ref(null)
const successMessage = ref('')
const errorMessage = ref('')

const confirmModal = reactive({
  show: false,
  title: '',
  message: '',
  confirmText: 'Confirm',
  type: 'danger',
  loading: false,
  onConfirm: () => {}
})


const newQuestion = reactive({
  text: '',
  theme: ''
})

const editQuestion = reactive({
  text: '',
  theme: ''
})

const errors = reactive({
  text: '',
  theme: ''
})


const questionnaireId = route.params.id


const loadData = async () => {
  loading.value = true
  try {
    
    const questionnaireResponse = await questionnaireAPI.getDetails(questionnaireId)
    questionnaire.value = questionnaireResponse.data
    
    
    const questionsResponse = await questionnaireAPI.getQuestions(questionnaireId)
    const rawQuestions = questionsResponse.data || []
    
    questions.value = rawQuestions.map(q => ({
      ...q,
      createdAt: q.created_at || q.createdAt || q.date_created || q.created || Date.now()
    }))
  } catch (error) {
    console.error('Error loading data:', error)
    showError('Failed to load questionnaire data')
  } finally {
    loading.value = false
  }
}


const refreshData = () => {
  loadData()
}


const goBack = () => {
  router.push('/questionnaires')
}


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
    
    
    newQuestion.text = ''
    newQuestion.theme = ''
    showAddQuestionForm.value = false
    
    
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


const cancelAddQuestion = () => {
  newQuestion.text = ''
  newQuestion.theme = ''
  errors.text = ''
  errors.theme = ''
  showAddQuestionForm.value = false
}


const editQuestionHandler = (question) => {
  if (questionnaire.value?.is_published) {
    console.warn('Cannot edit questions in published questionnaires')
    return
  }
  
  errors.text = ''
  errors.theme = ''
  
  editingQuestion.value = question
  editQuestion.text = question.text || ''
  editQuestion.theme = question.theme || ''
  showEditQuestionForm.value = true
}


const cancelEditQuestion = () => {
  showEditQuestionForm.value = false
  editingQuestion.value = null
  editQuestion.text = ''
  editQuestion.theme = ''
  
  errors.text = ''
  errors.theme = ''
}


const submitEditQuestion = async () => {
  errors.text = ''
  errors.theme = ''
  
  if (!editQuestion.text.trim()) {
    errors.text = 'Question text is required'
    return
  }
  
  if (editQuestion.text.trim().length < 5) {
    errors.text = 'Question text must be at least 5 characters'
    return
  }
  
  if (editQuestion.theme && editQuestion.theme.trim().length > 255) {
    errors.theme = 'Theme cannot exceed 255 characters'
    return
  }
  
  try {
    editSubmitting.value = true
    
    const updateData = {
      text: editQuestion.text.trim(),
      theme: editQuestion.theme.trim() || null
    }
    
    await questionnaireAPI.updateQuestion(questionnaireId, editingQuestion.value.id, updateData)
    
    
    const questionIndex = questions.value.findIndex(q => q.id === editingQuestion.value.id)
    if (questionIndex !== -1) {
      questions.value[questionIndex] = {
        ...questions.value[questionIndex],
        ...updateData
      }
    }
    
    cancelEditQuestion()
    showSuccess('Question updated successfully!')
  } catch (error) {
    console.error('Error updating question:', error)
    showError(error.response?.data?.error || 'Failed to update question')
  } finally {
    editSubmitting.value = false
  }
}


const deleteQuestionHandler = async (questionId) => {
  if (questionnaire.value?.is_published) {
    console.warn('Cannot delete questions from published questionnaires')
    return
  }
  
  const question = questions.value.find(q => q.id === questionId)
  if (!question) return
  
  confirmModal.show = true
  confirmModal.title = 'Delete Question'
  confirmModal.message = `Are you sure you want to delete the question "${question.text}"? This action cannot be undone.`
  confirmModal.confirmText = 'Delete'
  confirmModal.type = 'danger'
  confirmModal.onConfirm = () => executeDeleteQuestion(questionnaireId, questionId)
}

const executeDeleteQuestion = async (questionnaireId, questionId) => {
  try {
    confirmModal.loading = true
    await questionnaireAPI.deleteQuestion(questionnaireId, questionId)
    
    questions.value = questions.value.filter(q => q.id !== questionId)
    
    showSuccess('Question deleted successfully!')
    closeConfirmModal()
  } catch (error) {
    console.error('Error deleting question:', error)
    showError(error.response?.data?.error || 'Failed to delete question')
  } finally {
    confirmModal.loading = false
  }
}


const formatDate = (date) => {
  if (!date) return 'No date'
  
  let dateObj
  if (typeof date === 'number') {
    dateObj = new Date(date)
  } else if (typeof date === 'string') {
    dateObj = new Date(date)
  } else {
    return 'Invalid date'
  }
  
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


const showSuccess = (message) => {
  successMessage.value = message
  setTimeout(() => {
    successMessage.value = ''
  }, 3000)
}


const showError = (message) => {
  errorMessage.value = message
  setTimeout(() => {
    errorMessage.value = ''
  }, 5000)
}

const closeConfirmModal = () => {
  confirmModal.show = false
  confirmModal.loading = false
}

const hasAnswers = (question) => {
  return question.edges?.answers && question.edges.answers.length > 0
}

const getTotalAnswers = (question) => {
  if (!question.edges?.answers) return 0
  return question.edges.answers.length
}

const getAnswerCount = (question, value) => {
  if (!question.edges?.answers) return 0
  return question.edges.answers.filter(a => a.answer_value === value).length
}

const getPercentage = (question, value) => {
  const total = getTotalAnswers(question)
  if (total === 0) return 0
  const count = getAnswerCount(question, value)
  return Math.round((count / total) * 100)
}

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

.question-results {
  margin-bottom: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.results-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.75rem;
  font-size: 0.9rem;
}

.results-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.result-item {
  display: grid;
  grid-template-columns: 80px 1fr 120px;
  align-items: center;
  gap: 0.75rem;
}

.result-label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.9rem;
  font-weight: 500;
  color: #374151;
}

.result-bar {
  height: 24px;
  background: #e5e7eb;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.result-fill {
  height: 100%;
  transition: width 0.3s ease;
  border-radius: inherit;
}

.result-fill.yes {
  background: linear-gradient(90deg, #10b981, #059669);
}

.result-fill.no {
  background: linear-gradient(90deg, #ef4444, #dc2626);
}

.result-fill.pass {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.result-count {
  font-size: 0.85rem;
  font-weight: 600;
  color: #374151;
  text-align: right;
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

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
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