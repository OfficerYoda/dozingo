import { ref } from 'vue'

const twoFactorModalOpen = ref(false)

export function useTwoFactorModal() {
    return {
        twoFactorModalOpen,
        openTwoFactorModal: () => { twoFactorModalOpen.value = true },
        closeTwoFactorModal: () => { twoFactorModalOpen.value = false },
    }
}
