export class ApiError extends Error {
    constructor(public readonly status: number, message: string) {
        super(message)
        this.name = 'ApiError'
    }
}

export async function apiFetch<T>(url: string, options?: RequestInit): Promise<T> {
    const isFormData = options?.body instanceof FormData

    const res = await fetch(url, {
        ...options,
        credentials: 'include',
        headers: {
            ...(!isFormData ? { 'Content-Type': 'application/json' } : {}),
            ...options?.headers,
        }
    })

    const body = await res.json().catch(() => null)

    if (!res.ok) {
        const message = body?.errors?.[0]?.message ?? body?.detail ?? body?.message ?? 'Request failed'
        throw new ApiError(res.status, message)
    }

    return body
}
