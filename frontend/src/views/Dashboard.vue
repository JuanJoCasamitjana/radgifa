<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Dashboard</h1>
      <p>Welcome back, {{ currentUser?.username }}!</p>
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <Icon name="plus" />
        </div>
        <div class="stat-info">
          <h3>{{ stats.totalQuestionnaires }}</h3>
          <p>Total Questionnaires</p>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon responses">
          <Icon name="user" />
        </div>
        <div class="stat-info">
          <h3>{{ stats.totalResponses }}</h3>
          <p>Total Responses</p>
        </div>
      </div>
    </div>

    <div class="dashboard-actions">
      <button @click="createQuestionnaire" class="action-btn primary">
        <Icon name="plus" />
        Create New Questionnaire
      </button>
      
      <button @click="viewAll" class="action-btn secondary">
        <Icon name="list" />
        View All Questionnaires
      </button>
    </div>

    <div class="recent-section">
      <div class="section-header">
        <h2>Recent Questionnaires</h2>
        <button @click="refreshData" class="refresh-btn">
          <Icon name="refresh" />
          Refresh
        </button>
      </div>

      <div v-if="loading" class="loading-state">
        <Icon name="loading" />
        Loading questionnaires...
      </div>

      <div v-else-if="questionnaires.length === 0" class="empty-state">
        <Icon name="plus" />
        <h3>No questionnaires yet</h3>
        <p>Create your first questionnaire to get started!</p>
        <button @click="createQuestionnaire" class="create-btn">
          Create Questionnaire
        </button>
      </div>

      <div v-else class="questionnaires-grid">
        <div 
          v-for="questionnaire in questionnaires" 
          :key="questionnaire.id"
          class="questionnaire-card"
          @click="openQuestionnaire(questionnaire.id)"
        >
          <div class="card-header">
            <h3>{{ questionnaire.title }}</h3>
            <span class="status-badge" :class="questionnaire.status">
              {{ questionnaire.status }}
            </span>
          </div>
          
          <p class="card-description">{{ questionnaire.description }}</p>
          
          <div class="card-meta">
            <span class="meta-item">
              <Icon name="user" />
              {{ questionnaire.responseCount }} responses
            </span>
            <span class="meta-item">
              <Icon name="calendar" />
              {{ formatDate(questionnaire.createdAt) }}
            </span>
          </div>
          
          <div class="card-actions">
            <button @click.stop="editQuestionnaire(questionnaire.id)" class="action-link">
              Edit
            </button>
            <button @click.stop="shareQuestionnaire(questionnaire.id)" class="action-link">
              Share
            </button>
            <button @click.stop="deleteQuestionnaire(questionnaire.id)" class="action-link danger">
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div v-if="showDeleteModal" class="modal-overlay" @click="closeDeleteModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>Delete Questionnaire</h2>
        <button @click="closeDeleteModal" class="close-btn">
          <Icon name="x" />
        </button>
      </div>
      
      <div class="modal-body">
        <div class="warning-icon">
          <Icon name="alert-triangle" />
        </div>
        <p class="warning-text">
          Are you sure you want to delete 
          <strong>"{{ questionnaireToDelete?.title }}"</strong>?
        </p>
        <p class="warning-subtext">
          This action cannot be undone. All questions and responses will be permanently deleted.
        </p>
      </div>
      
      <div class="modal-footer">
        <button @click="closeDeleteModal" class="cancel-btn">
          Cancel
        </button>
        <button @click="confirmDelete" class="delete-btn">
          <Icon name="trash" />
          Delete Questionnaire
        </button>
      </div>
    </div>
  </div>

  <QRModal
    :show="qrModal.show"
    :url="qrModal.url"
    @close="qrModal.show = false"
  />
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getters } from '../store/auth'
import { questionnaireAPI } from '../services/api.js'
import Icon from '../components/Icon.vue'
import QRModal from '../components/QRModal.vue'

const router = useRouter()


const loading = ref(false)
const questionnaires = ref([])


const showDeleteModal = ref(false)
const qrModal = reactive({
  show: false,
  url: ''
})
const questionnaireToDelete = ref(null)


const stats = reactive({
  totalQuestionnaires: 0,
  totalResponses: 0
})


const currentUser = getters.currentUser


const loadData = async () => {
  loading.value = true
  try {

    const response = await questionnaireAPI.getMyQuestionnaires()
    
    const cleanedData = (response.data || []).map(q => ({
      ...q,
      title: (q.title || '').replace(/^"|"$/g, ''),
      description: (q.description || '').replace(/^"|"$/g, ''),
      responseCount: parseInt(q.responseCount || q.responses || 0, 10) || 0,
      createdAt: q.created_at || q.createdAt || q.date || Date.now()
    }))
    
    questionnaires.value = cleanedData
    console.log('Dashboard - Cleaned questionnaires:', cleanedData)
    updateStats()
    console.log('Dashboard - Stats after update:', stats)
  } catch (error) {
    console.error('Error loading questionnaires:', error)
    questionnaires.value = []
    updateStats()
  } finally {
    loading.value = false
  }
}

