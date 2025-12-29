import { createRouter, createWebHistory } from 'vue-router'

// Importar componentes existentes
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Dashboard from '../views/Dashboard.vue'
import CreateQuestionnaire from '../views/CreateQuestionnaire.vue'
import Questionnaires from '../views/Questionnaires.vue'

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
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Guard de navegación para rutas protegidas
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const isAuthenticated = !!token

  // Si la ruta requiere autenticación y no está autenticado
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  }
  // Si está autenticado y va a login/register, redirigir al dashboard
  else if (isAuthenticated && (to.name === 'Login' || to.name === 'Register')) {
    next('/dashboard')
  }
  // Permitir navegación normal
  else {
    next()
  }
})

export default router