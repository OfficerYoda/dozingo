<template>
    <section>
        <div class="container">
            <div class="top-item-bar">
                <div>
                    <h2 class="mb-1">{{ board.title }}</h2>
                    <div class="stats">
                        <span class="stat-item stat-plays">
                            <Play :size="15"/> {{ board.play_count }}
                        </span>
                        <span class="stat-item stat-likes">
                            <Heart :size="15"/> {{ board.score }}
                        </span>
                        <span class="stat-item stat-size">
                            <LayoutGrid :size="15"/> {{ board.size }}x{{ board.size }}
                        </span>
                    </div>
                </div>
                <div>
                    <span class="checked-title">Timer</span>
                    <span class="checked-counter">{{ formattedTime }}</span>
                </div>
                <div>
                    <span class="checked-title">Checked</span>
                    <span class="checked-counter">{{ checkedCells.size }} / {{ selectedCells.length }}</span>
                </div>
            </div>
            
            <div :class="['board', 'mt-3', { stopped: gameState === 'stopped' }]">
                <div class="board-scroll" ref="boardContainerRef" @scroll="updateShadow">
                    <div class="board-container"
                         :style="`grid-template-columns: repeat(${board.size}, 1fr)`">
                        <div v-for="(cell, i) in selectedCells" :key="cell.cell_id"
                             :class="{ revealed: revealedCells.has(i), checked: checkedCells.has(cell.cell_id) }"
                             @click="handleCellClick(cell.cell_id)">
                            {{ cell.content }}
                        </div>
                    </div>
                </div>
                <div class="board-shadow" v-bind:class="(showShadowRight)?'board-shadow-right':''"></div>
                <div class="board-shadow" v-bind:class="(showShadowLeft)?'board-shadow-left':''"></div>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, useTemplateRef, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Heart, LayoutGrid, Play } from 'lucide-vue-next'

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
    value: 0
}

useI18n()
const route = useRoute()

const board = ref<Board>({
    board_id: 'tmp-1',
    title: 'Temporäres Board',
    description: 'Mock-Daten',
    play_count: 42,
    score: 7,
    size: 4,
    author_id: 'user-1',
})

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const cells = ref<Cell[]>([])
const selectedCells = ref<Cell[]>(
    Array.from({ length: board.value.size ** 2 }, (_, i) => ({
        cell_id: `tmp-cell-${i}`,
        content: `Zelle ${i + 1}`,
        value: 0,
    }))
)

// 'stopped' | 'running'
const gameState = ref<'stopped' | 'running'>('stopped')
const revealedCells = ref<Set<number>>(new Set())
const checkedCells = ref<Set<string>>(new Set())
const isRevealing = ref(false)

// --- Timer ---
const elapsedSeconds = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null

const formattedTime = computed(() => {
    const m = Math.floor(elapsedSeconds.value / 60).toString().padStart(2, '0')
    const s = (elapsedSeconds.value % 60).toString().padStart(2, '0')
    return `${m}:${s}`
})

function startTimer() {
    if (timerInterval) return
    timerInterval = setInterval(() => { elapsedSeconds.value++ }, 1000)
}

function stopTimer() {
    if (timerInterval) { clearInterval(timerInterval); timerInterval = null }
}

// --- Game flow ---
async function startGame() {
    if (gameState.value === 'running') return
    isRevealing.value = true
    const indices = Array.from({ length: selectedCells.value.length }, (_, i) => i)
        .sort(() => Math.random() - 0.5)
    for (const i of indices) {
        await new Promise(r => setTimeout(r, 80))
        revealedCells.value = new Set(revealedCells.value).add(i)
        await nextTick()
        await new Promise(r => requestAnimationFrame(r))
    }
    isRevealing.value = false
    gameState.value = 'running'
    startTimer()
}

function resetGame() {
    stopTimer()
    elapsedSeconds.value = 0
    checkedCells.value = new Set()
    revealedCells.value = new Set()
    isRevealing.value = false
    gameState.value = 'stopped'
}

function handleCellClick(cellId: string) {
    if (gameState.value !== 'running') return
    const next = new Set(checkedCells.value)
    next.has(cellId) ? next.delete(cellId) : next.add(cellId)
    checkedCells.value = next
}

