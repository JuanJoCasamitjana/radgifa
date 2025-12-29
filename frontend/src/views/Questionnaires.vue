<template>
  <div class="questionnaires">
    <div class="page-header">
      <div class="header-content">
        <h1>All Questionnaires</h1>
        <p>Manage and view all your questionnaires</p>
      </div>
      
      <div class="header-actions">
        <button @click="createNew" class="action-btn primary">
          <Icon name="plus" />
          Create New
        </button>
        
        <button @click="refreshData" class="action-btn secondary">
          <Icon name="refresh" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <div class="filter-group">
        <label>Search:</label>
        <input
          v-model="filters.search"
          type="text"
          placeholder="Search by title..."
          class="search-input"
        />
      </div>
      
      <div class="results-info">
        <span>{{ filteredQuestionnaires.length }} questionnaire(s) found</span>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <Icon name="loading" />
      Loading questionnaires...
    </div>

    <!-- Empty State -->
    <div v-else-if="questionnaires.length === 0" class="empty-state">
      <Icon name="plus" />
      <h3>No questionnaires found</h3>
      <p>You haven't created any questionnaires yet.</p>
      <button @click="createNew" class="create-btn">
        <Icon name="plus" />
        Create Your First Questionnaire
      </button>
    </div>

    <!-- Questionnaires List -->
    <div v-else class="questionnaires-container">
      <!-- List View Toggle -->
      <div class="view-controls">
        <div class="view-toggle">
          <button 
            @click="viewMode = 'grid'"
            :class="{ 'active': viewMode === 'grid' }"
            class="toggle-btn"
          >
            Grid
          </button>
          <button 
            @click="viewMode = 'list'"
            :class="{ 'active': viewMode === 'list' }"
            class="toggle-btn"
          >
            List
          </button>
        </div>
        
        <div class="sort-controls">
          <label>Sort by:</label>
          <select v-model="sortBy" class="sort-select">
            <option value="created_desc">Newest First</option>
            <option value="created_asc">Oldest First</option>
            <option value="title_asc">Title A-Z</option>
            <option value="title_desc">Title Z-A</option>
            <option value="responses_desc">Most Responses</option>
          </select>
        </div>
      </div>

      <!-- Grid View -->
      <div v-if="viewMode === 'grid'" class="questionnaires-grid">
        <template v-for="questionnaire in sortedQuestionnaires" :key="questionnaire.id">
          <div 
            v-if="questionnaire && questionnaire.title"
            class="questionnaire-card"
            @click="openQuestionnaire(questionnaire.id)"
          >
          <div class="card-header">
            <h3>{{ questionnaire.title }}</h3>
          </div>
          
          <p class="card-description">{{ questionnaire.description || 'No description provided' }}</p>
          
          <div class="card-stats">
            <div class="stat-item">
              <Icon name="user" />
              <span>{{ questionnaire.responseCount }} responses</span>
            </div>
            <div class="stat-item">
              <Icon name="calendar" />
              <span>{{ formatDate(questionnaire.createdAt) }}</span>
            </div>
          </div>
          
          <div class="card-actions">
            <button @click.stop="editQuestionnaire(questionnaire.id)" class="action-link">
              Edit
            </button>
            <button @click.stop="viewResponses(questionnaire.id)" class="action-link">
              Responses
            </button>
            <button @click.stop="shareQuestionnaire(questionnaire.id)" class="action-link">
              Share
            </button>
            <button @click.stop="deleteQuestionnaire(questionnaire.id)" class="action-link danger">
              Delete
            </button>
          </div>
        </div>
        </template>
      </div>

      <!-- List View -->
      <div v-else class="questionnaires-list">
        <div class="list-header">
          <div class="col-title">Title</div>
          <div class="col-responses">Responses</div>
          <div class="col-created">Created</div>
          <div class="col-actions">Actions</div>
        </div>
        
        <template v-for="questionnaire in sortedQuestionnaires" :key="questionnaire.id">
          <div 
            v-if="questionnaire && questionnaire.title"
            class="list-item"
            @click="openQuestionnaire(questionnaire.id)"
          >
          <div class="col-title">
            <div class="item-title">{{ questionnaire.title }}</div>
            <div class="item-description">{{ questionnaire.description || 'No description' }}</div>
          </div>
          <div class="col-responses">{{ questionnaire.responseCount }}</div>
          <div class="col-created">{{ formatDate(questionnaire.createdAt) }}</div>
          <div class="col-actions">
            <button @click.stop="editQuestionnaire(questionnaire.id)" class="action-btn small">
              Edit
            </button>
            <button @click.stop="shareQuestionnaire(questionnaire.id)" class="action-btn small">
              Share
            </button>
            <button @click.stop="deleteQuestionnaire(questionnaire.id)" class="action-btn small danger">
              Delete
            </button>
          </div>
        </div>
        </template>
      </div>

      <!-- No Results -->
      <div v-if="filteredQuestionnaires.length === 0 && questionnaires.length > 0" class="no-results">
        <Icon name="x" />
        <h3>No questionnaires match your filters</h3>
        <p>Try adjusting your search terms or filters.</p>
        <button @click="clearFilters" class="clear-btn">Clear Filters</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { questionnaireAPI } from '../services/api.js'
import Icon from '../components/Icon.vue'

const router = useRouter()

// Estado
const loading = ref(false)
const questionnaires = ref([])
const viewMode = ref('grid')
const sortBy = ref('created_desc')

// Filtros
const filters = reactive({
  search: ''
})

// Cuestionarios filtrados
const filteredQuestionnaires = computed(() => {
  let filtered = questionnaires.value

  if (filters.search) {
    const search = filters.search.toLowerCase()
    filtered = filtered.filter(q => 
      q.title.toLowerCase().includes(search) ||
      (q.description && q.description.toLowerCase().includes(search))
    )
  }

  return filtered
})

