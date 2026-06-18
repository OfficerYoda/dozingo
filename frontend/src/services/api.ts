export async function apiFetch<T>(url: string, options?: RequestInit): Promise<T>{
    const isFormData = options?.body instanceof FormData

    const res = await fetch(url, {
        ...options,   
        credentials: 'include',
        headers: {
            ...(!isFormData ? { 'Content-Type' : 'application/json' } : {}),
            ...options?.headers,
        
        }
    })

    const body = await res.json()

    if (!res.ok) {
        throw new Error(body.errors?.[0]?.message ?? 'Request failed')
    }

    return body
}