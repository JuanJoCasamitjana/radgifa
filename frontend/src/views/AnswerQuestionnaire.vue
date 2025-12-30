<template>
  <div class="answer-page">
    <div class="page-container">
      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>Cargando cuestionario...</p>
      </div>

      <div v-else-if="error" class="alert alert-error">
        <div class="alert-icon">
          <Icon name="alert-triangle" />
        </div>
        <div>
          <h3>Error</h3>
          <p>{{ error }}</p>
        </div>
      </div>

      <div v-else>
        <div v-if="showSuccessMessage" class="banner banner-success animate-fade-in">
          <Icon name="check-circle" />
          <div>
            <h3>¡Listo para participar!</h3>
            <p>Te has unido exitosamente al cuestionario. Responde las preguntas a continuación.</p>
          </div>
        </div>

        <div v-if="answerSuccess" class="banner banner-secondary animate-fade-in">
          <Icon name="check-circle" />
          <p>{{ answerSuccess }}</p>
        </div>

        <div class="card">
          <div class="card-header">
            <h1>{{ questionnaire.title }}</h1>
            <p v-if="questionnaire.description" class="subtitle">{{ questionnaire.description }}</p>
            <div class="participant">
              <Icon name="user" />
              <span>Participando como: <strong>{{ memberInfo.display_name || memberInfo.unique_identifier }}</strong></span>
            </div>
          </div>

          <div v-if="questions.length === 0" class="empty-state">
            <Icon name="clipboard" />
            <p>Este cuestionario aún no tiene preguntas.</p>
          </div>

          <div v-else class="content">
            <div class="progress-box">
              <div class="progress-header">
                <Icon name="info" />
                <p>Progreso: {{ answeredQuestions.length }} de {{ questions.length }} preguntas respondidas</p>
              </div>
              <div class="progress-bar">
                <div class="progress-bar-fill" :style="{ width: progressPercentage + '%' }"></div>
              </div>
            </div>

            <div class="questions-list">
              <div v-for="(question, index) in questions" :key="question.id" class="question-card">
                <div class="question-header">
                  <div>
                    <h3>{{ index + 1 }}. {{ question.text }}</h3>
                    <p v-if="question.theme" class="question-theme">
                      <Icon name="tag" class="inline-icon" />
                      {{ question.theme }}
                    </p>
                  </div>
                  <div v-if="getAnswerForQuestion(question.id)" class="answered-badge">
                    <Icon name="check-circle" />
                    <span>Respondida</span>
                  </div>
                </div>

                <div class="question-body">
                  <div class="options">
                    <button
                      v-for="option in answerOptions"
                      :key="option.value"
                      @click="submitAnswer(question.id, option.value)"
                      :disabled="submittingAnswers[question.id]"
                      :class="['option-button', {
                        selected: getAnswerForQuestion(question.id)?.answer_value === option.value,
                        loading: submittingAnswers[question.id] && pendingAnswers[question.id] === option.value
                      }]"
                    >
                      <template v-if="submittingAnswers[question.id] && pendingAnswers[question.id] === option.value">
                        <div class="mini-spinner"></div>
                        {{ option.label }}
                      </template>
                      <template v-else>
                        <Icon :name="option.icon" class="inline-icon" />
                        {{ option.label }}
                      </template>
                    </button>
                  </div>

                  <div v-if="getAnswerForQuestion(question.id)" class="answer-meta">
                    Respondido el {{ formatDate(getAnswerForQuestion(question.id).created_at) }}
                  </div>
                </div>
              </div>
            </div>

            <div v-if="answeredQuestions.length === questions.length && questions.length > 0" class="completion-box">
              <Icon name="check-circle" />
              <div>
                <h3>¡Cuestionario completo!</h3>
                <p>Has respondido todas las preguntas. Puedes cambiar tus respuestas en cualquier momento.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { participationAPI, questionnaireAPI } from '../services/api'
import Icon from '../components/Icon.vue'

