import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiFetch, ApiError } from '../api'

function mockFetch(status: number, body: unknown) {
    return vi.fn().mockResolvedValue({
        ok: status >= 200 && status < 300,
        status,
        json: () => Promise.resolve(body),
    })
}

beforeEach(() => {
    vi.restoreAllMocks()
})

describe('ApiError', () => {
    it('carries status and message', () => {
        const err = new ApiError(404, 'not found')
        expect(err.status).toBe(404)
        expect(err.message).toBe('not found')
        expect(err.name).toBe('ApiError')
        expect(err).toBeInstanceOf(Error)
    })
})

describe('apiFetch', () => {
    it('returns parsed body on 2xx', async () => {
        globalThis.fetch = mockFetch(200, { user_id: '1' })
        const result = await apiFetch('/api/users/me')
        expect(result).toEqual({ user_id: '1' })
    })

    it('always sends credentials: include', async () => {
        globalThis.fetch = mockFetch(200, {})
        await apiFetch('/api/test')
        const [, opts] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
        expect(opts.credentials).toBe('include')
    })

    it('sets Content-Type: application/json for JSON bodies', async () => {
        globalThis.fetch = mockFetch(200, {})
        await apiFetch('/api/test', { method: 'POST', body: JSON.stringify({ x: 1 }) })
        const [, opts] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
        expect(opts.headers['Content-Type']).toBe('application/json')
    })

    it('omits Content-Type header for FormData bodies', async () => {
        globalThis.fetch = mockFetch(200, {})
        await apiFetch('/api/upload', { method: 'POST', body: new FormData() })
        const [, opts] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
        expect(opts.headers['Content-Type']).toBeUndefined()
    })

    it('throws ApiError with errors[0].message when present', async () => {
        globalThis.fetch = mockFetch(422, { errors: [{ message: 'field required' }] })
        await expect(apiFetch('/api/test')).rejects.toMatchObject({
            status: 422,
            message: 'field required',
        })
    })

    it('throws ApiError with detail when no errors array', async () => {
        globalThis.fetch = mockFetch(400, { detail: 'bad input' })
        await expect(apiFetch('/api/test')).rejects.toMatchObject({
            status: 400,
            message: 'bad input',
        })
    })

    it('falls back to "Request failed" when body has no message fields', async () => {
        globalThis.fetch = mockFetch(500, {})
        await expect(apiFetch('/api/test')).rejects.toMatchObject({
            status: 500,
            message: 'Request failed',
        })
    })

    it('falls back gracefully when body is not JSON', async () => {
        globalThis.fetch = vi.fn().mockResolvedValue({
            ok: false,
            status: 502,
            json: () => Promise.reject(new SyntaxError('invalid json')),
        })
        await expect(apiFetch('/api/test')).rejects.toMatchObject({
            status: 502,
            message: 'Request failed',
        })
    })
})
