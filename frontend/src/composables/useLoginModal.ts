import { ref } from 'vue'

const loginModalOpen = ref(false)

export function useLoginModal() {
  return {
    loginModalOpen,
    openLoginModal: () => { loginModalOpen.value = true },
    closeLoginModal: () => { loginModalOpen.value = false },
  }
}