const updateStats = () => {
  stats.totalQuestionnaires = questionnaires.value.length

  stats.totalResponses = questionnaires.value.reduce((sum, q) => {
    const responseCount = parseInt(q.responseCount, 10) || 0
    return sum + responseCount
  }, 0)
}

const formatDate = (date) => {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }).format(date)
}

const createQuestionnaire = () => {
  router.push('/questionnaire/create')
}

const viewAll = () => {
  router.push('/questionnaires')
}

const refreshData = () => {
  loadData()
}

const openQuestionnaire = (id) => {
  router.push(`/questionnaire/${id}/questions`)
}

const editQuestionnaire = (id) => {
  router.push('/questionnaires')
}

const shareQuestionnaire = async (id) => {
  const questionnaire = questionnaires.value.find(q => q.id === id)
  if (!questionnaire) return
  
  if (!questionnaire.is_published) {
    console.warn('Cannot share unpublished questionnaire')
    return
  }

  try {
    const inviteResponse = await questionnaireAPI.generateInvite(id)
    const shareUrl = inviteResponse.data.join_url || inviteResponse.data.url || inviteResponse.data.invite_url
    
    if (shareUrl) {
      const fullUrl = shareUrl.startsWith('http') ? shareUrl : `${window.location.origin}${shareUrl}`
      qrModal.url = fullUrl
      qrModal.show = true
    } else {
      console.error('No URL received from invite generation')
      console.log('Response data:', inviteResponse.data)
    }
  } catch (error) {
    console.error('Error generating share URL:', error)
  }
}

const deleteQuestionnaire = (id) => {
  const questionnaire = questionnaires.value.find(q => q.id === id)
  if (!questionnaire) return
  
  questionnaireToDelete.value = questionnaire
  showDeleteModal.value = true
}

const confirmDelete = async () => {
  if (!questionnaireToDelete.value) return
  
  try {
    await questionnaireAPI.delete(questionnaireToDelete.value.id)
    questionnaires.value = questionnaires.value.filter(q => q.id !== questionnaireToDelete.value.id)
    updateStats()
    closeDeleteModal()
  } catch (error) {
    console.error('Error deleting questionnaire:', error)
  }
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  questionnaireToDelete.value = null
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.dashboard-header {
  margin-bottom: 2rem;
}

.dashboard-header h1 {
  font-size: 2.5rem;
  color: #111827;
  margin-bottom: 0.5rem;
}

.dashboard-header p {
  font-size: 1.1rem;
  color: #6b7280;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 3rem;
  height: 3rem;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eef2ff;
  color: #4f46e5;
}

.stat-icon.active {
  background: #ecfdf5;
  color: #10b981;
}

.stat-icon.completed {
  background: #fef3c7;
  color: #f59e0b;
}

.stat-icon.responses {
  background: #fce7f3;
  color: #ec4899;
}

.stat-info h3 {
  font-size: 1.75rem;
  font-weight: bold;
  color: #111827;
  margin: 0;
}

.stat-info p {
  color: #6b7280;
  margin: 0;
  font-size: 0.9rem;
}

.dashboard-actions {
  display: flex;
  gap: 1rem;
  margin-bottom: 3rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn.primary {
  background: #4f46e5;
  color: white;
}

.action-btn.primary:hover {
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

.recent-section {
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-header h2 {
  font-size: 1.5rem;
  color: #111827;
  margin: 0;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: none;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.refresh-btn:hover {
  background: #f3f4f6;
  color: #4f46e5;
}

.loading-state, .empty-state {
  text-align: center;
  padding: 3rem;
  color: #6b7280;
}

.empty-state h3 {
  color: #111827;
  margin: 1rem 0 0.5rem 0;
}

.create-btn {
  margin-top: 1rem;
  padding: 0.75rem 1.5rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
}

.create-btn:hover {
  background: #4338ca;
}

.questionnaires-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.questionnaire-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.2s;
}

.questionnaire-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 1rem;
}

.card-header h3 {
  color: #111827;
  margin: 0;
  font-size: 1.1rem;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: capitalize;
}

.status-badge.active {
  background: #ecfdf5;
  color: #065f46;
}

.status-badge.completed {
  background: #fef3c7;
  color: #92400e;
}

.card-description {
  color: #6b7280;
  margin-bottom: 1rem;
  line-height: 1.5;
}

.card-meta {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.85rem;
  color: #6b7280;
}

.card-actions {
  display: flex;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #f3f4f6;
}

.action-link {
  background: none;
  border: none;
  color: #4f46e5;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.2s;
}

.action-link:hover {
  color: #4338ca;
}

.action-link.danger {
  color: #dc2626;
}

.action-link.danger:hover {
  color: #b91c1c;
}
</style>