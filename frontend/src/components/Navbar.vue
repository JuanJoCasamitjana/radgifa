<template>
  <nav class="navbar">
    <div class="navbar-container">
      <div class="navbar-brand">
        <router-link to="/" class="brand-link">
          {{ appName }}
        </router-link>
      </div>

      <div class="navbar-nav" :class="{ 'active': showMobileMenu }">
        <template v-if="isAuthenticated">
          <router-link to="/dashboard" class="nav-link">
            <Icon name="home" />
            <span>Dashboard</span>
          </router-link>
          
          <router-link to="/questionnaire/create" class="nav-link">
            <Icon name="plus" />
            <span>Create</span>
          </router-link>
          
          <div class="user-menu" ref="userMenuRef">
            <button 
              class="user-button" 
              @click="toggleUserMenu"
              :class="{ 'active': showUserMenu }"
            >
              <Icon name="user" />
              <span>{{ currentUser?.display_name || currentUser?.username || 'User' }}</span>
            </button>
            
            <div class="user-dropdown" v-show="showUserMenu">
              <div class="user-info">
                <p class="user-name">{{ currentUser?.name || currentUser?.username }}</p>
                <p class="user-username">@{{ currentUser?.username }}</p>
              </div>
              <hr class="dropdown-divider">
              <button @click="handleLogout" class="dropdown-item logout-item">
                <Icon name="logout" />
                <span>Sign Out</span>
              </button>
            </div>
          </div>
        </template>

        <template v-else>
          <router-link to="/login" class="nav-link">Sign In</router-link>
          <router-link to="/register" class="nav-link btn-primary">Get Started</router-link>
        </template>
      </div>

      <button 
        class="mobile-menu-button" 
        @click="toggleMobileMenu"
        v-if="isAuthenticated"
      >
        <Icon name="menu" />
      </button>
    </div>
  </nav>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getters, actions } from '../store/auth'
import Icon from './Icon.vue'

const router = useRouter()

const appName = import.meta.env.VITE_APP_NAME || 'Radgifa'
const showMobileMenu = ref(false)
const showUserMenu = ref(false)
const userMenuRef = ref(null)


const isAuthenticated = getters.isAuthenticated
const currentUser = getters.currentUser


const toggleMobileMenu = () => {
  showMobileMenu.value = !showMobileMenu.value
}


const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
}


const handleLogout = () => {
  actions.logout()
  showUserMenu.value = false
  router.push('/')
}


const handleClickOutside = (event) => {
  if (userMenuRef.value && !userMenuRef.value.contains(event.target)) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.navbar {
  background: white;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 50;
}

.navbar-container {
  padding: 0 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 4rem;
}

.navbar-brand {
  flex-shrink: 0;
}

.brand-link {
  font-size: 1.5rem;
  font-weight: bold;
  color: #4f46e5;
  text-decoration: none;
  transition: color 0.2s;
}

.brand-link:hover {
  color: #4338ca;
}

.navbar-nav {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  color: #374151;
  text-decoration: none;
  border-radius: 6px;
  transition: background-color 0.2s, color 0.2s;
  font-weight: 500;
}

.nav-link:hover {
  background-color: #f3f4f6;
  color: #4f46e5;
}

.nav-link.router-link-active {
  background-color: #eef2ff;
  color: #4f46e5;
}

.nav-link.btn-primary {
  background-color: #4f46e5;
  color: white;
}

.nav-link.btn-primary:hover {
  background-color: #4338ca;
  color: white;
}

.user-menu {
  position: relative;
}

.user-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: none;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  font-weight: 500;
  color: #374151;
}

.user-button:hover,
.user-button.active {
  border-color: #4f46e5;
  box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
}

.user-dropdown {
  position: absolute;
  right: 0;
  top: 100%;
  margin-top: 0.5rem;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  min-width: 200px;
  z-index: 10;
}

.user-info {
  padding: 1rem;
}

.user-name {
  font-weight: 600;
  color: #111827;
  margin: 0 0 0.25rem 0;
}

.user-username {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
}

.dropdown-divider {
  border: 0;
  border-top: 1px solid #e5e7eb;
  margin: 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.75rem 1rem;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: background-color 0.2s;
  color: #374151;
  font-weight: 500;
}

.dropdown-item:hover {
  background-color: #f3f4f6;
}

.logout-item:hover {
  background-color: #fef2f2;
  color: #dc2626;
}

.mobile-menu-button {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  color: #374151;
}
</style>