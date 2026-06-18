import { reactive, readonly } from 'vue'
import * as authService from '../services/auth.services' 
import type { User } from '../services/api.type'
 
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

async function login(username: string, password: string): Promise<void> {
    state.user = await authService.login(username, password)
}

async function register(username: string, password: string, email?: string): Promise<void> {
    state.user = await authService.register(username, password, email)
}

async function logout(): Promise<void> {
    await authService.logout()
    state.user = null
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
