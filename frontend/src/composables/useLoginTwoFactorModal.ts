import { ref } from 'vue'

const loginTwoFactorModalOpen = ref(false)

export function useLoginTwoFactorModal() {
    return {
        loginTwoFactorModalOpen,
        openLoginTwoFactorModal: () => { loginTwoFactorModalOpen.value = true },
        closeLoginTwoFactorModal: () => { loginTwoFactorModalOpen.value = false },
    }
}
