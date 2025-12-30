<template>
  <div v-if="show" class="modal-overlay" @click="handleCancel">
    <div class="confirm-modal" @click.stop>
      <div class="modal-header">
        <div class="icon-container" :class="type">
          <Icon :name="iconName" />
        </div>
        <h3>{{ title }}</h3>
      </div>
      
      <div class="modal-body">
        <p>{{ message }}</p>
      </div>
      
      <div class="modal-footer">
        <button 
          @click="handleCancel" 
          class="cancel-btn"
          :disabled="loading"
        >
          {{ cancelText }}
        </button>
        <button 
          @click="handleConfirm" 
          class="confirm-btn"
          :class="type"
          :disabled="loading"
        >
          <Icon v-if="loading" name="loading" />
          {{ loading ? 'Processing...' : confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Confirm Action'
  },
  message: {
    type: String,
    required: true
  },
  confirmText: {
    type: String,
    default: 'Confirm'
  },
  cancelText: {
    type: String,
    default: 'Cancel'
  },
  type: {
    type: String,
    default: 'warning',
    validator: (value) => ['warning', 'danger', 'info'].includes(value)
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['confirm', 'cancel'])

const iconName = computed(() => {
  switch (props.type) {
    case 'danger': return 'trash'
    case 'warning': return 'alert-triangle'
    case 'info': return 'info'
    default: return 'alert-triangle'
  }
})

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
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
  backdrop-filter: blur(2px);
}

.confirm-modal {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 400px;
  width: 90%;
  margin: 0 20px;
  overflow: hidden;
}

.modal-header {
  padding: 24px 24px 16px 24px;
  text-align: center;
}

.icon-container {
  width: 48px;
  height: 48px;
  margin: 0 auto 16px auto;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-container.warning {
  background: #fef3c7;
  color: #f59e0b;
}

.icon-container.danger {
  background: #fee2e2;
  color: #ef4444;
}

.icon-container.info {
  background: #dbeafe;
  color: #3b82f6;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.modal-body {
  padding: 0 24px 24px 24px;
  text-align: center;
}

.modal-body p {
  margin: 0;
  color: #6b7280;
  line-height: 1.5;
}

.modal-footer {
  padding: 16px 24px 24px 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.cancel-btn, .confirm-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 80px;
  justify-content: center;
}

.cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background: #e5e7eb;
}

.confirm-btn {
  color: white;
}

.confirm-btn.warning {
  background: #f59e0b;
}

.confirm-btn.warning:hover:not(:disabled) {
  background: #d97706;
}

.confirm-btn.danger {
  background: #ef4444;
}

.confirm-btn.danger:hover:not(:disabled) {
  background: #dc2626;
}

.confirm-btn.info {
  background: #3b82f6;
}

.confirm-btn.info:hover:not(:disabled) {
  background: #2563eb;
}

.confirm-btn:disabled, .cancel-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 480px) {
  .modal-footer {
    flex-direction: column;
  }
  
  .cancel-btn, .confirm-btn {
    width: 100%;
  }
}
</style>