// --- Board shadow ---
const boardContainerRef = useTemplateRef<HTMLElement>('boardContainerRef')
const showShadowRight = ref(true)
const showShadowLeft = ref(false)

function updateShadow() {
    const el = boardContainerRef.value
    if (!el) return
    const atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1
    const atStart = el.scrollLeft === 0
    showShadowRight.value = el.scrollWidth > el.clientWidth && !atEnd
    showShadowLeft.value = !atStart
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
    resizeObserver = new ResizeObserver(updateShadow)
    if (boardContainerRef.value) resizeObserver.observe(boardContainerRef.value)
    updateShadow()
    startGame()
})

onUnmounted(() => {
    resizeObserver?.disconnect()
    stopTimer()
})

// --- unused fetch stubs (to be replaced later) ---
const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')
const showModal = ref(false)
const authorName = ref<string | null>(null)

async function fetchAllBoards() {
    const params = new URLSearchParams()
    if (appliedFiler.value) params.set('sort', appliedFiler.value)
    if (search.value) params.set('search', search.value)
    const query = params.toString() ? '?' + params.toString() : ''
    const boardsRes = await fetch('/api/boards' + query, { credentials: 'include' })
    if (!boardsRes.ok) { error.value = 'Failed to load boards'; return }
    boards.value = await boardsRes.json()
}

async function fetchAllCellsForBoard(boardID: string) {
    const cellsRes = await fetch('/api/boards/' + boardID + '/cells')
    if (!cellsRes.ok) { error.value = 'Failed to load cells for board ' + boardID; return }
    cells.value = await cellsRes.json()
    const found = boards.value.find(b => b.board_id === boardID)
    const numberOfCells = (found?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)
    if (found?.author_id) {
        const userRes = await fetch('/api/users/' + found.author_id)
        if (userRes.ok) authorName.value = (await userRes.json()).username
    }
    showModal.value = true
}

</script>

<style scoped>

.top-item-bar{
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.stats {
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
    color: #4052B6;
}

.stat-likes {
    color: #C0185A;
}

.stat-size {
    color: #2E7D32;
}

.checked-title{
    text-align: center;
    font-size: 0.8rem;
    display: block;
}

.checked-counter{
    text-align: center;
    font-weight: 700;
    font-size: 1.2rem;
    display: block;
}

.board{
    background-color: #E3DFFF;
    border-radius: var(--radius-sm);
    padding: 0.5rem;
    position: relative;
}

.board-scroll {
    overflow: auto;
    scroll-snap-type: both mandatory;
}

.board-container{
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr;
    width: max-content;
    min-width: 100%;
}

.board-container div{
    position: relative;
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
    transition: transform 0.6s ease, border-width 0.3s, padding 0.3s;
    transform-style: preserve-3d;
    backface-visibility: hidden;
    will-change: transform;
}

.board-container div::before {
    content: '?';
    color: #fff;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 1rem;
    font-weight: 700;
    position: absolute;
    inset: 0;
    background-color: #5A5781;
    border-radius: var(--radius-lg);
    backface-visibility: hidden;
    transform: perspective(600px) rotateY(180deg);
}

.board.stopped .board-container div:not(.revealed) {
    transform: perspective(600px) rotateY(180deg);
    cursor: default;
    pointer-events: none;
}

.board-container div.revealed {
    transform: perspective(600px) rotateY(0deg);
}
.board-container div:hover:not(.checked):not(.board.stopped *){
    border-width: 0.1rem;
    padding: calc(5px + 0.4rem);
}

.board-shadow{
    position: absolute;
    width: 100%;
    height: 100%;
    left: 0;
    top: 0;
    box-shadow: inset 0 0 0 #E3DFFF;
    pointer-events: none;
    border-radius: var(--radius-sm);
    transition: 0.3s;
}

.board-shadow-right{
    box-shadow: inset -30px 0 30px #E3DFFF;
}

.board-shadow-left{
    box-shadow: inset 30px 0 30px #E3DFFF;
}

.board-container .checked{
    background-color: #5A5781;
    color: #E3DFFF;
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

.board-container .checked::after{
    top: 0;
    right: 0;
    height: 40px;
    width: 40px;
    opacity: 1;
}


</style>