export default {
  name: 'AnswerQuestionnaire',
  components: {
    Icon
  },
  setup() {
    const route = useRoute()
    
    const loading = ref(true)
    const error = ref(null)
      const showSuccessMessage = ref(false)
    const answerSuccess = ref('')
    const questionnaire = ref({})
    const questions = ref([])
    const myAnswers = ref([])
    const memberInfo = ref({})
    
    const submittingAnswers = reactive({})
    const pendingAnswers = reactive({})
    
    const token = computed(() => route.params.token)
    
    const answerOptions = [
      { value: 'Yes', label: 'Yes', icon: 'check' },
      { value: 'No', label: 'No', icon: 'x' },
      { value: 'Pass', label: 'Pass', icon: 'minus' }
    ]
    
    const answeredQuestions = computed(() => {
      return questions.value.filter(q => getAnswerForQuestion(q.id))
    })
    
    const progressPercentage = computed(() => {
      if (questions.value.length === 0) return 0
      return Math.round((answeredQuestions.value.length / questions.value.length) * 100)
    })
    
    const getAnswerForQuestion = (questionId) => {
      return myAnswers.value.find(answer => answer.edges?.question?.id === questionId)
    }
    
    const formatDate = (dateString) => {
      if (!dateString) return ''
      try {
        const date = new Date(dateString)
        return date.toLocaleString('es-ES', {
          day: '2-digit',
          month: '2-digit', 
          year: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        })
      } catch (err) {
        return dateString
      }
    }
    
    const loadQuestionnaireData = async () => {
      try {
        loading.value = true
        error.value = null
        
        const memberToken = localStorage.getItem('member_token')
        const userToken = localStorage.getItem('token')
        const authToken = memberToken || userToken
        const authConfig = authToken
          ? { headers: { Authorization: `Bearer ${authToken}` } }
          : {}
        
        const tokenInfo = await participationAPI.getQuestionnaireInfo(token.value)
        const questionnaireId = tokenInfo.data.questionnaire_id
        
        if (!questionnaireId) {
          throw new Error('No se pudo obtener la información del cuestionario')
        }
        
        const [questResponse, questionsResponse, answersResponse] = await Promise.all([
          questionnaireAPI.getDetails(questionnaireId, authConfig),
          questionnaireAPI.getQuestions(questionnaireId, authConfig),
          questionnaireAPI.getMyAnswers(questionnaireId, authConfig)
        ])
        
        questionnaire.value = questResponse.data
        questions.value = questionsResponse.data || []
        myAnswers.value = answersResponse.data || []
        
        const memberId = localStorage.getItem('member_id')
        if (memberId) {
          memberInfo.value = {
            id: memberId,
            unique_identifier: localStorage.getItem('member_identifier') || 'Participante',
            display_name: localStorage.getItem('member_display_name') || null
          }
        }
        
      } catch (err) {
        console.error('Error loading questionnaire:', err)
        if (err.response?.status === 403) {
          error.value = 'No tienes permisos para acceder a este cuestionario.'
        } else if (err.response?.status === 404) {
          error.value = 'Cuestionario no encontrado.'
        } else if (err.response?.status === 400) {
          error.value = 'El enlace de invitación ha expirado o es inválido.'
        } else if (!localStorage.getItem('member_token') && !localStorage.getItem('token')) {
          error.value = 'Inicia sesión o únete desde el enlace de invitación para continuar.'
        } else {
          error.value = 'Error al cargar el cuestionario. Intenta nuevamente.'
        }
      } finally {
        loading.value = false
        
              showSuccessMessage.value = true
              setTimeout(() => {
                showSuccessMessage.value = false
              }, 5000)
      }
    }
    
    const submitAnswer = async (questionId, answerValue) => {
      submittingAnswers[questionId] = true
      pendingAnswers[questionId] = answerValue
      
      try {
        const memberToken = localStorage.getItem('member_token')
        const userToken = localStorage.getItem('token')
        const authToken = memberToken || userToken

        if (!authToken) {
          throw new Error('No hay token válido')
        }

        const authConfig = { headers: { Authorization: `Bearer ${authToken}` } }

        await participationAPI.answerQuestion(questionId, { answer_value: answerValue }, authConfig)
        
        const existingAnswerIndex = myAnswers.value.findIndex(
          answer => answer.edges?.question?.id === questionId
        )
        
        const newAnswer = {
          id: Date.now(),
          answer_value: answerValue,
          created_at: new Date().toISOString(),
          edges: {
            question: { id: questionId }
          }
        }
        
        if (existingAnswerIndex >= 0) {
          myAnswers.value[existingAnswerIndex] = newAnswer
        } else {
          myAnswers.value.push(newAnswer)
        }

        answerSuccess.value = 'Respuesta guardada'
        setTimeout(() => {
          answerSuccess.value = ''
        }, 3000)
        
      } catch (err) {
        console.error('Error submitting answer:', err)
        let errorMessage = 'Error al enviar respuesta'
        
        if (err.response?.status === 401) {
          errorMessage = 'Sesión expirada. Necesitas unirte nuevamente al cuestionario.'
        } else if (err.response?.status === 403) {
          errorMessage = 'No tienes permisos para responder esta pregunta.'
        } else if (err.response?.status === 404) {
          errorMessage = 'Pregunta no encontrada.'
        } else if (err.message === 'No hay token válido') {
          errorMessage = 'Inicia sesión o únete desde el enlace antes de responder.'
        }
        
        alert(errorMessage)
      } finally {
        submittingAnswers[questionId] = false
        delete pendingAnswers[questionId]
      }
    }
    
    onMounted(() => {
      loadQuestionnaireData()
    })
    
    return {
      loading,
      error,
      showSuccessMessage,
      answerSuccess,
      questionnaire,
      questions,
      myAnswers,
      memberInfo,
      submittingAnswers,
      pendingAnswers,
      answerOptions,
      answeredQuestions,
      progressPercentage,
      getAnswerForQuestion,
      formatDate,
      submitAnswer
    }
  }
}
</script>

