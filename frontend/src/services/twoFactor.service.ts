import { apiFetch } from './api'
import type { TwoFASetup, TwoFAConfirm } from './api.type'

export async function setup(): Promise<TwoFASetup> {
    return apiFetch('/api/auth/2fa/setup', { method: 'POST' })
}

export async function confirm(code: string): Promise<TwoFAConfirm> {
    return apiFetch('/api/auth/2fa/confirm', {
        method: 'POST',
        body: JSON.stringify({ code }),
    })
}

export async function verify(code: string): Promise<void> {
    return apiFetch('/api/auth/2fa/verify', {
        method: 'POST',
        body: JSON.stringify({ code }),
    })
}

export async function verifyRecovery(code: string): Promise<void> {
    return apiFetch('/api/auth/2fa/verify-recovery', {
        method: 'POST',
        body: JSON.stringify({ code }),
    })
}

export async function regenerateCodes(
    password: string,
    totpCode?: string,
    recoveryCode?: string,
): Promise<TwoFAConfirm> {
    return apiFetch('/api/auth/2fa/regenerate-codes', {
        method: 'POST',
        body: JSON.stringify({
            password,
            ...(totpCode ? { code: totpCode } : {}),
            ...(recoveryCode ? { recovery_code: recoveryCode } : {}),
        }),
    })
}

export async function disable(
    password: string,
    totpCode?: string,
    recoveryCode?: string,
): Promise<void> {
    return apiFetch('/api/auth/2fa', {
        method: 'DELETE',
        body: JSON.stringify({
            password,
            ...(totpCode ? { code: totpCode } : {}),
            ...(recoveryCode ? { recovery_code: recoveryCode } : {}),
        }),
    })
}
