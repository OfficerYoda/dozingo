import { apiFetch } from './api'
import type { User, UserSecurity } from './api.type'

export async function getMe(): Promise<User> {
    return apiFetch('/api/users/me')
}

export async function getUserById(userId: string): Promise<User> {
    return apiFetch(`/api/users/${userId}`)
}

export async function getMeSecurity(): Promise<UserSecurity> {
    return apiFetch('/api/users/me/security')
}

export async function updateUsername(username: string): Promise<User> {
    return apiFetch('/api/users/me', {
        method: 'PATCH',
        body: JSON.stringify({ username }),
    })
}

export async function uploadAvatar(file: File): Promise<User> {
    const form = new FormData()
    form.append('avatar', file)
    return apiFetch('/api/users/me/avatar', {
        method: 'PUT',
        body: form,
    })
}

export async function deleteMe(password: string): Promise<void> {
    return apiFetch('/api/users/me', {
        method: 'DELETE',
        body: JSON.stringify({ password }),
    })
}
