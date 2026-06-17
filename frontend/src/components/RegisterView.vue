<template>
    <Teleport to="body">
        <div v-if="registerModalOpen" class="modal-overlay" @click.self="closeRegisterModal">
            <div class="card register-card">
                <div class="heading mb-3">
                    <h2 class="mb-0">Join the Game</h2>
                    <small>Start your journey here at dozingo</small>
                </div>
                <form @submit.prevent="handleRegister">
                    <div class="mb-3">
                        <label for="reg-username">Username</label>
                        <div class="input-group">
                            <span><User :size="20" /></span>
                            <input v-model="username" type="text" id="reg-username" required>
                        </div>
                    </div>
                    <div class="mb-3">
                        <label for="reg-email">Email</label>
                        <div class="input-group">
                            <span><Mail :size="20" /></span>
                            <input v-model="email" type="email" id="reg-email">
                        </div>
                    </div>
                    <div class="mb-3">
                        <label for="reg-password">Password</label>
                        <div class="input-group">
                            <span><KeyRound :size="20" /></span>
                            <input v-model="password" type="password" id="reg-password" required>
                        </div>
                    </div>
                    <div class="mb-3">
                        <label for="reg-confirm-password">Confirm Password</label>
                        <div class="input-group">
                            <span><KeyRound :size="20" /></span>
                            <input v-model="confirmPassword" type="password" id="reg-confirm-password" required>
                        </div>
                    </div>
                    <div class="privacy-consent mb-3">
                        <input type="checkbox" id="reg-privacy" v-model="privacyAccepted" required>
                        <label for="reg-privacy">
                            {{ $t('register.privacyConsentBefore') }}<RouterLink to="/privacy" target="_blank" class="consent-link">{{ $t('register.privacyLink') }}</RouterLink>{{ $t('register.privacyConsentAfter') }}
                        </label>
                    </div>
                    <p v-if="error" class="auth-error">{{ error }}</p>
                    <button type="submit" class="btn btn-primary register-btn" :disabled="loading">
                        {{ loading ? 'Creating account...' : 'Create Account' }}
                    </button>
                </form>
                <small class="login-notice">
                    Already have an account?
                    <button class="link-btn" @click="closeRegisterModal(); openLoginModal()">Log In</button>
                </small>
            </div>
        </div>
    </Teleport>
</template>

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

    .register-card {
        width: 100%;
        max-width: 420px;
        border-top: 5px solid var(--card-blue);
        padding: 32px;
    }

    .heading {
        text-align: center;
    }

    .heading small {
        color: var(--color-text-subtle);
    }

    .input-group {
        display: flex;
        background-color: var(--color-bg-card-tinted);
        border-radius: var(--radius-sm);
        margin-top: 4px;
    }

    .input-group input {
        border-start-start-radius: 0;
        margin: 0;
        background-color: var(--color-bg-card-tinted);
        outline: none;
        border: none;
        border-radius: var(--radius-sm);
        padding: 8px 10px;
        display: block;
        width: 100%;
    }

    .input-group span {
        display: flex;
        justify-content: center;
        align-items: center;
        padding-inline: 8px;
        color: var(--color-subheading);
    }

    .register-btn {
        width: 100%;
        border-radius: 999px !important;
    }

    .login-notice {
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .link-btn {
        background: none;
        border: none;
        padding: 0;
        cursor: pointer;
        font: inherit;
        font-weight: 600;
        color: var(--color-primary-600);
    }

    .auth-error {
        color: var(--color-danger, #e53e3e);
        font-size: 0.85rem;
        margin-bottom: 8px;
    }

    .privacy-consent {
        display: flex;
        align-items: flex-start;
        gap: 10px;
    }

    .privacy-consent input[type="checkbox"] {
        flex-shrink: 0;
        margin-top: 3px;
        width: 16px;
        height: 16px;
        accent-color: var(--color-primary-600);
        cursor: pointer;
    }

    .privacy-consent label {
        font-size: 0.8rem;
        color: var(--color-text-subtle);
        cursor: pointer;
        line-height: 1.5;
    }

    .consent-link {
        color: var(--color-primary-600);
        text-decoration: underline;
    }
</style>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { User, Mail, KeyRound } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import { useRegisterModal } from '@/composables/useRegisterModal'
import { useLoginModal } from '@/composables/useLoginModal'

const router = useRouter()
const { register } = useAuth()
const { registerModalOpen, closeRegisterModal } = useRegisterModal()
const { openLoginModal } = useLoginModal()

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const privacyAccepted = ref(false)
const error = ref('')
const loading = ref(false)

watch(registerModalOpen, (open) => {
    if (open) {
        username.value = ''
        email.value = ''
        password.value = ''
        confirmPassword.value = ''
        privacyAccepted.value = false
        error.value = ''
    }
})

async function handleRegister() {
    error.value = ''
    if (password.value !== confirmPassword.value) {
        error.value = 'Passwords do not match.'
        return
    }
    loading.value = true
    try {
        const status = await register(username.value, password.value, email.value || undefined)
        if (status === 409) {
            error.value = 'Username or email is already taken.'
            return
        }
        if (status !== null) {
            error.value = 'Something went wrong. Please try again.'
            return
        }
        closeRegisterModal()
        router.push('/')
    } finally {
        loading.value = false
    }
}
</script>
