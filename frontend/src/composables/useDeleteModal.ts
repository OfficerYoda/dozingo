import { ref } from 'vue'

const deleteModalOpen = ref(false)

export function useDeleteModal() {
  return {
    deleteModalOpen,
    openDeleteModal: () => { deleteModalOpen.value = true },
    closeDeleteModal: () => { deleteModalOpen.value = false },
  }
}
