<template>
    <section class="pwforgot">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">Stupid?</h2>
                        <small>Reset your password</small>
                    </div>
                    <form @submit.prevent="handlePWRequest">
                        <div class="mb-3">
                            <label for="email">Email</label>
                            <div class="input-group">
                                <span><Mail :size="20" /></span>
                                <input v-model="email" type="email" id="email" required>
                            </div>
                        </div>
                        <p v-if="error" class="auth-error">{{ error }}</p>
                        <p v-if="info" class="auth-info">{{ info }}</p>
                        <button type="submit" class="btn btn-primary" :disabled="loading">
                            {{ loading ? 'Logging in...' : 'Log In' }}
                        </button>
                    </form>
                    <small class="pwforgot-notice">
                        Remember again? <RouterLink to="/login">Log In</RouterLink>
                    </small>
                </div>
            </div>
        </div>
    </section>
</template>

<style>
    .pwforgot .card{
        border-top: 5px solid var(--card-blue);
    }

    .heading{
        text-align: center;
    }

    .heading small{
        color: var(--color-text-subtle);
    }

    .input-group{
        display: flex;
        background-color: var(--color-bg-card-tinted);
        border-radius: var(--radius-sm);
        margin-top: 4px;
    }

    .input-group input{
        border-start-start-radius: 0;
        margin: 0;
    }

    .input-group span{
        display: flex;
        justify-content: center;
        align-items: center;
        padding-inline: 8px;
        color: var(--color-subheading);
    }

    .pwforgot input{
        background-color: var(--color-bg-card-tinted);
        outline: none;
        border: none;
        border-radius: var(--radius-sm);
        padding: 8px 10px;
        display: block;
        width: 100%;
    }

    .password-top{
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .password-top a{
        font-size: 0.7rem;
        font-weight: 700;
        color: var(--color-primary-600);
    }

    .pwforgot button{
        width: 100%;
        border-radius: 999px !important;
    }

    .pwforgot-notice{
        color: var(--color-text-subtle);
        text-align: center;
        display: block;
        padding-top: 20px;
        padding-bottom: 5px;
    }

    .pwforgot-notice a{
        font-weight: 600;
        color: var(--color-primary-600);
    }

    .auth-error{
        color: var(--color-danger, #e53e3e);
        font-size: 0.85rem;
        margin-bottom: 8px;
    }

    .auth-info{
        color: #3498db;
        font-size: 0.85rem;
        margin-bottom: 8px;
    }
</style>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Mail } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { pwRequest } = useAuth()

const email = ref('')
const error = ref('')
const info = ref('')
const loading = ref(false)

async function handlePWRequest() {
    error.value = ''
    info.value = ''
    loading.value = true
    try {
        const status = await pwRequest(email.value)
        if (status !== null) {
            error.value = 'Something went wrong. Please try again.'
            return
        }
        info.value = 'If we find your account, you will receive a mail to reset your password. This can take up to 5 minutes.'
        return;
    } finally {
        loading.value = false
    }
}
</script>