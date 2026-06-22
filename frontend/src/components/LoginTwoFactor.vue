<template>
    <Teleport to="body">
        <div v-if="loginTwoFactorModalOpen" class="modal-overlay">
            <div class="modal">
                <template v-if="!useRecovery">
                    <h2>{{ $t('settings.security.twoFaVerify.title') }}</h2>
                    <p class="desc">{{ $t('settings.security.twoFaVerify.desc') }}</p>
                    <div class="field">
                        <label for="totp-code">{{ $t('settings.security.twoFaVerify.title') }}</label>
                        <input
                            id="totp-code"
                            v-model="totpCode"
                            type="text"
                            inputmode="numeric"
                            maxlength="6"
                            pattern="\d{6}"
                            autocomplete="one-time-code"
                            placeholder="123456"
                            @keyup.enter="submitTotp"
                        />
                    </div>
                    <p v-if="error" class="error">{{ error }}</p>
                    <div class="actions">
                        <button class="link-btn" @click="useRecovery = true">
                            {{ $t('settings.security.twoFaVerify.useRecovery') }}
                        </button>
                        <button class="btn btn-primary" :disabled="loading" @click="submitTotp">
                            {{ loading ? '...' : $t('settings.security.twoFaVerify.verifyBtn') }}
                        </button>
                    </div>
                </template>
                <template v-else>
                    <h2>{{ $t('settings.security.twoFaVerify.recoveryTitle') }}</h2>
                    <p class="desc">{{ $t('settings.security.twoFaVerify.recoveryDesc') }}</p>
                    <div class="field">
                        <label for="recovery-code">Recovery Code</label>
                        <input
                            id="recovery-code"
                            v-model="recoveryCode"
                            type="text"
                            placeholder="A1B2C3D4-E5F6A7B8"
                            @keyup.enter="submitRecovery"
                        />
                    </div>
                    <p v-if="error" class="error">{{ error }}</p>
                    <div class="actions">
                        <button class="link-btn" @click="useRecovery = false">
                            {{ $t('settings.security.twoFaVerify.backToTotp') }}
                        </button>
                        <button class="btn btn-primary" :disabled="loading" @click="submitRecovery">
                            {{ loading ? '...' : $t('settings.security.twoFaVerify.verifyBtn') }}
                        </button>
                    </div>
                </template>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useLoginTwoFactorModal } from '@/composables/useLoginTwoFactorModal'
import { useAuth } from '@/composables/useAuth'
import * as twoFactorService from '@/services/twoFactor.service'
import { ApiError } from '@/services/api'

const { loginTwoFactorModalOpen, closeLoginTwoFactorModal } = useLoginTwoFactorModal()
const { fetchUser } = useAuth()

const totpCode = ref('')
const recoveryCode = ref('')
const useRecovery = ref(false)
const error = ref('')
const loading = ref(false)

watch(loginTwoFactorModalOpen, (open) => {
    if (!open) {
        totpCode.value = ''
        recoveryCode.value = ''
        useRecovery.value = false
        error.value = ''
    }
})

async function submitTotp() {
    if (totpCode.value.length !== 6) return
    error.value = ''
    loading.value = true
    try {
        await twoFactorService.verify(totpCode.value)
        await fetchUser()
        closeLoginTwoFactorModal()
        window.location.reload()
    } catch (e) {
        if (e instanceof ApiError && e.status === 400) {
            error.value = 'Ungültiger Code. Bitte versuche es erneut.'
        } else if (e instanceof ApiError && e.status === 429) {
            error.value = 'Zu viele Versuche. Bitte warte kurz.'
        } else {
            error.value = 'Ein Fehler ist aufgetreten.'
        }
    } finally {
        loading.value = false
    }
}

async function submitRecovery() {
    if (!recoveryCode.value.trim()) return
    error.value = ''
    loading.value = true
    try {
        await twoFactorService.verifyRecovery(recoveryCode.value.trim().toUpperCase())
        await fetchUser()
        closeLoginTwoFactorModal()
        window.location.reload()
    } catch (e) {
        if (e instanceof ApiError && e.status === 400) {
            error.value = 'Ungültiger Recovery Code.'
        } else if (e instanceof ApiError && e.status === 429) {
            error.value = 'Zu viele Versuche. Bitte warte kurz.'
        } else {
            error.value = 'Ein Fehler ist aufgetreten.'
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
    border-top: 5px solid var(--card-blue);
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

.desc {
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
    letter-spacing: 0.1em;
}

.error {
    color: var(--color-danger, #e53e3e);
    font-size: 0.875rem;
    margin: 0;
}

.actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.5rem;
}

.link-btn {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    font: inherit;
    font-size: 0.85rem;
    color: var(--color-primary-600);
    font-weight: 600;
}
</style>
