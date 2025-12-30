<template>
  <div class="home">
    <header class="hero">
      <h1>{{ appName }}</h1>
      <p class="subtitle">
        Struggling to reach a decision with your friends? 
        Radgifa is here to help you make an informed choice.
      </p>
      
      <div class="cta-buttons" v-if="!isAuthenticated">
        <router-link to="/register" class="btn btn-primary">
          Get Started
        </router-link>
        <router-link to="/login" class="btn btn-secondary">
          Sign In
        </router-link>
      </div>
      
      <div class="cta-buttons" v-else>
        <router-link to="/dashboard" class="btn btn-primary">
          Go to Dashboard
        </router-link>
      </div>
    </header>
    
    <section class="features">
      <h2>How it works</h2>
      <div class="features-grid">
        <div class="feature">
          <h3>1. Create</h3>
          <p>Create a questionnaire with your decision options</p>
        </div>
        <div class="feature">
          <h3>2. Share</h3>
          <p>Invite friends to participate with a simple link</p>
        </div>
        <div class="feature">
          <h3>3. Decide</h3>
          <p>See results and make informed group decisions</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { getters } from '../store/auth'

const appName = import.meta.env.VITE_APP_NAME || 'Radgifa'
const isAuthenticated = getters.isAuthenticated
</script>

<style scoped>
.home {
  min-height: 100vh;
}

.hero {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  text-align: center;
  padding: 4rem 2rem;
}

.hero h1 {
  font-size: 3rem;
  margin-bottom: 1rem;
  font-weight: bold;
}

.subtitle {
  font-size: 1.25rem;
  margin-bottom: 2rem;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn {
  padding: 0.75rem 2rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  transition: transform 0.2s, box-shadow 0.2s;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

.btn-primary {
  background: #4f46e5;
  color: white;
}

.btn-secondary {
  background: transparent;
  color: white;
  border: 2px solid white;
}

.features {
  padding: 4rem 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.features h2 {
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 3rem;
  color: #1f2937;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.feature {
  text-align: center;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.feature h3 {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  color: #4f46e5;
}

.feature p {
  color: #6b7280;
  line-height: 1.6;
}
</style>