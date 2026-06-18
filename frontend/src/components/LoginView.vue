<template>
    <Teleport to="body">
        <div v-if="loginModalOpen" class="modal-overlay" @click.self="closeLoginModal" @keydown.esc="closeLoginModal">
            <div class="card login-card" ref="loginCardRef" @keydown="handleTabTrap">
                <div class="heading mb-3">
                    <h2 class="mb-0">Welcome Back</h2>
                    <small>Continue your bingo journey</small>
                </div>
                <form @submit.prevent="handleLogin">
                    <div class="mb-3">
                        <label for="username">Username or Email</label>
                        <div class="input-group">
                            <span><User :size="20" /></span>
                            <input v-model="username" type="text" id="username" required tabindex="1">
                        </div>
                    </div>
                    <div class="mb-3">
                        <div class="password-top">
                            <label for="password">Password</label>
                            <a href="/forgotpw" tabindex="5">Forgot Password?</a>
                        </div>
                        <div class="input-group">
                            <span><KeyRound :size="20" /></span>
                            <input v-model="password" type="password" id="password" required tabindex="2">
                        </div>
                    </div>
                    <p v-if="error" class="auth-error">{{ error }}</p>
                    <button type="submit" class="btn btn-primary login-btn" :disabled="loading" tabindex="3">
                        {{ loading ? 'Logging in...' : 'Log In' }}
                    </button>
                </form>
                <small class="register-notice">
                    Don't have an account?
                    <button class="link-btn" @click="closeLoginModal(); openRegisterModal()" tabindex="4">Join the Squad</button>
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

    .login-card {
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

    .password-top {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .password-top a {
        font-size: 0.7rem;
        font-weight: 700;
        color: var(--color-primary-600);
    }

    .login-btn {
        width: 100%;
        border-radius: 999px !important;
    }

    .register-notice {
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .register-notice a {
        font-weight: 600;
        color: var(--color-primary-600);
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
</style>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { User, KeyRound } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import { useLoginModal } from '@/composables/useLoginModal'
import { useRegisterModal } from '@/composables/useRegisterModal'
import { ApiError } from '@/services/api'

const router = useRouter()
const { login } = useAuth()
const { loginModalOpen, closeLoginModal } = useLoginModal()
const { openRegisterModal } = useRegisterModal()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const loginCardRef = ref<HTMLElement | null>(null)

watch(loginModalOpen, (open) => {
    if (open) {
        username.value = ''
        password.value = ''
        error.value = ''
        nextTick(() => {
            loginCardRef.value?.querySelector<HTMLElement>('input, button, a, [tabindex]')?.focus()
        })
    }
})

function getFocusable(): HTMLElement[] {
    const els = Array.from(
        loginCardRef.value?.querySelectorAll<HTMLElement>(
            'a, button, input, textarea, select, [tabindex]:not([tabindex="-1"])'
        ) ?? []
    ).filter(el => !el.hasAttribute('disabled'))

    const withIndex = els.filter(el => (el.tabIndex ?? 0) > 0)
        .sort((a, b) => a.tabIndex - b.tabIndex)
    const withoutIndex = els.filter(el => (el.tabIndex ?? 0) === 0)
    return [...withIndex, ...withoutIndex]
}

function handleTabTrap(e: KeyboardEvent) {
    if (e.key !== 'Tab') return
    const focusable = getFocusable()
    if (!focusable.length) return
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    if (!first || !last) return
    if (e.shiftKey) {
        if (document.activeElement === first) {
            e.preventDefault()
            last.focus()
        }
    } else {
        if (document.activeElement === last) {
            e.preventDefault()
            first.focus()
        }
    }
}

async function handleLogin() {
    error.value = ''
    loading.value = true
    try {
        await login(username.value, password.value)
        closeLoginModal()
        router.push('/')
    } catch (e) {
        if (e instanceof ApiError && e.status === 401) {
            error.value = 'Invalid username or password.'
        } else if (e instanceof ApiError && e.status === 429) {
            error.value = 'Too many attempts. Please wait a moment.'
        } else {
            error.value = 'Something went wrong. Please try again.'
        }
    } finally {
        loading.value = false
    }
}
</script>