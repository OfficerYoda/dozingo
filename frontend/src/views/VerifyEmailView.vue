<template>
    <section class="verify-email">
        <div class="container">
            <div class="grid">
                <div class="col-4 md-2 sm-0"></div>
                <div class="card col-4 md-8 sm-12">
                    <div class="heading mb-3">
                        <h2 class="mb-0">{{ t('verifyEmail.title') }}</h2>
                    </div>
                    <div v-if="loading" class="status-msg">
                        <Loader2 :size="32" class="spinner" />
                        <p>{{ t('verifyEmail.verifying') }}</p>
                    </div>
                    <div v-else-if="success" class="status-msg">
                        <CircleCheck :size="32" class="icon-success" />
                        <p>{{ t('verifyEmail.success') }}</p>
                    </div>
                    <div v-else class="status-msg">
                        <CircleX :size="32" class="icon-error" />
                        <p class="auth-error">{{ error }}</p>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
    .verify-email .card {
        border-top: 5px solid var(--card-blue);
    }

    .heading {
        text-align: center;
    }

    .status-msg {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 12px;
        padding: 24px 0;
    }

    .spinner {
        animation: spin 1s linear infinite;
        color: var(--color-primary-600);
    }

    @keyframes spin {
        from { transform: rotate(0deg); }
        to { transform: rotate(360deg); }
    }

    .icon-success {
        color: #22c55e;
    }

    .icon-error {
        color: var(--color-danger, #e53e3e);
    }

    .auth-error {
        color: var(--color-danger, #e53e3e);
        font-size: 0.85rem;
    }
</style>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Loader2, CircleCheck, CircleX } from 'lucide-vue-next'
import { verifyEmail } from '@/services/auth.services'
import { ApiError } from '@/services/api'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const loading = ref(true)
const success = ref(false)
const error = ref('')

onMounted(async () => {
    const token = route.query.token as string
    if (!token) {
        loading.value = false
        error.value = t('verifyEmail.noToken')
        return
    }
    try {
        await verifyEmail(token)
        success.value = true
        setTimeout(() => router.push('/'), 2000)
    } catch (e) {
        if (e instanceof ApiError) {
            if (e.status === 410) error.value = t('verifyEmail.expired')
            else if (e.status === 404) error.value = t('verifyEmail.invalid')
            else error.value = e.message
        } else {
            error.value = t('verifyEmail.genericError')
        }
    } finally {
        loading.value = false
    }
})
</script>
