<template>
  <Teleport to="body">
    <div v-if="deleteModalOpen" class="modal-overlay" @click.self="closeDeleteModal">
      <div class="modal">
        <h2>Account löschen</h2>
        <p class="warning">Diese Aktion ist <strong>unwiderruflich</strong>. Dein Account und alle zugehörigen Daten werden dauerhaft gelöscht.</p>

        <div class="field">
          <label>Passwort bestätigen</label>
          <input type="password" v-model="password" @keyup.enter="submit" autofocus />
        </div>

        <p v-if="error" class="error">{{ error }}</p>

        <div class="actions">
          <button class="btn btn-secondary" @click="closeDeleteModal">Abbrechen</button>
          <button class="btn btn-danger" :disabled="loading" @click="submit">
            {{ loading ? 'Wird gelöscht...' : 'Account löschen' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useDeleteModal } from '@/composables/useDeleteModal'
import { useAuth } from '@/composables/useAuth'

const { deleteModalOpen, closeDeleteModal } = useDeleteModal()
const { logout } = useAuth()
const router = useRouter()

const password = ref('')
const error = ref('')
const loading = ref(false)

watch(deleteModalOpen, (open) => {
  if (!open) {
    password.value = ''
    error.value = ''
  }
})

async function submit() {
  error.value = ''
  if (!password.value) {
    error.value = 'Bitte Passwort eingeben.'
    return
  }

  loading.value = true
  try {
    const res = await fetch('/api/users/me', {
      method: 'DELETE',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value }),
    })

    if (res.ok) {
      await logout()
      router.push('/')
    } else if (res.status === 401) {
      error.value = 'Falsches Passwort.'
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

.warning {
  color: var(--color-text-subtle);
  font-size: 0.9rem;
  margin: 0;
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

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}
</style>