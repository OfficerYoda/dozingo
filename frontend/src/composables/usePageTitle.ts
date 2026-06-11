import { ref } from 'vue'

const pageTitle = ref('')

export function usePageTitle(title?: string) {
    if (title !== undefined) pageTitle.value = title
    return { pageTitle }
}
