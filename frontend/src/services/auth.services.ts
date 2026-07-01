import { apiFetch } from "./api"
import type { User, LoginTwoFAPending } from "./api.type"

export async function login(identifier: string, password: string): Promise<User | LoginTwoFAPending> {
    return apiFetch("/api/auth/login", {
        method: 'POST',
        body: JSON.stringify({ identifier, password }),
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

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
    return apiFetch("/api/auth/change-password", {
        method: 'POST',
        body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    })
}

export async function verifyEmail(token: string): Promise<void> {
    return apiFetch("/api/auth/verify-email", {
        method: 'POST',
        body: JSON.stringify({ token }),
    })
}

export async function newPassword(token: string, newPassword: string): Promise<void> {
    return apiFetch("/api/auth/new-password", {
        method: 'POST',
        body: JSON.stringify({ token, new_password: newPassword }),
    })
}