<style scoped>
.answer-page {
  background: #f5f6fb;
  min-height: 100vh;
  padding: 2rem 1rem;
}

.page-container {
  max-width: 960px;
  margin: 0 auto;
}

.loading-state {
  text-align: center;
  padding: 3rem 0;
  color: #4b5563;
}

.spinner {
  width: 48px;
  height: 48px;
  border: 4px solid #e5e7eb;
  border-top-color: #4f46e5;
  border-radius: 50%;
  margin: 0 auto 1rem auto;
  animation: spin 0.8s linear infinite;
}

.alert {
  display: flex;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 10px;
  border: 1px solid #f5c2c7;
  background: #fef2f2;
  color: #b91c1c;
  align-items: flex-start;
}

.alert-icon {
  color: #ef4444;
  flex-shrink: 0;
}

.alert h3 {
  margin: 0 0 0.25rem 0;
  font-size: 0.95rem;
  font-weight: 600;
}

.alert p {
  margin: 0;
  font-size: 0.95rem;
}

.banner {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  margin-bottom: 1rem;
  border: 1px solid transparent;
  background: #f0fdf4;
  color: #166534;
}

.banner-secondary {
  background: #ecfdf3;
  border-color: #bbf7d0;
  color: #065f46;
}

.banner h3 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  font-weight: 600;
}

.banner p {
  margin: 0;
  font-size: 0.95rem;
  color: #14532d;
}

.card {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.05);
  border: 1px solid #e5e7eb;
  padding: 1.75rem;
}

.card-header h1 {
  margin: 0;
  color: #111827;
  font-size: 1.6rem;
  line-height: 1.3;
}

.subtitle {
  margin-top: 0.5rem;
  color: #4b5563;
}

.participant {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-top: 1rem;
  color: #6b7280;
  font-size: 0.95rem;
}

.participant strong {
  color: #111827;
}

.empty-state {
  text-align: center;
  color: #6b7280;
  padding: 2.5rem 1rem;
}

.empty-state svg {
  width: 48px;
  height: 48px;
  color: #9ca3af;
  margin-bottom: 0.75rem;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 1rem;
}

.progress-box {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 1rem;
  background: #f8fafc;
}

.progress-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1d4ed8;
  font-weight: 600;
}

.progress-header p {
  margin: 0;
  color: #1f2937;
  font-weight: 500;
}

.progress-bar {
  margin-top: 0.75rem;
  width: 100%;
  height: 10px;
  border-radius: 999px;
  background: #dbeafe;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: #2563eb;
  border-radius: inherit;
  transition: width 0.3s ease;
}

.questions-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.question-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  background: #ffffff;
}

.question-header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  margin-bottom: 0.75rem;
}

.question-header h3 {
  margin: 0;
  font-size: 1.05rem;
  color: #111827;
}

.question-theme {
  margin: 0.35rem 0 0 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.inline-icon {
  width: 14px;
  height: 14px;
  margin-right: 4px;
  vertical-align: middle;
}

.answered-badge {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.6rem;
  background: #ecfdf3;
  color: #15803d;
  border: 1px solid #bbf7d0;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 600;
}

.question-body {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.65rem;
}

.option-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  padding: 0.75rem;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: #ffffff;
  color: #374151;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.option-button:hover {
  border-color: #4f46e5;
  color: #4338ca;
  box-shadow: 0 4px 10px rgba(79, 70, 229, 0.12);
}

.option-button.selected {
  border-color: #10b981;
  background: #ecfdf3;
  color: #047857;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.18);
}

.option-button:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.option-button.loading {
  color: #374151;
}

.mini-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid #e5e7eb;
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.answer-meta {
  font-size: 0.9rem;
  color: #6b7280;
}

.completion-box {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  border: 1px solid #bbf7d0;
  background: #f0fdf4;
  color: #15803d;
  margin-top: 0.5rem;
}

.completion-box h3 {
  margin: 0 0 0.2rem 0;
  font-size: 1.05rem;
}

.completion-box p {
  margin: 0;
  color: #166534;
}

.animate-fade-in {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .card {
    padding: 1.25rem;
  }

  .question-header {
    flex-direction: column;
  }

  .progress-header {
    align-items: flex-start;
  }
}
</style>