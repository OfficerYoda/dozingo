<template>
    <Teleport to="body">
        <div v-if="disableTwoFactorModalOpen" class="modal-overlay" @click.self="close">
            <div class="modal">
                <h2>{{ $t('settings.security.twoFaDisable.title') }}</h2>
                <p class="desc">{{ $t('settings.security.twoFaDisable.desc') }}</p>

                <div class="field">
                    <label for="disable-password">{{ $t('settings.security.twoFaDisable.password') }}</label>
                    <input id="disable-password" type="password" v-model="password" @keyup.enter="submit" />
                </div>

                <div class="auth-switch">
                    <label>
                        <input type="radio" v-model="authMethod" value="totp" />
                        {{ $t('settings.security.twoFaDisable.useTotpCode') }}
                    </label>
                    <label>
                        <input type="radio" v-model="authMethod" value="recovery" />
                        {{ $t('settings.security.twoFaDisable.useRecoveryCode') }}
                    </label>
                </div>

                <div v-if="authMethod === 'totp'" class="field">
                    <label for="disable-totp">TOTP-Code</label>
                    <input
                        id="disable-totp"
                        type="text"
                        inputmode="numeric"
                        maxlength="6"
                        v-model="totpCode"
                        placeholder="123456"
                        @keyup.enter="submit"
                    />
                </div>
                <div v-else class="field">
                    <label for="disable-recovery">Recovery Code</label>
                    <input
                        id="disable-recovery"
                        type="text"
                        v-model="recoveryCode"
                        placeholder="A1B2C3D4-E5F6A7B8"
                        @keyup.enter="submit"
                    />
                </div>

                <p v-if="error" class="error">{{ error }}</p>

                <div class="actions">
                    <button class="btn btn-secondary" @click="close">{{ $t('settings.security.twoFaDisable.cancel') }}</button>
                    <button class="btn btn-danger" :disabled="loading" @click="submit">
                        {{ loading ? '...' : $t('settings.security.twoFaDisable.disableBtn') }}
                    </button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import * as twoFactorService from '@/services/twoFactor.service'
import { ApiError } from '@/services/api'

const disableTwoFactorModalOpen = defineModel<boolean>({ required: true })
const emit = defineEmits<{ (e: 'done'): void }>()

const password = ref('')
const totpCode = ref('')
const recoveryCode = ref('')
const authMethod = ref<'totp' | 'recovery'>('totp')
const error = ref('')
const loading = ref(false)

watch(disableTwoFactorModalOpen, (open) => {
    if (!open) {
        password.value = ''
        totpCode.value = ''
        recoveryCode.value = ''
        authMethod.value = 'totp'
        error.value = ''
    }
})

function close() {
    disableTwoFactorModalOpen.value = false
}

async function submit() {
    error.value = ''
    if (!password.value) {
        error.value = 'Passwort ist erforderlich.'
        return
    }
    const code = authMethod.value === 'totp' ? totpCode.value || undefined : undefined
    const recovery = authMethod.value === 'recovery' ? recoveryCode.value.trim().toUpperCase() || undefined : undefined

    loading.value = true
    try {
        await twoFactorService.disable(password.value, code, recovery)
        close()
        emit('done')
    } catch (e) {
        if (e instanceof ApiError && e.status === 401) {
            error.value = 'Falsches Passwort.'
        } else if (e instanceof ApiError && e.status === 400) {
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
    border-top: 5px solid var(--card-red);
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
}

.auth-switch {
    display: flex;
    gap: 1.5rem;
}

.auth-switch label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.875rem;
    color: var(--color-heading);
    cursor: pointer;
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
