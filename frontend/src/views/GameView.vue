<template>
    <section>
        <div class="container">
            <p v-if="error" class="game-load-error">{{ error }}</p>
            <GameTopBar :board="board" :userVote="userVote" :formattedTime="formattedTime" @likeClick="handleLikeClick" />
            <GameBoard
                class="mt-3"
                :board="board"
                :gameState="gameState"
                :selectedCells="selectedCells"
                :checkedCells="checkedCells"
                :revealedCells="revealedCells"
                :isRevealing="isRevealing"
                :sweepingCells="sweepingCells"
                @cellClick="handleCellClick"
            />
        </div>
        <GameBingoToast :show="bingoToast" :dismissed="bingoModalDismissed" />
        <GameBingoModal v-model="showBingoModal" @continue="onModalContinue" @finish="onModalFinish" />
        <GamePartyOverlay :show="gameState === 'completed'" :formattedTime="formattedTime" :cellCount="selectedCells.length" />
    </section>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageTitle } from '@/composables/usePageTitle'
import { useGameTimer } from '@/composables/useGameTimer'
import { useGameAudio } from '@/composables/useGameAudio'
import * as boardService from '@/services/board.service'
import * as gameService from '@/services/game.service'
import * as voteService from '@/services/vote.service'
import type { Board, Cell, GameCell } from '@/services/api.type'
import { seedCompletedLines as _seedCompletedLines, checkBingo as _checkBingo } from '@/utils/bingo'
import GameTopBar from '@/components/GameTopBar.vue'
import GameBoard from '@/components/GameBoard.vue'
import GameBingoToast from '@/components/GameBingoToast.vue'
import GameBingoModal from '@/components/GameBingoModal.vue'
import GamePartyOverlay from '@/components/GamePartyOverlay.vue'

const route = useRoute()
const router = useRouter()
const { pageTitle } = usePageTitle('Bingo Game')
const { elapsedSeconds, formattedTime, startTimer, stopTimer } = useGameTimer()
const { startTechno, stopTechno } = useGameAudio()

// --- State: board data ---
const board = ref<Board | null>(null)
const selectedCells = ref<Cell[]>([])
const gameId = ref<string>('')
const userVote = ref<number | null>(null)
const error = ref('')

// --- State: game ---
const gameState = ref<'stopped' | 'running' | 'completed'>('stopped')
const revealedCells = ref<Set<number>>(new Set())
const checkedCells = ref<Set<string>>(new Set())
const isRevealing = ref(false)
const completedLines = ref(new Set<string>())

// --- State: bingo / party ---
const bingoToast = ref(false)
let bingoToastTimeout: ReturnType<typeof setTimeout> | null = null
const showBingoModal = ref(false)
const bingoModalDismissed = ref(false)
const sweepingCells = ref(new Map<string, number>())

// === Vote ===

async function loadVote(boardId: string) {
    try {
        const data = await voteService.getBoardVote(boardId)
        userVote.value = data.user_vote
        if (board.value) board.value.score = data.score
    } catch { /* ignore */ }
}

async function handleLikeClick() {
    if (!board.value) return
    const boardId = board.value.board_id
    const wasLiked = userVote.value === 1

    if (wasLiked) {
        userVote.value = null
        board.value.score--
        try {
            await voteService.deleteVote(boardId)
        } catch {
            userVote.value = 1
            board.value.score++
        }
    } else {
        const prev = userVote.value
        userVote.value = 1
        board.value.score += prev === -1 ? 2 : 1
        try {
            await voteService.voteBoard(boardId, 1)
        } catch {
            userVote.value = prev
            board.value.score -= prev === -1 ? 2 : 1
        }
    }
}

// === Game loading ===

