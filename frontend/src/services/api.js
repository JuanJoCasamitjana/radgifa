import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || window.location.origin,
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
  
  getMyQuestionnaires: (config = {}) => api.get('/api/questionnaires', config),
  
  
  create: (data, config = {}) => api.post('/api/questionnaires', data, config),
  
  
  update: (id, data, config = {}) => api.put(`/api/questionnaires/${id}`, data, config),
  
  
  delete: (id, config = {}) => api.delete(`/api/questionnaires/${id}`, config),
  
  
  publish: (id, config = {}) => api.post(`/api/questionnaires/${id}/publish`, null, config),
  
  
  getDetails: (id, config = {}) => api.get(`/api/questionnaires/${id}`, config),
  
  
  getQuestions: (id, config = {}) => api.get(`/api/questionnaires/${id}/questions`, config),
  
  
  getMyAnswers: (id, config = {}) => api.get(`/api/questionnaires/${id}/my-answers`, config),
  
  
  getMembers: (id, config = {}) => api.get(`/api/questionnaires/${id}/members`, config),
  
  generateInvite: (id, config = {}) => api.post(`/api/questionnaires/${id}/invite`, null, config),
  
  
  createQuestion: (id, questionData, config = {}) => api.post(`/api/questionnaires/${id}/question`, questionData, config),
  
  
  updateQuestion: (questionnaireId, questionId, questionData, config = {}) => api.put(`/api/questionnaires/${questionnaireId}/questions/${questionId}`, questionData, config),
  
  
  deleteQuestion: (questionnaireId, questionId, config = {}) => api.delete(`/api/questionnaires/${questionnaireId}/questions/${questionId}`, config),
}

export const participationAPI = {
  joinQuestionnaire: (token, memberData) => api.post(`/join/${token}`, memberData),
  
  checkMemberIdentifier: (token, identifier) => api.post(`/check/member/${token}`, { value: identifier }),
  
  getQuestionnaireInfo: (token) => api.get(`/join/${token}/info`),
  
  answerQuestion: (questionId, answer, config = {}) => api.post(`/api/question/${questionId}`, answer, config),
}


export const healthAPI = {
  check: () => api.get('/health'),
}

export default api