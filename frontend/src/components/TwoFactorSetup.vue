<template>
    <Teleport to="body">
        <div v-if="twoFactorModalOpen" class="modal-overlay" @click.self="close">
            <div class="modal">

                <!-- Schritt 1: QR-Code -->
                <template v-if="step === 1">
                    <h2>{{ $t('settings.security.twoFaSetup.title') }}</h2>
                    <p class="desc">{{ $t('settings.security.twoFaSetup.scanQr') }}</p>
                    <div v-if="loadingSetup" class="loading-box">...</div>
                    <div v-else class="qr-box">
                        <canvas ref="qrCanvas" class="qr-image" />
                        <div class="secret-row">
                            <span class="secret-label">{{ $t('settings.security.twoFaSetup.orEnterSecret') }}</span>
                            <code class="secret-code">{{ secret }}</code>
                        </div>
                    </div>
                    <p v-if="error" class="error">{{ error }}</p>
                    <div class="actions">
                        <button class="btn btn-secondary" @click="close">{{ $t('settings.security.twoFaDisable.cancel') }}</button>
                        <button class="btn btn-primary" :disabled="loadingSetup" @click="step = 2">
                            {{ $t('settings.security.twoFaSetup.next') }}
                        </button>
                    </div>
                </template>

                <!-- Schritt 2: Code bestätigen -->
                <template v-else-if="step === 2">
                    <h2>{{ $t('settings.security.twoFaSetup.confirmTitle') }}</h2>
                    <p class="desc">{{ $t('settings.security.twoFaSetup.confirmDesc') }}</p>
                    <div class="field">
                        <label for="confirm-code">Code</label>
                        <input
                            id="confirm-code"
                            v-model="confirmCode"
                            type="text"
                            inputmode="numeric"
                            maxlength="6"
                            pattern="\d{6}"
                            autocomplete="one-time-code"
                            placeholder="123456"
                            @keyup.enter="submitConfirm"
                        />
                    </div>
                    <p v-if="error" class="error">{{ error }}</p>
                    <div class="actions">
                        <button class="btn btn-secondary" @click="step = 1">Zurück</button>
                        <button class="btn btn-primary" :disabled="loading" @click="submitConfirm">
                            {{ loading ? '...' : $t('settings.security.twoFaSetup.confirmBtn') }}
                        </button>
                    </div>
                </template>

                <!-- Schritt 3: Recovery Codes -->
                <template v-else>
                    <h2>{{ $t('settings.security.twoFaSetup.codesTitle') }}</h2>
                    <p class="desc">{{ $t('settings.security.twoFaSetup.codesDesc') }}</p>
                    <div class="codes-grid">
                        <code v-for="code in recoveryCodes" :key="code" class="recovery-code">{{ code }}</code>
                    </div>
                    <div class="actions">
                        <button class="btn btn-primary" @click="finish">
                            {{ $t('settings.security.twoFaSetup.done') }}
                        </button>
                    </div>
                </template>

            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import QRCode from 'qrcode'
import { useTwoFactorModal } from '@/composables/useTwoFactorModal'
import * as twoFactorService from '@/services/twoFactor.service'
import { ApiError } from '@/services/api'

const emit = defineEmits<{ (e: 'done'): void }>()

const { twoFactorModalOpen, closeTwoFactorModal } = useTwoFactorModal()

const step = ref(1)
const otpAuthUrl = ref('')
const secret = ref('')
const confirmCode = ref('')
const recoveryCodes = ref<string[]>([])
const loadingSetup = ref(false)
const loading = ref(false)
const error = ref('')
const qrCanvas = ref<HTMLCanvasElement | null>(null)

watch(twoFactorModalOpen, async (open) => {
    if (open) {
        step.value = 1
        confirmCode.value = ''
        recoveryCodes.value = []
        error.value = ''
        otpAuthUrl.value = ''
        secret.value = ''
        await runSetup()
    }
})

async function runSetup() {
    loadingSetup.value = true
    error.value = ''
    try {
        const data = await twoFactorService.setup()
        otpAuthUrl.value = data.otp_auth_url
        secret.value = data.secret
        loadingSetup.value = false
        await nextTick()
        if (qrCanvas.value) {
            await QRCode.toCanvas(qrCanvas.value, data.otp_auth_url, { width: 200, margin: 2 })
        }
    } catch (e) {
        loadingSetup.value = false
        if (e instanceof ApiError && e.status === 409) {
            error.value = '2FA ist bereits aktiv.'
        } else {
            error.value = 'Setup fehlgeschlagen. Bitte versuche es erneut.'
        }
    }
}

async function submitConfirm() {
    if (confirmCode.value.length !== 6) return
    error.value = ''
    loading.value = true
    try {
        const data = await twoFactorService.confirm(confirmCode.value)
        recoveryCodes.value = data.recovery_codes
        step.value = 3
    } catch (e) {
        if (e instanceof ApiError && e.status === 400) {
            error.value = 'Ungültiger Code. Bitte versuche es erneut.'
        } else {
            error.value = 'Ein Fehler ist aufgetreten.'
        }
    } finally {
        loading.value = false
    }
}

function close() {
    closeTwoFactorModal()
}

function finish() {
    closeTwoFactorModal()
    emit('done')
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
    max-width: 420px;
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

.loading-box {
    text-align: center;
    padding: 2rem;
    color: var(--color-text-subtle);
}

.qr-box {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
}

.qr-image {
    width: 200px;
    height: 200px;
    border-radius: var(--radius-sm);
}

.secret-row {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    width: 100%;
}

.secret-label {
    font-size: 0.8rem;
    color: var(--color-text-subtle);
}

.secret-code {
    font-family: monospace;
    font-size: 0.85rem;
    background: var(--color-bg-surface);
    padding: 0.4rem 0.75rem;
    border-radius: var(--radius-sm);
    word-break: break-all;
    text-align: center;
    width: 100%;
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
    letter-spacing: 0.15em;
}

.codes-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
}

.recovery-code {
    font-family: monospace;
    font-size: 0.8rem;
    background: var(--color-bg-surface);
    padding: 0.35rem 0.6rem;
    border-radius: var(--radius-sm);
    text-align: center;
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
