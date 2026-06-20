<template>
    <Teleport to="body">
        <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
            <div class="card">
                <div class="header-modal mb-0">
                    <div>
                        <h2 class="mb-0 header-modal-title">{{ board.title }}</h2>
                        <small class="header-modal-subtitle">{{ board.description }}</small>
                    </div>
                    <X :size="20" @click="emit('update:modelValue', false)" />
                </div>

                <hr class="mb-3">

                <p class="confirm-text">Are you sure you want to delete this board? This action cannot be undone.</p>

                <hr class="mb-3">

                <div class="bottom-bar">
                    <div></div>
                    <div class="right-buttons-bottom">
                        <button class="btn btn-secondary button-bottom-row" @click="emit('update:modelValue', false)">
                            <p class="mb-0">Cancel</p>
                        </button>
                        <button class="btn btn-danger button-bottom-row" @click="confirmDelete">
                            <Trash2 :size="20" />
                            <p class="mb-0">Delete</p>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { X, Trash2 } from 'lucide-vue-next'
import * as boardService from '@/services/board.service'
import type { Board } from '@/services/api.type'

const props = defineProps<{
    modelValue: boolean
    board: Board
}>()

const emit = defineEmits<{
    'update:modelValue': [value: boolean]
    'deleted': []
}>()

async function confirmDelete() {
    try {
        await boardService.deleteBoard(props.board.board_id)
        emit('update:modelValue', false)
        emit('deleted')
    } catch { /* ignore */ }
}
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    padding: 16px;
}

.modal-overlay > .card {
    width: 100%;
    max-width: 480px;
    display: flex;
    flex-direction: column;
}

.header-modal {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: start;
}

.header-modal-title {
    color: var(--color-heading);
}

.header-modal-subtitle {
    color: var(--card-blue);
    font-weight: 600;
}

.confirm-text {
    color: var(--color-text-subtle);
    margin: 0;
}

.bottom-bar {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
}

.right-buttons-bottom {
    display: flex;
    flex-direction: row;
    gap: 8px;
}

.button-bottom-row {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 8px;
}
</style>
