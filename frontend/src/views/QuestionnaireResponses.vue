<template>
  <div class="questionnaire-responses">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <div class="breadcrumb">
          <button @click="goBack" class="breadcrumb-link">
            <Icon name="arrow-left" />
            Questionnaires
          </button>
          <span class="breadcrumb-separator">/</span>
          <span class="breadcrumb-current">Responses</span>
        </div>
        <h1 v-if="questionnaire">{{ questionnaire.title }}</h1>
        <h1 v-else>Loading...</h1>
        <p v-if="questionnaire">{{ questionnaire.description || 'No description provided' }}</p>
      </div>
      
      <div class="header-actions">
        <button @click="exportResponses" class="action-btn primary" :disabled="responses.length === 0">
          <Icon name="download" />
          Export CSV
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
      Loading responses...
    </div>

    <!-- Statistics -->
    <div v-else-if="questionnaire" class="stats-section">
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon">
            <Icon name="users" />
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ responses.length }}</div>
            <div class="stat-label">Total Responses</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">
            <Icon name="calendar" />
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ formatDate(questionnaire.createdAt) }}</div>
            <div class="stat-label">Created</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">
            <Icon name="check" />
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ questionnaire.is_published ? 'Published' : 'Draft' }}</div>
            <div class="stat-label">Status</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">
            <Icon name="help" />
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ questions.length }}</div>
            <div class="stat-label">Questions</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Responses Content -->
    <div v-if="!loading" class="responses-container">
      <!-- Empty State -->
      <div v-if="responses.length === 0" class="empty-state">
        <Icon name="inbox" />
        <h3>No responses yet</h3>
        <p v-if="!questionnaire?.is_published">
          Publish this questionnaire to start receiving responses.
        </p>
        <p v-else>
          Share your questionnaire to start collecting responses.
        </p>
        <button v-if="!questionnaire?.is_published" @click="publishQuestionnaire" class="create-btn">
          <Icon name="globe" />
          Publish Questionnaire
        </button>
      </div>

      <!-- Responses Table -->
      <div v-else class="responses-table-container">
        <div class="table-header">
          <h3>Responses ({{ responses.length }})</h3>
          <div class="table-filters">
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="Search responses..." 
              class="search-input"
            />
          </div>
        </div>
        
        <div class="table-wrapper">
          <table class="responses-table">
            <thead>
              <tr>
                <th>Respondent</th>
                <th>Submitted</th>
                <th>Completion</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="response in filteredResponses" :key="response.id" class="response-row">
                <td class="respondent-cell">
                  <div class="respondent-info">
                    <div class="respondent-name">{{ response.respondent_name || 'Anonymous' }}</div>
                    <div class="respondent-email">{{ response.respondent_email || 'No email' }}</div>
                  </div>
                </td>
                <td class="date-cell">
                  {{ formatDateTime(response.submitted_at) }}
                </td>
                <td class="completion-cell">
                  <div class="completion-bar">
                    <div class="completion-fill" :style="{ width: response.completion_percentage + '%' }"></div>
                  </div>
                  <span class="completion-text">{{ response.completion_percentage }}%</span>
                </td>
                <td class="actions-cell">
                  <button @click="viewResponse(response)" class="action-btn small">
                    <Icon name="eye" />
                    View
                  </button>
                  <button @click="deleteResponse(response.id)" class="action-btn small danger">
                    <Icon name="trash" />
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Success/Error Messages -->
    <div v-if="successMessage" class="success-toast">
      <Icon name="check" />
      {{ successMessage }}
    </div>
    
    <div v-if="errorMessage" class="error-toast">
      <Icon name="x" />
      {{ errorMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { questionnaireAPI } from '../services/api.js'
import Icon from '../components/Icon.vue'

const router = useRouter()
const route = useRoute()

// Estado
const loading = ref(false)
const questionnaire = ref(null)
const questions = ref([])
const responses = ref([])
const searchQuery = ref('')

// Messages
const successMessage = ref('')
const errorMessage = ref('')

// Get questionnaire ID from route
const questionnaireId = route.params.id

// Computed
const filteredResponses = computed(() => {
  if (!searchQuery.value) return responses.value
  
  const query = searchQuery.value.toLowerCase()
  return responses.value.filter(response => 
    (response.respondent_name || '').toLowerCase().includes(query) ||
    (response.respondent_email || '').toLowerCase().includes(query)
  )
})

// Load data
const loadData = async () => {
  loading.value = true
  try {
    // Load questionnaire details
    const questionnaireResponse = await questionnaireAPI.getDetails(questionnaireId)
    questionnaire.value = questionnaireResponse.data
    
    // Load questions
    const questionsResponse = await questionnaireAPI.getQuestions(questionnaireId)
    questions.value = questionsResponse.data || []
    
    // TODO: Load responses when API is available
    // For now, using mock data
    responses.value = generateMockResponses()
  } catch (error) {
    console.error('Error loading data:', error)
    showError('Failed to load questionnaire data')
  } finally {
    loading.value = false
  }
}

// Mock data generator (remove when real API is available)
const generateMockResponses = () => {
  if (!questionnaire.value?.is_published) return []
  
  const mockResponses = []
  const names = ['John Doe', 'Jane Smith', 'Mike Johnson', 'Sarah Wilson', 'David Brown']
  const emails = ['john@example.com', 'jane@example.com', 'mike@example.com', 'sarah@example.com', 'david@example.com']
  
  for (let i = 0; i < Math.min(names.length, 3); i++) {
    mockResponses.push({
      id: i + 1,
      respondent_name: names[i],
      respondent_email: emails[i],
      submitted_at: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
      completion_percentage: Math.floor(Math.random() * 40) + 60 // 60-100%
    })
  }
  
  return mockResponses
}

// Actions
const refreshData = () => {
  loadData()
}

const goBack = () => {
  router.push('/questionnaires')
}

const publishQuestionnaire = async () => {
  try {
    await questionnaireAPI.publishQuestionnaire(questionnaireId)
    questionnaire.value.is_published = true
    showSuccess('Questionnaire published successfully!')
  } catch (error) {
    console.error('Error publishing questionnaire:', error)
    showError(error.response?.data?.error || 'Failed to publish questionnaire')
  }
}

const exportResponses = () => {
  alert('Export functionality coming soon!')
}

const viewResponse = (response) => {
  alert(`Viewing response from ${response.respondent_name} - functionality coming soon!`)
}

const deleteResponse = async (responseId) => {
  if (confirm('Are you sure you want to delete this response? This action cannot be undone.')) {
    try {
      // TODO: Implement delete response API call
      responses.value = responses.value.filter(r => r.id !== responseId)
      showSuccess('Response deleted successfully!')
    } catch (error) {
      console.error('Error deleting response:', error)
      showError('Failed to delete response')
    }
  }
}

// Utility functions
const formatDate = (date) => {
  if (!date) return 'No date'
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }).format(new Date(date))
}

