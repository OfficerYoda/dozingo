<template>
    <section>
        <div class="container">
            <h2 class="mb-1">Board title</h2>
            <small>Subtitle Description</small>
            <div class="modal-stats">
                <span class="stat-item stat-plays">
                    <Play :size="15"/> XX
                </span>
                <span class="stat-item stat-likes">
                    <Heart :size="15"/> XX
                </span>
                <span class="stat-item stat-size">
                    <LayoutGrid :size="15"/> XxX
                </span>
            </div>

            <div class="board mt-3">
                <div class="board-container" ref="boardContainerRef" @scroll="updateShadow" @click="toggleCell">
                    <div>1</div>
                    <div>2</div>
                    <div>3</div>
                    <div>4</div>
                    <div>5</div>
                    <div>6</div>
                    <div>7</div>
                    <div>8</div>
                    <div>9</div>
                    <div>10</div>
                    <div>11</div>
                    <div>12</div>
                    <div>13</div>
                    <div>14</div>
                    <div>15</div>
                    <div>16</div>
                </div>
                <div class="board-shadow" v-show="showShadow"></div>
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
                <div class="header-modal mb-0">
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

                <ul class="background-seperate-cells ">
                    <li v-for="cell in selectedCells" :key="cell.cell_id" class="card cell-btn">
                        <p>{{ cell.content }}</p>
                    </li>
                </ul>

                <hr class="mb-3">

                <div class="bottom-bar">
                    <div class="bottom-bar-text">
                        <small>Createt by</small>
                        <span>{{ authorName ?? '…' }}</span>
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
import { ref, watch, useTemplateRef, onMounted, onUnmounted } from 'vue'
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
    author_id: string
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
const authorName = ref<string | null>(null)

const boardContainerRef = useTemplateRef<HTMLElement>('boardContainerRef')
const showShadow = ref(true)

function updateShadow() {
    const el = boardContainerRef.value
    if (!el) return
    // hide shadow when no horizontal overflow exists or scroll is at the right end
    const atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1
    showShadow.value = el.scrollWidth > el.clientWidth && !atEnd
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
    resizeObserver = new ResizeObserver(updateShadow)
    if (boardContainerRef.value) resizeObserver.observe(boardContainerRef.value)
    updateShadow()
})

onUnmounted(() => {
    resizeObserver?.disconnect()
})

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

    authorName.value = null
    if (selecetedBoard.value?.author_id) {
        const userRes = await fetch('/api/users/' + selecetedBoard.value.author_id)
        if (userRes.ok) {
            const user = await userRes.json()
            authorName.value = user.username
        }
    }

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

function toggleCell(event: MouseEvent) {
    const cell = (event.target as HTMLElement).closest<HTMLElement>('.board-container > div')
    cell?.classList.toggle('marked')
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
    border-block: 5px solid var(--card-blue);
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
















.board{
    background-color: #E3DFFF;
    border-radius: var(--radius-sm);
    padding: 0.5rem;
    position: relative;
}

.board-container{
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr;
    overflow: auto;
    scroll-snap-type: both mandatory;
}

.board-container div{
    width: 100%;
    min-width: 160px;
    height: 100%;
    min-height: 100px;
    scroll-snap-align: start;
    background-color: #fff;
    border: solid 0.5rem #E3DFFF;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 5px;
    text-align: center;
    border-radius: var(--radius-lg);
    cursor: pointer;
    transition: 0.3s;
}

.board-container div:hover:not(.marked){
    box-shadow: 0 0 40px #E3DFFF;
    border-width: 0.1rem;
    padding: calc(5px + 0.4rem);
}

.board-shadow{
    position: absolute;
    width: 100%;
    height: 100%;
    left: 0;
    top: 0;
    box-shadow: inset -30px 0 30px #E3DFFF;
    pointer-events: none;
    border-radius: var(--radius-sm);
}

.board-container .marked{
    background-color: #5A5781;
    color: #E3DFFF;
    position: relative;
}

.board-container div::after{
    content: "✓";
    position: absolute;
    top: 10px;
    right: 10px;
    height: 0;
    width: 0;
    border-radius: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    color: #5A5781;
    background-color: #E3DFFF;
    transform: translate(10%, -10%);
    box-shadow: 0 0 5px #000;
    opacity: 0;
    transition: 0.3s;
    overflow: hidden;
}

.board-container .marked::after{
    top: 0;
    right: 0;
    height: 40px;
    width: 40px;
    opacity: 1;
}

</style>
