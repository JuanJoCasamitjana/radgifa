import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
})

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

api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {

    if (error.response?.status === 401 && !window.location.pathname.includes('/login')) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authAPI = {
  register: (userData) => api.post('/register', userData),
  login: (credentials) => api.post('/login', credentials),
  checkUsername: (username) => api.post('/check/username', { value: username }),
}


export const questionnaireAPI = {
  
  getMyQuestionnaires: () => api.get('/api/questionnaires'),
  
  
  create: (data) => api.post('/api/questionnaires', data),
  
  
  update: (id, data) => api.put(`/api/questionnaires/${id}`, data),
  
  
  delete: (id) => api.delete(`/api/questionnaires/${id}`),
  
  
  publish: (id) => api.post(`/api/questionnaires/${id}/publish`),
  
  
  getDetails: (id) => api.get(`/api/questionnaires/${id}`),
  
  
  getQuestions: (id) => api.get(`/api/questionnaires/${id}/questions`),
  
  
  getMyAnswers: (id) => api.get(`/api/questionnaires/${id}/my-answers`),
  
  
  getMembers: (id) => api.get(`/api/questionnaires/${id}/members`),
  
  generateInvite: (id) => api.post(`/api/questionnaires/${id}/invite`),
  
  
  createQuestion: (id, questionData) => api.post(`/api/questionnaires/${id}/question`, questionData),
  
  
  updateQuestion: (questionnaireId, questionId, questionData) => api.put(`/api/questionnaires/${questionnaireId}/questions/${questionId}`, questionData),
  
  
  deleteQuestion: (questionnaireId, questionId) => api.delete(`/api/questionnaires/${questionnaireId}/questions/${questionId}`),
}

export const participationAPI = {
  
  joinQuestionnaire: (token, memberData) => api.post(`/join/${token}`, memberData),
  
  
  checkMemberIdentifier: (token, identifier) => api.post(`/check/member/${token}`, { value: identifier }),
  
  
  answerQuestion: (questionId, answer) => api.post(`/api/question/${questionId}`, answer),
}


export const healthAPI = {
  check: () => api.get('/health'),
}

export default api