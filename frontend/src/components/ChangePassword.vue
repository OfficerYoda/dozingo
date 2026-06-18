<template>
  <Teleport to="body">
    <div v-if="changePasswordModalOpen" class="modal-overlay" @click.self="close">
      <div class="modal">
        <h2>Passwort ändern</h2>

        <div class="field">
          <label>Aktuelles Passwort</label>
          <input type="password" v-model="oldPassword" @keyup.enter="submit" />
        </div>
        <div class="field">
          <label>Neues Passwort</label>
          <input type="password" v-model="newPassword" @keyup.enter="submit" />
        </div>
        <div class="field">
          <label>Neues Passwort bestätigen</label>
          <input type="password" v-model="confirmPassword" @keyup.enter="submit" />
        </div>

        <p v-if="error" class="error">{{ error }}</p>
        <p v-if="success" class="success">Passwort erfolgreich geändert.</p>

        <div class="actions">
          <button class="btn btn-secondary" @click="close">Abbrechen</button>
          <button class="btn btn-primary" :disabled="loading" @click="submit">
            {{ loading ? 'Wird gespeichert...' : 'Speichern' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useChangePasswordModal } from '@/composables/useChangePasswordModal'
import * as authService from '@/services/auth.services'
import { ApiError } from '@/services/api'

const { changePasswordModalOpen, closeChangePasswordModal } = useChangePasswordModal()

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const success = ref(false)
const loading = ref(false)

watch(changePasswordModalOpen, (open) => {
  if (!open) {
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    error.value = ''
    success.value = false
  }
})

function close() {
  closeChangePasswordModal()
}

async function submit() {
  error.value = ''
  success.value = false

  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Die neuen Passwörter stimmen nicht überein.'
    return
  }
  if (newPassword.value.length < 8) {
    error.value = 'Das Passwort muss mindestens 8 Zeichen lang sein.'
    return
  }

  loading.value = true
  try {
    await authService.changePassword(oldPassword.value, newPassword.value)
    success.value = true
    setTimeout(close, 1500)
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) {
      error.value = 'Das aktuelle Passwort ist falsch.'
    } else {
      error.value = 'Ein Fehler ist aufgetreten. Bitte versuche es erneut.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--color-bg-card-tinted);
  border-radius: var(--radius-sm);
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal h2 {
  margin: 0;
  color: var(--color-heading);
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.field label {
  font-size: 0.875rem;
  color: var(--color-text-subtle);
  font-weight: 500;
}

.field input {
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-interactive-track);
  background: var(--color-bg-surface);
  color: var(--color-heading);
  font-size: 1rem;
}

.error {
  color: var(--color-danger, #e53e3e);
  font-size: 0.875rem;
  margin: 0;
}

.success {
  color: var(--color-success, #38a169);
  font-size: 0.875rem;
  margin: 0;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
</style>