import { createRouter, createWebHistory } from 'vue-router'

// Importar componentes existentes
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Dashboard from '../views/Dashboard.vue'
import CreateQuestionnaire from '../views/CreateQuestionnaire.vue'
import Questionnaires from '../views/Questionnaires.vue'
import QuestionnaireQuestions from '../views/QuestionnaireQuestions.vue'
import QuestionnaireResponses from '../views/QuestionnaireResponses.vue'
import JoinQuestionnaire from '../views/JoinQuestionnaire.vue'
import AnswerQuestionnaire from '../views/AnswerQuestionnaire.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home,
    meta: { requiresAuth: false }
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register', 
    component: Register,
    meta: { requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/questionnaire/create',
    name: 'CreateQuestionnaire',
    component: CreateQuestionnaire,
    meta: { requiresAuth: true }
  },
  {
    path: '/questionnaires',
    name: 'Questionnaires',
    component: Questionnaires,
    meta: { requiresAuth: true }
  },
  {
    path: '/questionnaire/:id/questions',
    name: 'QuestionnaireQuestions',
    component: QuestionnaireQuestions,
    meta: { requiresAuth: true }
  },
  {
    path: '/questionnaire/:id/responses',
    name: 'QuestionnaireResponses',
    component: QuestionnaireResponses,
    meta: { requiresAuth: true }
  },
  {
    path: '/join/:token',
    name: 'JoinQuestionnaire',
    component: JoinQuestionnaire,
    meta: { requiresAuth: false }
  },
  {
    path: '/answer/:token',
    name: 'AnswerQuestionnaire', 
    component: AnswerQuestionnaire,
    meta: { requiresAuth: false }
  },
  // Duplicate routes removed
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const isAuthenticated = !!token

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  }
  else if (isAuthenticated && (to.name === 'Login' || to.name === 'Register')) {
    next('/dashboard')
  }
  else {
    next()
  }
})

export default router