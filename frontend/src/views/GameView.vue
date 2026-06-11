<template>
    <section>
        <div class="container">
            <div class="top-item-bar">
                <div class="top-board-info">
                    <div class="stats">
                        <span class="stat-item stat-plays">
                            <Play :size="13"/> {{ board?.play_count ?? '—' }}
                        </span>
                        <span class="stat-item stat-likes">
                            <Heart :size="13"/> {{ board?.score ?? '—' }}
                        </span>
                        <span class="stat-item stat-size">
                            <LayoutGrid :size="13"/> {{ board?.size ?? '—' }}x{{ board?.size ?? '—' }}
                        </span>
                    </div>
                </div>
                <div class="top-stat-pill">
                    <Timer :size="15" class="top-stat-icon"/>
                    <span class="top-stat-value">{{ formattedTime }}</span>
                </div>
                <div class="top-stat-pill">
                    <CheckSquare :size="15" class="top-stat-icon"/>
                    <span class="top-stat-value">{{ checkedCells.size }}<span class="top-stat-total"> / {{ selectedCells.length }}</span></span>
                </div>
            </div>
            
            <div :class="['board', 'mt-3', { stopped: gameState === 'stopped' }]">
                <div :class="['board-scroll', { 'is-revealing': isRevealing }]" ref="boardContainerRef" @scroll="updateShadow">
                    <div class="board-container"
                     :style="`grid-template-columns: repeat(${board?.size ?? 4}, 1fr)`">
                        <button v-for="(cell, i) in selectedCells" :key="cell.cell_id"
                             type="button"
                             :class="{ revealed: revealedCells.has(i), checked: checkedCells.has(cell.cell_id) }"
                             :disabled="!revealedCells.has(i) || gameState === 'stopped'"
                             :aria-pressed="checkedCells.has(cell.cell_id)"
                             @click="handleCellClick(cell.cell_id)">
                            {{ cell.content }}
                        </button>
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
import { Heart, LayoutGrid, Play, Timer, CheckSquare } from 'lucide-vue-next'
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
    isRevealing.value = true
    const indices = Array.from({ length: selectedCells.value.length }, (_, i) => i)
        .sort(() => Math.random() - 0.5)
    for (const i of indices) {
        await new Promise(r => setTimeout(r, 80))
        revealedCells.value = new Set(revealedCells.value).add(i)
        await nextTick()
        await new Promise(r => requestAnimationFrame(r))
    }
    await new Promise(r => setTimeout(r, 600))
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

.top-item-bar {
    display: flex;
    align-items: center;
    gap: 8px;
}

.top-board-info {
    flex: 1;
}

.stats {
    display: flex;
    flex-direction: row;
    gap: 6px;
    flex-wrap: wrap;
}

.stat-item {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-size: 0.75rem;
    font-weight: 600;
    background-color: var(--color-bg-muted);
    border-radius: var(--radius-sm);
    padding: 3px 8px;
}

.stat-plays { color: #4052B6; }
.stat-likes { color: #C0185A; }
.stat-size  { color: #2E7D32; }

.top-stat-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    background-color: #E3DFFF;
    border-radius: var(--radius-sm);
    padding: 6px 12px;
    white-space: nowrap;
}

.top-stat-icon {
    color: #5A5781;
    flex-shrink: 0;
}

.top-stat-value {
    font-size: 1rem;
    font-weight: 700;
    color: #2C2A51;
    line-height: 1;
}

.top-stat-total {
    font-size: 0.75rem;
    font-weight: 500;
    color: #5A5781;
}

.board{
    background-color: #E3DFFF;
    border-radius: var(--radius-lg);
    padding: 0.5rem;
    position: relative;
}

.board-scroll {
    overflow-x: auto;
    overflow-y: hidden;
    scroll-snap-type: both mandatory;
}

.board-scroll.is-revealing {
    overflow: hidden;
    scroll-snap-type: none;
}

.board-scroll.is-revealing button {
    scroll-snap-align: none;
}

.board-container{
    display: grid;
    grid-template-columns: 1fr 1fr 1fr 1fr;
    min-width: 100%;
    overflow-y: hidden;
}

.board-container button{
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
    font: inherit;
    color: inherit;
    transition: transform 0.6s ease, border-width 0.3s, padding 0.3s, border-color 0.3s;
    transform-style: preserve-3d;
    backface-visibility: hidden;
    will-change: transform;
}

.board-container button::before {
    content: '';
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

.board-container button:disabled {
    transform: perspective(600px) rotateY(180deg);
    cursor: default;
}

.board-container button.revealed {
    transform: perspective(600px) rotateY(0deg);
}

.board-container button:not(:disabled):not(.checked):hover {
    border-color: var(--color-primary-300);
}

.board-container button:focus-visible {
    outline: 3px solid #4052B6;
    outline-offset: -3px;
}

.board-shadow{
    position: absolute;
    width: 100%;
    height: 100%;
    left: 0;
    top: 0;
    box-shadow: inset 0 0 0 #E3DFFF;
    pointer-events: none;
    border-radius: var(--radius-lg);
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
    transform: perspective(600px) rotateX(-360deg) !important;
}

.board-container button::after{
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
