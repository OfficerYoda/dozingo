import { describe, it, expect, vi, beforeEach } from 'vitest'

// useAuth holds module-level reactive state — reset module between tests so
// state doesn't leak across test cases.
beforeEach(() => {
    vi.resetModules()
})

async function setup() {
    // Fresh import after module reset
    const { useAuth } = await import('../useAuth')
    return useAuth()
}

describe('useAuth – fetchUser', () => {
    it('sets user on success', async () => {
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockResolvedValue({ user_id: '1', username: 'alice', email: null, avatar_url: null }),
        }))
        const auth = await setup()
        expect(auth.state.ready).toBe(false)
        await auth.fetchUser()
        expect(auth.state.user).toMatchObject({ user_id: '1', username: 'alice' })
        expect(auth.state.ready).toBe(true)
    })

    it('sets user to null and ready on failure', async () => {
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error('401')),
        }))
        const auth = await setup()
        await auth.fetchUser()
        expect(auth.state.user).toBeNull()
        expect(auth.state.ready).toBe(true)
    })
})

describe('useAuth – login', () => {
    it('sets user on successful non-2FA login', async () => {
        const mockUser = { user_id: '2', username: 'bob', email: null, avatar_url: null }
        const mockLogin = vi.fn().mockResolvedValue(mockUser)
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error()),
            login: mockLogin,
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal: vi.fn() }),
        }))
        const auth = await setup()
        await auth.login('bob', 'pw')
        expect(auth.state.user).toMatchObject({ user_id: '2' })
    })

    it('forwards username identifier to authService.login', async () => {
        const mockLogin = vi.fn().mockResolvedValue({ user_id: '2', username: 'bob', email: null, avatar_url: null })
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error()),
            login: mockLogin,
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal: vi.fn() }),
        }))
        const auth = await setup()
        await auth.login('bob', 'pw')
        expect(mockLogin).toHaveBeenCalledWith('bob', 'pw')
    })

    it('forwards email identifier to authService.login', async () => {
        const mockLogin = vi.fn().mockResolvedValue({ user_id: '2', username: 'bob', email: 'bob@example.com', avatar_url: null })
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error()),
            login: mockLogin,
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal: vi.fn() }),
        }))
        const auth = await setup()
        await auth.login('bob@example.com', 'pw')
        expect(mockLogin).toHaveBeenCalledWith('bob@example.com', 'pw')
    })

    it('opens 2FA modal and does not set user when two_fa_pending', async () => {
        const openLoginTwoFactorModal = vi.fn()
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error()),
            login: vi.fn().mockResolvedValue({ two_fa_pending: true }),
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal }),
        }))
        const auth = await setup()
        await auth.login('bob', 'pw')
        expect(auth.state.user).toBeNull()
        expect(openLoginTwoFactorModal).toHaveBeenCalledOnce()
    })
})

describe('useAuth – logout', () => {
    it('clears user immediately even if server call fails', async () => {
        const mockUser = { user_id: '3', username: 'carol', email: null, avatar_url: null }
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockResolvedValue(mockUser),
            logout: vi.fn().mockRejectedValue(new Error('network')),
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal: vi.fn() }),
        }))
        const auth = await setup()
        await auth.fetchUser()
        expect(auth.state.user).not.toBeNull()

        await auth.logout()
        expect(auth.state.user).toBeNull()
    })
})

describe('useAuth – register', () => {
    it('sets user on successful registration', async () => {
        const mockUser = { user_id: '4', username: 'dave', email: 'dave@x.com', avatar_url: null }
        vi.doMock('@/services/auth.services', () => ({
            getMe: vi.fn().mockRejectedValue(new Error()),
            register: vi.fn().mockResolvedValue(mockUser),
        }))
        vi.doMock('@/composables/useLoginTwoFactorModal', () => ({
            useLoginTwoFactorModal: () => ({ openLoginTwoFactorModal: vi.fn() }),
        }))
        const auth = await setup()
        await auth.register('dave', 'pw', 'dave@x.com')
        expect(auth.state.user).toMatchObject({ user_id: '4' })
    })
})