// Cuestionarios ordenados
const sortedQuestionnaires = computed(() => {
  const sorted = [...filteredQuestionnaires.value]

  switch (sortBy.value) {
    case 'created_desc':
      return sorted.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
    case 'created_asc':
      return sorted.sort((a, b) => new Date(a.createdAt) - new Date(b.createdAt))
    case 'title_asc':
      return sorted.sort((a, b) => a.title.localeCompare(b.title))
    case 'title_desc':
      return sorted.sort((a, b) => b.title.localeCompare(a.title))
    case 'responses_desc':
      return sorted.sort((a, b) => (b.responseCount || 0) - (a.responseCount || 0))
    default:
      return sorted
  }
})

// Cargar datos
const loadData = async () => {
  loading.value = true
  try {
    // Load real questionnaires from API
    const response = await questionnaireAPI.getMyQuestionnaires()
    console.log('API Response:', response)
    console.log('API Data:', response.data)
    
    // Clean up and normalize the data
    const cleanedData = (response.data || []).map(q => ({
      ...q,
      // Remove extra quotes from strings
      title: (q.title || '').replace(/^"|"$/g, ''),
      description: (q.description || '').replace(/^"|"$/g, ''),
      // Add missing properties with defaults
      responseCount: q.responseCount || q.responses || 0,
      createdAt: q.created_at || q.createdAt || q.date || Date.now()
    }))
    
    questionnaires.value = cleanedData
    
    // Log first questionnaire structure for debugging
    if (questionnaires.value.length > 0) {
      console.log('First cleaned questionnaire:', questionnaires.value[0])
    }
  } catch (error) {
    console.error('Error loading questionnaires:', error)
    // Initialize empty if API fails
    questionnaires.value = []
  } finally {
    loading.value = false
  }
}

// Refrescar datos
const refreshData = () => {
  loadData()
}

// Formatear fecha
const formatDate = (date) => {
  if (!date) return 'No date'
  
  const dateObj = new Date(date)
  if (isNaN(dateObj.getTime())) {
    return 'Invalid date'
  }
  
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  }).format(dateObj)
}

// Limpiar filtros
const clearFilters = () => {
  filters.search = ''
}

// Acciones
const createNew = () => {
  router.push('/questionnaire/create')
}

const openQuestionnaire = (id) => {
  alert(`Opening questionnaire ${id} - functionality coming soon`)
}

const editQuestionnaire = (id) => {
  alert(`Editing questionnaire ${id} - functionality coming soon`)
}

const viewResponses = (id) => {
  alert(`Viewing responses for questionnaire ${id} - functionality coming soon`)
}

const shareQuestionnaire = (id) => {
  alert(`Sharing questionnaire ${id} - functionality coming soon`)
}

const deleteQuestionnaire = (id) => {
  if (confirm('Are you sure you want to delete this questionnaire? This action cannot be undone.')) {
    questionnaires.value = questionnaires.value.filter(q => q.id !== id)
  }
}

// Cargar datos al montar
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.questionnaires {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.header-content h1 {
  font-size: 2.5rem;
  color: #111827;
  margin: 0 0 0.5rem 0;
}

.header-content p {
  color: #6b7280;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.filters-bar {
  background: white;
  padding: 1.5rem;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
  display: flex;
  gap: 2rem;
  align-items: center;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-group label {
  font-weight: 500;
  color: #374151;
}

.filter-select,
.search-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  min-width: 120px;
}

.search-input {
  min-width: 200px;
}

.results-info {
  margin-left: auto;
  color: #6b7280;
  font-size: 0.875rem;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #6b7280;
}

.empty-state h3 {
  color: #111827;
  margin: 1rem 0 0.5rem 0;
}

.create-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 1rem auto 0 auto;
  padding: 0.75rem 1.5rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}

.questionnaires-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.view-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.view-toggle {
  display: flex;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  overflow: hidden;
}

.toggle-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: white;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn.active {
  background: #4f46e5;
  color: white;
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sort-select {
  padding: 0.25rem 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
}

.questionnaires-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
  padding: 1.5rem;
}

.questionnaire-card {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1.5rem;
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

.status-badge.draft {
  background: #f3f4f6;
  color: #374151;
}

.card-description {
  color: #6b7280;
  margin-bottom: 1rem;
  line-height: 1.5;
}

.card-stats {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  color: #6b7280;
}

.card-actions {
  display: flex;
  gap: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #f3f4f6;
}

.questionnaires-list {
  display: flex;
  flex-direction: column;
}

.list-header {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1.5fr;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.list-item {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1.5fr;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: background-color 0.2s;
}

.list-item:hover {
  background: #f9fafb;
}

.item-title {
  font-weight: 600;
  color: #111827;
}

.item-description {
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.col-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
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

.action-btn.small {
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.action-btn.small:hover {
  background: #e5e7eb;
}

.action-link {
  background: none;
  border: none;
  color: #4f46e5;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: color 0.2s;
}

.action-link:hover {
  color: #4338ca;
}

.action-link.danger,
.action-btn.danger {
  color: #dc2626;
}

.action-link.danger:hover,
.action-btn.danger:hover {
  color: #b91c1c;
  background: #fef2f2;
}

.no-results {
  text-align: center;
  padding: 4rem 2rem;
  color: #6b7280;
}

.no-results h3 {
  color: #111827;
  margin: 1rem 0 0.5rem 0;
}

.clear-btn {
  margin-top: 1rem;
  padding: 0.5rem 1rem;
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
}

.clear-btn:hover {
  background: #e5e7eb;
}
</style>