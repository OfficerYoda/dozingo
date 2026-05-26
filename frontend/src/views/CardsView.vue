<template>
    <section>
        <div class="container">
            <div class="list-header mb-2">
                <h2 class="mb-0">Explore all Cards</h2>
                <div class="header-actions">
                    <input class="btn btn-secondary" type="search" placeholder="Search..">
                    <button class="btn btn-primary">Never played</button>
                    <button class="btn btn-secondary">New</button>
                </div>
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="grid">
                <button
                    v-for="cell in cells"
                    :key="cell.cell_id"
                    class="card card-border-blue col-4 md-6 sm-12"
                >
                    <div class="card-body">
                        <h3>{{ cell.content }}</h3>
                        <small>Value: {{ cell.value }}</small>
                    </div>
                    <hr class="mb-2">
                    <div class="card-footer">
                        <span class="card-meta-text">Board {{ cell.board_id.slice(0, 8) }}</span>
                        <div class="like-group">
                            <Heart :size="20" />
                            <span class="card-meta-text">0</span>
                        </div>
                    </div>
                </button>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Heart } from 'lucide-vue-next'

interface Cell {
    cell_id: string
    board_id: string
    content: string
    author_id: string | null
    value: number
}

interface Board {
    board_id: string
}

useI18n()

const cells = ref<Cell[]>([])
const error = ref<string | null>(null)

async function fetchAllCells() {
    const boardsRes = await fetch('/api/boards', { credentials: 'include' })
    if (!boardsRes.ok) {
        error.value = 'Failed to load boards'
        return
    }
    const boards: Board[] = await boardsRes.json()

    const cellsPerBoard = await Promise.all(
        boards.map(board =>
            fetch(`/api/boards/${board.board_id}/cells`, { credentials: 'include' })
                .then(res => res.ok ? res.json() as Promise<Cell[]> : [])
        )
    )
    cells.value = cellsPerBoard.flat()
}

onMounted(fetchAllCells)
</script>

<style scoped>
/* Header */
.list-header {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.header-actions {
    display: flex;
    flex-direction: row;
    gap: 8px;
}

/* Card body */
.card-body {
    text-align: left;
}

/* Card footer */
.card-footer {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.like-group {
    display: flex;
    flex-direction: row;
    gap: 4px;
}

.like-group svg {
    color: #5A5781;
}

.card-meta-text {
    color: #5A5781;
    font-weight: 600;
    font-size: 13px;
}
</style>
