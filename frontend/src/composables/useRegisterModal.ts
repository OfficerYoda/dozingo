import { ref } from 'vue'

const registerModalOpen = ref(false)

export function useRegisterModal() {
  return {
    registerModalOpen,
    openRegisterModal: () => { registerModalOpen.value = true },
    closeRegisterModal: () => { registerModalOpen.value = false },
  }
}
