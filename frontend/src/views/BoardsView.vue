<template>
    <section>
        <div class="container">
            <div class="list-header mb-4">
                <h2 class="mb-0">Explore all Cards</h2>
                <div class="header-actions">
                    <input class="btn btn-secondary" type="search" placeholder="Search.." v-model="search">
                    <select class="btn btn-secondary" v-model="appliedFiler">
                        <option value="">No Filter</option>
                        <option value="newest">Newest</option>
                        <option value="most-liked">Most liked</option>
                        <option value="most-played">Most played</option>
                        <option value="oldest">Oldest</option>
                        <option value="least-liked">Least liked</option>
                        <option value="least-played">Least played</option>
                    </select>
                </div>
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="grid">
                <button v-for="board in boards" :key="board.board_id" @click="clickBoard(board.board_id)"
                    class="card card-border-blue col-4 md-6 sm-12">
                    <div class="card-body">
                        <h3>{{ board.title }}</h3>
                        <small>{{ board.description }}</small>
                    </div>
                    <hr class="mb-2">
                    <div class="card-footer">
                        <span class="card-meta-text">Played {{ board.play_count }} times</span>
                        <div class="like-group">
                            <Heart :size="20" />
                            <span class="card-meta-text">{{ board.score }}</span>
                        </div>
                    </div>
                </button>
            </div>
        </div>
    </section>

    <Teleport to="body">
        <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
            <div class="card">
                <div class="header-modal">
                    <div>
                        <h2 class="mb-0 header-modal-title">{{ selecetedBoard?.title }}</h2>
                        <small class="header-modal-subtitle">{{ selecetedBoard?.description }}</small>
                        <div class="modal-stats">
                            <span class="stat-item stat-plays">
                                <Play :size="15"/> {{ selecetedBoard?.play_count }}
                            </span>
                            <span class="stat-item stat-likes">
                                <Heart :size="15"/> {{ selecetedBoard?.score }}
                            </span>
                            <span class="stat-item stat-size">
                                <LayoutGrid :size="15"/> {{ selecetedBoard?.size }}x{{ selecetedBoard?.size }}
                            </span>
                        </div>
                    </div>
                    <X :size="20" @click="showModal = false" />
                </div>

                <hr class="mb-3">


                <ul class="background-seperate-cells">
                    <li v-for="cell in selectedCells" :key="cell.cell_id" class="card cell-btn">
                        <p>{{ cell.content }}</p>
                    </li>
                </ul>

                <hr>

                <div class="bottom-bar">
                    <div class="bottom-bar-text">
                        <small>Createt by</small>
                        <span>Hier Author eintragen</span>
                    </div>
                    <div class="right-buttons-bottom">
                        <button class="btn btn-secondary button-bottom-row" @click="shuffle">
                            <Dices :size="20" />
                            <p class="mb-0">Shuffle</p>
                        </button>
                        <button class="btn btn-primary button-bottom-row">
                            <Play :size="20" />
                            <p class="mb-0">Start the game</p>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Heart, X, Dices, LayoutGrid, Play } from 'lucide-vue-next'

interface Board {
    board_id: string
    title: string
    description: string
    play_count: number
    score: number
    size: number
}

interface Cell {
    cell_id: string,
    content: string,
    value: 0,
}

useI18n()
const route = useRoute()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const cells = ref<Cell[]>([])
const selecetedBoard = ref<Board>()
const selectedCells = ref<Cell[]>()

async function fetchAllBoards() {
    const params = new URLSearchParams()
    if (appliedFiler.value) params.set('sort', appliedFiler.value)
    if (search.value) params.set('search', search.value)

    const query = params.toString() ? '?' + params.toString() : ''
    const boardsRes = await fetch('/api/boards' + query, { credentials: 'include' })
    if (!boardsRes.ok) {
        error.value = 'Failed to load boards'
        return
    }

    boards.value = await boardsRes.json()
}

async function fetchAllCellsForBoard(boardID: string) {
    const cellsRes = await fetch('/api/boards/' + boardID + '/cells')
    if (!cellsRes.ok) {
        error.value = 'Failed to load cells for board ' + boardID
        return
    }

    cells.value = await cellsRes.json()
    selecetedBoard.value = boards.value.find(b => b.board_id === boardID)
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)
    showModal.value = true
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllBoards()
}, { immediate: true })

const showModal = ref(false)

function shuffle() {
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)
}

function clickBoard(boardID: string) {
    console.log("Statet loading the cells for board with boardid " + boardID)
    fetchAllCellsForBoard(boardID)
}


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
    color: #2C2A51;
}

.header-modal-subtitle {
    color: #4052B6;
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
    font-size: 12px;
    font-weight: 600;
}

.stat-plays {
    color: #4052B6;
}

.stat-likes {
    color: #C0185A;
}

.stat-size {
    color: #2E7D32;
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
    background-color: #E3DFFF;
    border-radius: var(--radius-sm);
    padding: 8px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    flex: 1;
    min-height: 0;
    overflow-y: auto;
}

@media (max-width: 600px) {
    .background-seperate-cells {
        grid-template-columns: 1fr;
    }
}

.background-seperate-cells > .card.cell-btn {
    min-height: 0;
    overflow: hidden;
    cursor: pointer;
    width: 100%;
    text-align: center;
    display: flex;
    align-items: center;
    justify-content: center;
}

.cell-detail-card {
    width: 100%;
    max-width: 400px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
}

.cell-detail-header {
    width: 100%;
    display: flex;
    justify-content: flex-end;
    cursor: pointer;
}

.cell-detail-text {
    font-size: 2rem;
    font-weight: 700;
    text-align: center;
    margin: 0;
    color: #2C2A51;
    word-break: break-word;
}

.background-seperate-cells > .cell-btn p {
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
