import axios from 'axios'

// Configuración base de Axios
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
})

// Interceptor para agregar token JWT automáticamente
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Interceptor para manejar respuestas y errores
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    // Si el token expiró (401), redirigir al login
    // PERO solo si NO estamos en la página de login
    if (error.response?.status === 401 && !window.location.pathname.includes('/login')) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // Redirigir al login (se puede hacer desde el router)
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API methods para autenticación
export const authAPI = {
  register: (userData) => api.post('/register', userData),
  login: (credentials) => api.post('/login', credentials),
  checkUsername: (username) => api.post('/check/username', { value: username }),
}

// API methods para cuestionarios
export const questionnaireAPI = {
  // Obtener todos mis cuestionarios
  getMyQuestionnaires: () => api.get('/api/questionnaires'),
  
  // Crear nuevo cuestionario
  create: (data) => api.post('/api/questionnaires', data),
  
  // Actualizar cuestionario
  update: (id, data) => api.put(`/api/questionnaires/${id}`, data),
  
  // Eliminar cuestionario
  delete: (id) => api.delete(`/api/questionnaires/${id}`),
  
  // Publicar cuestionario
  publish: (id) => api.post(`/api/questionnaires/${id}/publish`),
  
  // Obtener detalles de un cuestionario
  getDetails: (id) => api.get(`/api/questionnaires/${id}`),
  
  // Obtener preguntas de un cuestionario
  getQuestions: (id) => api.get(`/api/questionnaires/${id}/questions`),
  
  // Obtener mis respuestas
  getMyAnswers: (id) => api.get(`/api/questionnaires/${id}/my-answers`),
  
  // Obtener miembros (solo owners)
  getMembers: (id) => api.get(`/api/questionnaires/${id}/members`),
  
  // Generar invitación
  generateInvite: (id) => api.post(`/api/questionnaires/${id}/invite`),
  
  // Crear nueva pregunta
  createQuestion: (id, questionData) => api.post(`/api/questionnaires/${id}/question`, questionData),
  
  // Actualizar pregunta
  updateQuestion: (questionnaireId, questionId, questionData) => api.put(`/api/questionnaires/${questionnaireId}/questions/${questionId}`, questionData),
  
  // Eliminar pregunta
  deleteQuestion: (questionnaireId, questionId) => api.delete(`/api/questionnaires/${questionnaireId}/questions/${questionId}`),
}

// API methods para participación
export const participationAPI = {
  // Unirse a cuestionario
  joinQuestionnaire: (token, memberData) => api.post(`/join/${token}`, memberData),
  
  // Verificar disponibilidad de identificador
  checkMemberIdentifier: (token, identifier) => api.post(`/check/member/${token}`, { value: identifier }),
  
  // Responder pregunta
  answerQuestion: (questionId, answer) => api.post(`/api/question/${questionId}`, answer),
}

// API method para health check
export const healthAPI = {
  check: () => api.get('/health'),
}

export default api