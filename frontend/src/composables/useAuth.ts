import { reactive, readonly } from 'vue'
import * as authService from '../services/auth.services'
import type { User } from '../services/api.type'
import { useLoginTwoFactorModal } from './useLoginTwoFactorModal'

const state = reactive<{
    user: User | null
    ready: boolean
}>({
    user: null,
    ready: false,
})

async function fetchUser(): Promise<void> {
    try {
        state.user = await authService.getMe()
    } catch {
        state.user = null
    } finally {
        state.ready = true
    }
}

async function login(identifier: string, password: string): Promise<void> {
    const result = await authService.login(identifier, password)
    if ('two_fa_pending' in result && result.two_fa_pending) {
        const { openLoginTwoFactorModal } = useLoginTwoFactorModal()
        openLoginTwoFactorModal()
        return
    }
    state.user = result as import('../services/api.type').User
}

async function register(username: string, password: string, email?: string): Promise<void> {
    state.user = await authService.register(username, password, email)
}

async function logout(): Promise<void> {
    // UI-State sofort leeren — Reactive Updates triggern jetzt, nicht erst nach Server-Antwort.
    state.user = null
    // Server-Call best-effort: wenn die Session schon abgelaufen ist (401) oder das
    // Backend Probleme hat, soll der User trotzdem ausgeloggt erscheinen.
    try {
        await authService.logout()
    } catch {
        // bewusst ignorieren — lokal ist der User raus, Cookie wird ggf. später invalidiert
    }
}

async function pwRequest(email: string): Promise<void> {
    await authService.forgotPassword(email)
}

export function useAuth() {
    return {
        state: readonly(state),
        fetchUser,
        login,
        register,
        logout,
        pwRequest,
    }
}
