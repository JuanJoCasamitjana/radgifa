<script setup>
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import Navbar from './components/Navbar.vue'
import { actions } from './store/auth'

// Initialize authentication on app start
onMounted(() => {
  actions.initializeAuth()
})
</script>

<template>
  <div id="app">
    <Navbar />
    <main class="main-content">
      <!-- Aquí se renderizan todas las vistas del router -->
      <RouterView />
    </main>
  </div>
</template>

<style>
/* Estilos globales de la aplicación */
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  background-color: #f9fafb;
}

.main-content {
  flex: 1;
  width: 100%;
}

/* Form styles */
.form-container {
  padding: 2rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}
</style>