async function loadGame() {
    gameId.value = route.params.game_id as string

    let game
    try {
        game = await gameService.getGameById(gameId.value)
    } catch { error.value = 'Spiel nicht gefunden'; return }

    let boardData: Board
    let gameCells: GameCell[]
    try {
        [boardData, gameCells] = await Promise.all([
            boardService.getBoardById(game.board_id),
            gameService.getGameCells(gameId.value),
        ])
    } catch { error.value = 'Daten konnten nicht geladen werden'; return }

    board.value = boardData!
    pageTitle.value = boardData!.title
    loadVote(boardData!.board_id)

    selectedCells.value = gameCells!
        .sort((a, b) => a.position - b.position)
        .map(gc => ({ cell_id: gc.game_cell_id, content: gc.content, value: 0 }))

    checkedCells.value = new Set(
        gameCells!.filter(gc => gc.is_marked).map(gc => gc.game_cell_id)
    )

    seedCompletedLines(selectedCells.value, checkedCells.value, board.value?.size ?? 4)
    if (completedLines.value.size > 0) bingoModalDismissed.value = true

    if (game.status === 'completed') {
        await startGame()
        gameState.value = 'completed'
        stopTimer()
        return
    }

    try {
        await fetch(`/api/games/${gameId.value}/heartbeat`, { method: 'POST', credentials: 'include' })
        const res = await fetch(`/api/stats/playtime/games/${gameId.value}`, { credentials: 'include' })
        if (res.ok) {
            const data = await res.json()
            elapsedSeconds.value = data.total_seconds
        }
    } catch { }

    startGame()
}

// === Game flow ===

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
    startTimer(gameId.value)
}

async function completeGame() {
    if (gameState.value === 'completed') return
    gameState.value = 'completed'
    stopTimer()
    startTechno()
    try {
        await gameService.completeGame(gameId.value)
    } catch {}
}

async function handleCellClick(cellId: string) {
    if (gameState.value !== 'running') return

    const wasChecked = checkedCells.value.has(cellId)
    const nextChecked = !wasChecked

    const next = new Set(checkedCells.value)
    nextChecked ? next.add(cellId) : next.delete(cellId)
    checkedCells.value = next

    try {
        await gameService.markGameCell(gameId.value, cellId, nextChecked)
    } catch {
        const rollback = new Set(checkedCells.value)
        wasChecked ? rollback.add(cellId) : rollback.delete(cellId)
        checkedCells.value = rollback
        return
    }

    if (checkedCells.value.size === selectedCells.value.length && selectedCells.value.length > 0) {
        completeGame()
    } else {
        checkBingo()
    }
}

function onModalContinue() {
    showBingoModal.value = false
    bingoModalDismissed.value = true
}

function onModalFinish() {
    showBingoModal.value = false
    bingoModalDismissed.value = true
    stopTimer()
    gameService.completeGame(gameId.value).catch(() => {})
    router.push('/')
}

// === Bingo detection ===

function seedCompletedLines(cells: Cell[], checked: Set<string>, size: number) {
    _seedCompletedLines(cells, checked, size, completedLines.value)
}

function checkBingo() {
    const size = board.value?.size ?? 4
    const cells = selectedCells.value
    const checked = checkedCells.value

    const { newLines } = _checkBingo(cells, checked, size, completedLines.value)
    if (newLines.length === 0) return

    const next = new Map(sweepingCells.value)
    for (const line of newLines) {
        line.indices.forEach(([r, c], step) => {
            const cellId = cells[r * size + c]?.cell_id
            if (cellId) next.set(cellId, step)
        })
    }
    sweepingCells.value = next

    const duration = (size - 1) * 90 + 400
    setTimeout(() => {
        const cleaned = new Map(sweepingCells.value)
        for (const line of newLines) {
            line.indices.forEach(([r, c]) => {
                const cellId = cells[r * size + c]?.cell_id
                if (cellId) cleaned.delete(cellId)
            })
        }
        sweepingCells.value = cleaned
    }, duration)

    if (bingoToastTimeout) clearTimeout(bingoToastTimeout)
    if (!bingoModalDismissed.value) {
        showBingoModal.value = true
    } else {
        bingoToast.value = true
        bingoToastTimeout = setTimeout(() => { bingoToast.value = false }, 2500)
    }
}

// === Lifecycle ===

onMounted(() => {
    loadGame()
})

onUnmounted(() => {
    stopTimer()
    stopTechno()
    if (bingoToastTimeout) clearTimeout(bingoToastTimeout)
})
</script>

<style scoped>
.game-load-error {
    color: var(--color-danger, #e53e3e);
    padding: 1rem 0;
}
</style>