const formatDateTime = (date) => {
  if (!date) return 'No date'
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(date))
}

// Show messages
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

// Load data on mount
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.questionnaire-responses {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
  min-height: 100vh;
}

/* Header Styles */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 2rem;
}

.header-content {
  flex: 1;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-size: 0.875rem;
  color: #6b7280;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: none;
  border: none;
  color: #4f46e5;
  cursor: pointer;
  text-decoration: none;
}

.breadcrumb-link:hover {
  color: #4338ca;
}

.breadcrumb-separator {
  color: #d1d5db;
}

.breadcrumb-current {
  color: #111827;
  font-weight: 500;
}

.header-content h1 {
  font-size: 2rem;
  color: #111827;
  margin: 0 0 0.5rem 0;
  line-height: 1.2;
}

.header-content p {
  color: #6b7280;
  margin: 0;
  line-height: 1.5;
}

.header-actions {
  display: flex;
  gap: 1rem;
  flex-shrink: 0;
}

/* Stats Section */
.stats-section {
  margin-bottom: 2rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  background: #f3f4f6;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 1.5rem;
  font-weight: 700;
  color: #111827;
  line-height: 1;
}

.stat-label {
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

/* Loading and Empty States */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  color: #6b7280;
  gap: 1rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #6b7280;
}

.empty-state h3 {
  color: #111827;
  margin: 1rem 0 0.5rem 0;
  font-size: 1.25rem;
}

.empty-state p {
  margin-bottom: 1.5rem;
  line-height: 1.5;
}

.create-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.create-btn:hover {
  background: #4338ca;
}

/* Responses Container */
.responses-container {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.table-header h3 {
  margin: 0;
  color: #111827;
}

.search-input {
  padding: 0.5rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  width: 250px;
}

.table-wrapper {
  overflow-x: auto;
}

.responses-table {
  width: 100%;
  border-collapse: collapse;
}

.responses-table th {
  background: #f9fafb;
  padding: 1rem 1.5rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.response-row {
  border-bottom: 1px solid #f3f4f6;
}

.response-row:hover {
  background: #f9fafb;
}

.responses-table td {
  padding: 1rem 1.5rem;
  vertical-align: middle;
}

.respondent-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.respondent-name {
  font-weight: 500;
  color: #111827;
}

.respondent-email {
  font-size: 0.875rem;
  color: #6b7280;
}

.date-cell {
  color: #6b7280;
  font-size: 0.875rem;
}

.completion-cell {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.completion-bar {
  width: 80px;
  height: 8px;
  background: #f3f4f6;
  border-radius: 4px;
  overflow: hidden;
}

.completion-fill {
  height: 100%;
  background: #10b981;
  transition: width 0.3s ease;
}

.completion-text {
  font-size: 0.875rem;
  color: #6b7280;
  min-width: 35px;
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

/* Action Buttons */
.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
}

.action-btn.primary {
  background: #4f46e5;
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background: #4338ca;
}

.action-btn.secondary {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.action-btn.secondary:hover {
  background: #e5e7eb;
}

.action-btn.small {
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
}

.action-btn.danger {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.action-btn.danger:hover {
  background: #fee2e2;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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