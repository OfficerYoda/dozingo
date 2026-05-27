<template>
    <section class="register">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">Join the Game</h2>
                        <small>Start your journey here at dozingo</small>
                    </div>
                    <form @submit.prevent="handleRegister">
                        <div class="mb-3">
                            <label for="username">Username</label>
                            <div class="input-group">
                                <span><User :size="20" /></span>
                                <input v-model="username" type="text" id="username" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="email">Email</label>
                            <div class="input-group">
                                <span><Mail :size="20" /></span>
                                <input v-model="email" type="email" id="email">
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="password">Password</label>
                            <div class="input-group">
                                <span><KeyRound :size="20" /></span>
                                <input v-model="password" type="password" id="password" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="confirm-password">Confirm Password</label>
                            <div class="input-group">
                                <span><KeyRound :size="20" /></span>
                                <input v-model="confirmPassword" type="password" id="confirm-password" required>
                            </div>
                        </div>
                        <p v-if="error" class="auth-error">{{ error }}</p>
                        <button type="submit" class="btn btn-primary" :disabled="loading">
                            {{ loading ? 'Creating account...' : 'Create Account' }}
                        </button>
                    </form>
                    <small class="login-notice">
                        Already have an account? <RouterLink to="/login">Log In</RouterLink>
                    </small>
                </div>
            </div>
        </div>
    </section>
</template>

<style>
    .register .card{
        border-top: 5px solid var(--card-blue);
    }

    .heading{
        text-align: center;
    }

    .heading small{
        color: var(--color-text-subtle);
    }

    .register input{
        background-color: var(--color-bg-card-tinted);
        outline: none;
        border: none;
        border-radius: var(--radius-sm);
        padding: 8px 10px;
        display: block;
        width: 100%;
    }

    .register button{
        width: 100%;
        border-radius: 999px !important;
    }

    .login-notice{
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .login-notice a{
        font-weight: 600;
        color: var(--color-primary-600);
    }

    .auth-error{
        color: var(--color-danger, #e53e3e);
        font-size: 0.85rem;
        margin-bottom: 8px;
    }
</style>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { User, Mail, KeyRound } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { register } = useAuth()

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

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
        router.push('/')
    } finally {
        loading.value = false
    }
}
</script>