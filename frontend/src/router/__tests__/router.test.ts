import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createRouter, createMemoryHistory } from 'vue-router'
import { defineComponent } from 'vue'

const Stub = defineComponent({ template: '<div/>' })

// Mutable auth state controlled per test
const authState = { user: null as object | null, ready: true }
const fetchUser = vi.fn()
const openLoginModal = vi.fn()

vi.mock('@/composables/useAuth', () => ({
    useAuth: () => ({ state: authState, fetchUser }),
}))
vi.mock('@/composables/useLoginModal', () => ({
    useLoginModal: () => ({ openLoginModal }),
}))

// Import router AFTER mocks are registered
async function buildRouter() {
    vi.resetModules()
    const { default: router } = await import('../index')
    return router
}

beforeEach(() => {
    authState.user = null
    authState.ready = true
    fetchUser.mockReset()
    openLoginModal.mockReset()
})

describe('router guard – requiresAuth', () => {
    it('blocks navigation and opens login modal when not logged in', async () => {
        const router = await buildRouter()
        // Navigate to a requiresAuth route
        const result = await router.push('/profile')
        // Should have been cancelled (returns false in beforeEach guard)
        expect(openLoginModal).toHaveBeenCalledOnce()
        // Current route should NOT be /profile
        expect(router.currentRoute.value.path).not.toBe('/profile')
    })

    it('allows navigation when user is logged in', async () => {
        authState.user = { user_id: '1', username: 'alice' }
        const router = await buildRouter()
        await router.push('/profile')
        expect(router.currentRoute.value.name).toBe('profile')
        expect(openLoginModal).not.toHaveBeenCalled()
    })
})

describe('router guard – guestOnly', () => {
    it('redirects logged-in user to home for guestOnly routes', async () => {
        authState.user = { user_id: '1', username: 'alice' }
        const router = await buildRouter()
        await router.push('/forgotpw')
        expect(router.currentRoute.value.name).toBe('home')
    })

    it('allows anonymous user to access guestOnly routes', async () => {
        authState.user = null
        const router = await buildRouter()
        await router.push('/forgotpw')
        expect(router.currentRoute.value.name).toBe('forgotpw')
    })
})

describe('router guard – fetchUser', () => {
    it('calls fetchUser when state is not ready', async () => {
        authState.ready = false
        fetchUser.mockImplementation(() => {
            authState.ready = true
        })
        const router = await buildRouter()
        await router.push('/')
        expect(fetchUser).toHaveBeenCalledOnce()
    })

    it('skips fetchUser when state is already ready', async () => {
        authState.ready = true
        const router = await buildRouter()
        await router.push('/')
        expect(fetchUser).not.toHaveBeenCalled()
    })
})
