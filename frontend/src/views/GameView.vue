<template>
    <section>
        <div class="container">
            <div class="top-item-bar">
                <div>
                    <h2 class="mb-1">{{ board?.title }}</h2>
                    <div class="stats">
                        <span class="stat-item stat-plays">
                            <Play :size="15"/> {{ board?.play_count }}
                        </span>
                        <span class="stat-item stat-likes">
                            <Heart :size="15"/> {{ board?.score }}
                        </span>
                        <span class="stat-item stat-size">
                            <LayoutGrid :size="15"/> {{ board?.size }}x{{ board?.size }}
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
                     :style="`grid-template-columns: repeat(${board?.size ?? 4}, 1fr)`">
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
import { usePageTitle } from '@/composables/usePageTitle'

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

interface GameCell {
    game_cell_id: string
    cell_id: string | null
    content: string
    game_id: string
    is_marked: boolean
    position: number
}

useI18n()
const route = useRoute()
const { pageTitle } = usePageTitle('Bingo Game')

const board = ref<Board | null>(null)
const error = ref<string | null>(null)
const selectedCells = ref<Cell[]>([])

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

// --- Data loading ---
async function loadGame() {
    const gameId = route.params.game_id as string
    error.value = null

    const gameRes = await fetch(`/api/games/${gameId}`, { credentials: 'include' })
    if (!gameRes.ok) { error.value = 'Spiel nicht gefunden'; return }
    const game = await gameRes.json()

    const [boardRes, cellsRes] = await Promise.all([
        fetch(`/api/boards/${game.board_id}`, { credentials: 'include' }),
        fetch(`/api/games/${gameId}/cells`, { credentials: 'include' }),
    ])

    if (!boardRes.ok) { error.value = 'Board nicht gefunden'; return }
    if (!cellsRes.ok) { error.value = 'Zellen nicht gefunden'; return }

    board.value = await boardRes.json()
    if (board.value) pageTitle.value = board.value.title
    const gameCells: GameCell[] = await cellsRes.json()

    selectedCells.value = gameCells
        .sort((a, b) => a.position - b.position)
        .map(gc => ({
            cell_id: gc.game_cell_id,
            content: gc.content,
            value: 0,
        }))

    checkedCells.value = new Set(
        gameCells.filter(gc => gc.is_marked).map(gc => gc.game_cell_id)
    )

    startGame()
}

// --- Game flow ---
async function startGame() {
    if (gameState.value === 'running') return
    const scrollEl = boardContainerRef.value
    if (scrollEl) scrollEl.style.overflowX = 'hidden'
    isRevealing.value = true
    const indices = Array.from({ length: selectedCells.value.length }, (_, i) => i)
        .sort(() => Math.random() - 0.5)
    for (const i of indices) {
        await new Promise(r => setTimeout(r, 80))
        revealedCells.value = new Set(revealedCells.value).add(i)
        await nextTick()
        await new Promise(r => requestAnimationFrame(r))
    }
    await new Promise(r => setTimeout(r, 500))
    if (scrollEl) scrollEl.style.overflowX = ''
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
    loadGame()
})

onUnmounted(() => {
    resizeObserver?.disconnect()
    stopTimer()
})

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
    overflow-x: auto;
    scroll-snap-type: both mandatory;
}

.board-container{
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr;
    min-width: 100%;
}

.board-container div{
    position: relative;
    width: 100%;
    min-width: 100px;
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
    overflow-wrap: break-word; 
    hyphens: auto;
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
