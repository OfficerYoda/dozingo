import { ref } from 'vue'

const changePasswordModalOpen = ref(false)

export function useChangePasswordModal() {
  return {
    changePasswordModalOpen,
    openChangePasswordModal: () => { changePasswordModalOpen.value = true },
    closeChangePasswordModal: () => { changePasswordModalOpen.value = false },
  }
}