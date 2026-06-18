import { apiFetch } from "./api"
import type { User } from "./api.type"

export async function login(username: string, password: string): Promise<User> {
    return apiFetch("/api/auth/login", {
        method: 'POST',
        body: JSON.stringify({ username, password }),
    })
}

export async function register(username: string, password: string, email?: string): Promise<User> {
    return apiFetch("/api/auth/register", {
        method: 'POST',
        body: JSON.stringify({ username, password, ...(email ? { email } : {}) }),
    })
}

export async function logout(): Promise<void> {
    return apiFetch("/api/auth/logout", { method: 'POST' })
}

export async function getMe(): Promise<User> {
    return apiFetch("/api/users/me")
}

export async function forgotPassword(email: string): Promise<void> {
    return apiFetch("/api/auth/forgot-password", {
        method: 'POST',
        body: JSON.stringify({ email }),
    })
}
