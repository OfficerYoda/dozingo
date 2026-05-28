import { reactive, readonly } from 'vue'

interface User {
    user_id: string
    username: string
    email: string | null
}

const state = reactive<{
    user: User | null
    ready: boolean
}>({
    user: null,
    ready: false,
})

async function fetchUser(): Promise<void> {
    try {
        const res = await fetch('/api/auth/me', { credentials: 'include' })
        state.user = res.ok ? await res.json() : null
    } catch {
        state.user = null
    } finally {
        state.ready = true
    }
}

async function login(username: string, password: string): Promise<number | null> {
    const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username, password }),
    })
    if (!res.ok) return res.status
    state.user = await res.json()
    return null
}

async function register(username: string, password: string, email?: string): Promise<number | null> {
    const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username, password, ...(email ? { email } : {}) }),
    })
    if (!res.ok) return res.status
    state.user = await res.json()
    return null
}

async function logout(): Promise<void> {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' })
    state.user = null
}

async function pwRequest(email: string): Promise<number | null> {
    const res = await fetch('/api/auth/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ email }),
    })
    if (!res.ok) return res.status
    return null
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
