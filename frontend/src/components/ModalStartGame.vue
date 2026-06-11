<template>
    <Teleport to="body">
        <div v-if="modelValue" class="modal-overlay" @click.self="emit('update:modelValue', false)">
            <div class="card">
                <div class="header-modal mb-0">
                    <div>
                        <h2 class="mb-0 header-modal-title">{{ board.title }}</h2>
                        <small class="header-modal-subtitle">{{ board.description }}</small>
                        <div class="modal-stats">
                            <span class="stat-item stat-plays">
                                <Play :size="15" /> {{ board.play_count }}
                            </span>
                            <span class="stat-item stat-likes">
                                <Heart :size="15" /> {{ board.score }}
                            </span>
                            <span class="stat-item stat-size">
                                <LayoutGrid :size="15" /> {{ board.size }}x{{ board.size }}
                            </span>
                        </div>
                    </div>
                    <X :size="20" @click="emit('update:modelValue', false)" />
                </div>

                <hr class="mb-3">

                <ul class="background-seperate-cells">
                    <li v-for="cell in cells" :key="cell.cell_id" class="card cell-btn">
                        <p>{{ cell.content }}</p>
                    </li>
                </ul>

                <hr class="mb-3">

                <div class="bottom-bar">
                    <div class="bottom-bar-text">
                        <small>{{ t('boards.modal.createdBy') }}</small>
                        <span>{{ authorName ?? '…' }}</span>
                    </div>
                    <div class="right-buttons-bottom">
                        <button class="btn btn-primary button-bottom-row"
                            @click="createGameAndNav()">
                            <Play :size="20" />
                            <p class="mb-0">{{ t('boards.modal.startGame') }}</p>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Heart, X, LayoutGrid, Play } from 'lucide-vue-next'

interface Board {
    board_id: string
    title: string
    description: string
    play_count: number
    score: number
    size: number
    author_id: string
}

interface Cell {
    cell_id: string
    content: string
    value: number
}

const props = defineProps<{
    modelValue: boolean
    board: Board
    cells: Cell[]
    authorName: string | null
}>()

const emit = defineEmits<{
    'update:modelValue': [value: boolean]
}>()

async function createGameAndNav() {
    const cells: Cell[] = [...props.cells]

    shuffle(cells)

    const body = cells.map((cell, index) => ({
        cell_id: cell.cell_id,
        position: index,
    }))

    const createGame = await fetch('/api/boards/' + props.board.board_id + '/games', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(body),
    })

    if (!createGame.ok) return

    const game = await createGame.json()
    router.push('/game/' + game.game_id)
}

function shuffle(array: Cell[]) {
    let currentIndex = array.length;

    while (currentIndex != 0) {

        let randomIndex = Math.floor(Math.random() * currentIndex);
        currentIndex--;

        [array[currentIndex], array[randomIndex]] = [
            array[randomIndex]!, array[currentIndex]!]

    }
}

const { t } = useI18n()
const router = useRouter()
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

.modal-overlay>.card {
    width: 100%;
    max-width: 720px;
    max-height: 90vh;
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

.modal-stats {
    display: flex;
    flex-direction: row;
    gap: 8px;
    margin-top: 6px;
}

.stat-item {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.75rem;
    font-weight: 600;
}

.stat-plays {
    color: var(--card-blue);
}

.stat-likes {
    color: var(--card-red);
}

.stat-size {
    color: var(--card-green);
}

.bottom-bar {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;
}

.bottom-bar-text {
    display: flex;
    flex-direction: column;
}

.button-bottom-row {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 8px;
}

.background-seperate-cells {
    background-color: var(--color-input-bg);
    border-radius: var(--radius-sm);
    padding: 8px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    border-block: 5px solid var(--card-blue);
}

@media (max-width: 600px) {
    .background-seperate-cells {
        grid-template-columns: 1fr;
    }
}

.background-seperate-cells>.card.cell-btn {
    min-height: 0;
    overflow: hidden;
    cursor: pointer;
    width: 100%;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
}

.background-seperate-cells>.cell-btn p {
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    margin: 0;
}

.right-buttons-bottom {
    display: flex;
    flex-direction: row;
    gap: 8px;
}
</style>
