<template>
    <section class="reset-password">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">{{ t('resetPassword.title') }}</h2>
                        <small>{{ t('resetPassword.subtitle') }}</small>
                    </div>

                    <div v-if="success" class="status-msg">
                        <CircleCheck :size="32" class="icon-success" />
                        <p>{{ t('resetPassword.success') }}</p>
                    </div>

                    <form v-else @submit.prevent="handleSubmit">
                        <div class="mb-3">
                            <label for="new-password">{{ t('resetPassword.newPassword') }}</label>
                            <div class="input-group">
                                <span><Lock :size="20" /></span>
                                <input v-model="password" type="password" id="new-password"
                                       minlength="8" maxlength="72" required>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label for="confirm-password">{{ t('resetPassword.confirmPassword') }}</label>
                            <div class="input-group">
                                <span><Lock :size="20" /></span>
                                <input v-model="confirmPassword" type="password" id="confirm-password"
                                       minlength="8" maxlength="72" required>
                            </div>
                        </div>
                        <p v-if="error" class="auth-error">{{ error }}</p>
                        <button type="submit" class="btn btn-primary" :disabled="loading">
                            {{ loading ? t('resetPassword.submitting') : t('resetPassword.submit') }}
                        </button>
                    </form>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
    .reset-password .card {
        border-top: 5px solid var(--card-blue);
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
        margin: 0;
        background-color: transparent;
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

    .reset-password button[type="submit"] {
        width: 100%;
        border-radius: var(--radius-lg) !important;
    }

    .status-msg {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 12px;
        padding: 24px 0;
    }

    .icon-success {
        color: #22c55e;
    }

    .auth-error {
        color: var(--color-danger, #e53e3e);
        font-size: 0.85rem;
        margin-bottom: 8px;
    }
</style>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Lock, CircleCheck } from 'lucide-vue-next'
import { newPassword } from '@/services/auth.services'
import { ApiError } from '@/services/api'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const success = ref(false)
const error = ref('')

async function handleSubmit() {
    error.value = ''
    const token = route.query.token as string
    if (!token) {
        error.value = t('resetPassword.noToken')
        return
    }
    if (password.value !== confirmPassword.value) {
        error.value = t('resetPassword.mismatch')
        return
    }
    loading.value = true
    try {
        await newPassword(token, password.value)
        success.value = true
        setTimeout(() => router.push('/'), 2000)
    } catch (e) {
        if (e instanceof ApiError) {
            if (e.status === 410) error.value = t('resetPassword.expired')
            else if (e.status === 404) error.value = t('resetPassword.invalid')
            else error.value = e.message
        } else {
            error.value = t('resetPassword.genericError')
        }
    } finally {
        loading.value = false
    }
}
</script>
