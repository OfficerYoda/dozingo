<template>
    <div :class="['board', { stopped: gameState === 'stopped', completed: gameState === 'completed' }]">
        <div :class="['board-scroll', { 'is-revealing': isRevealing }]" ref="boardContainerRef" @scroll="updateShadow">
            <div class="board-container"
             :style="`grid-template-columns: repeat(${board?.size ?? 4}, 1fr)`">
                <button v-for="(cell, i) in selectedCells" :key="cell.cell_id"
                     type="button"
                     :class="{ revealed: revealedCells.has(i), checked: checkedCells.has(cell.cell_id), 'bingo-sweep': sweepingCells.has(cell.cell_id) }"
                     :style="sweepingCells.has(cell.cell_id) ? { '--sweep-delay': sweepingCells.get(cell.cell_id)! * 90 + 'ms' } : {}"
                     :disabled="!revealedCells.has(i) || gameState === 'stopped'"
                     :aria-pressed="checkedCells.has(cell.cell_id)"
                     @click="$emit('cellClick', cell.cell_id)">
                    {{ cell.content }}
                </button>
            </div>
        </div>
        <div class="board-shadow" :class="{ 'board-shadow-right': showShadowRight }"></div>
        <div class="board-shadow" :class="{ 'board-shadow-left': showShadowLeft }"></div>
    </div>
</template>

<script setup lang="ts">
import { ref, useTemplateRef, onMounted, onUnmounted } from 'vue'
import type { Board, Cell } from '@/services/api.type'

defineProps<{
    board: Board | null
    gameState: 'stopped' | 'running' | 'completed'
    selectedCells: Cell[]
    checkedCells: Set<string>
    revealedCells: Set<number>
    isRevealing: boolean
    sweepingCells: Map<string, number>
}>()

defineEmits<{ cellClick: [cellId: string] }>()

const boardContainerRef = useTemplateRef<HTMLElement>('boardContainerRef')
const showShadowRight = ref(false)
const showShadowLeft = ref(false)

function updateShadow() {
    const el = boardContainerRef.value
    if (!el) return
    const atEnd = el.scrollLeft + el.clientWidth >= el.scrollWidth - 1
    const atStart = el.scrollLeft <= 0
    showShadowRight.value = el.scrollWidth > el.clientWidth && !atEnd
    showShadowLeft.value = !atStart
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
</script>

<style scoped>
.board {
    background-color: var(--color-input-bg);
    border-radius: var(--radius-lg);
    padding: 8px;
    position: relative;
}

.board-scroll {
    overflow-x: auto;
    overflow-y: hidden;
}

.board-scroll.is-revealing {
    overflow: hidden;
}

.board-container {
    display: grid;
    min-width: 100%;
    gap: 5px;
}

.board-container button {
    position: relative;
    width: 100%;
    min-width: 130px;
    height: 100%;
    min-height: 110px;
    background-color: var(--game-cell-bg);
    border: 3px solid var(--color-input-bg);
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
    transition: transform 0.6s ease, padding 0.3s, border-color 0.3s, background-color 0.3s, color 0.3s;
    transform-style: preserve-3d;
    backface-visibility: hidden;
    will-change: transform;
}

.board-container button::before {
    content: '';
    color: var(--game-cell-accent-text);
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 1rem;
    font-weight: 700;
    position: absolute;
    inset: 0;
    background-color: var(--game-cell-accent-bg);
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
    border-color: var(--color-primary-500);
}

.board-container button:focus-visible {
    box-shadow: none;
}

.board-shadow {
    position: absolute;
    width: 100%;
    height: 100%;
    left: 0;
    top: 0;
    box-shadow: inset 0 0 0 var(--color-input-bg);
    pointer-events: none;
    border-radius: var(--radius-lg);
    transition: 0.3s;
}

.board-shadow-right {
    box-shadow: inset -30px 0 30px var(--color-input-bg);
}

.board-shadow-left {
    box-shadow: inset 30px 0 30px var(--color-input-bg);
}

.board-container .checked {
    background-color: var(--game-cell-accent-bg);
    color: var(--game-cell-accent-text);
    transform: perspective(600px) rotateX(0deg) !important;
}

.board-container button::after {
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
    color: var(--game-cell-accent-bg);
    background-color: var(--game-cell-accent-text);
    transform: translate(10%, -10%);
    box-shadow: 0 0 5px #000;
    opacity: 0;
    transition: 0.3s;
    overflow: hidden;
}

.board-container .checked::after {
    animation: check-marker 0.3s ease;
    top: 0;
    right: 0;
    height: 30px;
    width: 30px;
    opacity: 1;
}

@keyframes check-marker {
    75% {
        top: 0;
        right: 0;
        height: 40px;
        width: 40px;
        opacity: 1;
    }
}

.board-container button.bingo-sweep {
    animation: bingo-sweep 0.35s ease both;
    animation-delay: var(--sweep-delay, 0ms);
}

@keyframes bingo-sweep {
    0%   { background-color: var(--game-cell-bg); color: inherit; box-shadow: none; }
    40%  { background-color: #4CAF50; color: #fff; box-shadow: 0 0 14px rgba(76, 175, 80, 0.7); }
    100% { background-color: var(--game-cell-bg); color: inherit; box-shadow: none; }
}

.board-container button.checked.bingo-sweep {
    animation: bingo-sweep-checked 0.35s ease both;
    animation-delay: var(--sweep-delay, 0ms);
}

@keyframes bingo-sweep-checked {
    0%   { background-color: var(--game-cell-accent-bg); color: var(--game-cell-accent-text); box-shadow: none; }
    40%  { background-color: #4CAF50; color: #fff; box-shadow: 0 0 14px rgba(76, 175, 80, 0.7); }
    100% { background-color: var(--game-cell-accent-bg); color: var(--game-cell-accent-text); box-shadow: none; }
}
</style>
