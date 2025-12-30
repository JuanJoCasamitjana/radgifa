<template>
  <div v-if="show" class="modal-overlay" @click="handleClose">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>Share Questionnaire</h2>
        <button @click="handleClose" class="close-btn">
          <Icon name="x" />
        </button>
      </div>
      
      <div class="modal-body">
        <div class="qr-container">
          <canvas ref="qrCanvas"></canvas>
        </div>
        
        <div class="url-section">
          <label>Invitation Link</label>
          <div class="url-container">
            <input 
              ref="urlInput"
              :value="url" 
              readonly 
              class="url-input"
              @focus="$event.target.select()"
            />
            <button @click="copyToClipboard" class="copy-btn" :class="{ 'copied': copied }">
              <Icon :name="copied ? 'check' : 'copy'" />
              <span>{{ copied ? 'Copied!' : 'Copy' }}</span>
            </button>
          </div>
          <p v-if="copied" class="success-message">Link copied to clipboard!</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import QRCode from 'qrcode'
import Icon from './Icon.vue'

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  url: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['close'])

const qrCanvas = ref(null)
const urlInput = ref(null)
const copied = ref(false)
let copiedTimeout = null

watch(() => props.show, async (newVal) => {
  if (newVal && props.url) {
    await nextTick()
    generateQR()
  }
})

watch(() => props.url, () => {
  if (props.show && props.url) {
    generateQR()
  }
})

const generateQR = async () => {
  if (!qrCanvas.value || !props.url) return
  
  try {
    await QRCode.toCanvas(qrCanvas.value, props.url, {
      width: 256,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      }
    })
  } catch (error) {
    console.error('Error generating QR code:', error)
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(props.url)
    copied.value = true
    
    if (copiedTimeout) {
      clearTimeout(copiedTimeout)
    }
    
    copiedTimeout = setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (error) {
    console.error('Error copying to clipboard:', error)
    urlInput.value?.select()
    document.execCommand('copy')
    copied.value = true
    
    if (copiedTimeout) {
      clearTimeout(copiedTimeout)
    }
    
    copiedTimeout = setTimeout(() => {
      copied.value = false
    }, 2000)
  }
}

const handleClose = () => {
  copied.value = false
  if (copiedTimeout) {
    clearTimeout(copiedTimeout)
  }
  emit('close')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.5rem;
  border-bottom: 1px solid #f3f4f6;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  color: #6b7280;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #111827;
}

.modal-body {
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.qr-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1.5rem;
  background: #f9fafb;
  border-radius: 8px;
  border: 2px solid #e5e7eb;
}

.qr-container canvas {
  display: block;
  max-width: 100%;
  height: auto;
}

.url-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.url-section label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.url-container {
  display: flex;
  gap: 0.5rem;
}

.url-input {
  flex: 1;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  font-family: monospace;
  background: #f9fafb;
  color: #111827;
  cursor: text;
}

.url-input:focus {
  outline: none;
  border-color: #4f46e5;
  background: white;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.copy-btn:hover {
  background: #4338ca;
}

.copy-btn.copied {
  background: #10b981;
}

.success-message {
  margin: 0;
  font-size: 0.875rem;
  color: #10b981;
  font-weight: 500;
}

@media (max-width: 640px) {
  .modal-content {
    max-width: 100%;
    margin: 0;
    border-radius: 0;
  }
  
  .modal-body {
    padding: 1.5rem;
  }
  
  .url-container {
    flex-direction: column;
  }
  
  .copy-btn {
    justify-content: center;
  }
}
</style>
