<template>
    <section>
        <div class="container">
            <p v-if="error" class="game-load-error">{{ error }}</p>
            <div class="top-item-bar">
                <div class="top-board-info">
                    <div class="stats">
                        <a href="/" class="stat-item back"><ArrowLeft :size="13"/> Back</a>
                        <span class="stat-item stat-plays">
                            <Play :size="13"/> {{ board?.play_count ?? '—' }}
                        </span>
                        <button class="stat-item stat-likes" :class="{ liked: userVote === 1 }" @click="handleLikeClick" type="button" aria-label="Board liken">
                            <Heart :size="13" :fill="userVote === 1 ? 'currentColor' : 'none'"/> {{ board?.score ?? '—' }}
                        </button>
                    </div>
                </div>
                <div class="top-stat-pill">
                    <Timer :size="15" class="top-stat-icon"/>
                    <span class="top-stat-value">{{ formattedTime }}</span>
                </div>
                <button class="top-stat-pill top-action-btn" type="button" @click="shareGame" :aria-label="copyToast ? 'Link kopiert!' : 'Teilen'">
                    <Check v-if="copyToast" :size="15" class="share-check"/>
                    <Share2 v-else :size="15"/>
                </button>
                <button class="top-stat-pill top-action-btn top-fullscreen-btn" type="button" @click="toggleFullscreen" :aria-label="isFullscreen ? 'Fullscreen beenden' : 'Fullscreen'">
                    <Minimize2 v-if="isFullscreen" :size="15"/>
                    <Maximize2 v-else :size="15"/>
                </button>
            </div>
            
            <div :class="['board', 'mt-3', { stopped: gameState === 'stopped', completed: gameState === 'completed' }]">
                <div :class="['board-scroll', { 'is-revealing': isRevealing }]" ref="boardContainerRef" @scroll="updateShadow">
                    <div class="board-container"
                     :style="`grid-template-columns: repeat(${board?.size ?? 4}, 1fr)`">
                        <button v-for="(cell, i) in selectedCells" :key="cell.cell_id"
                             type="button"
                             :class="{ revealed: revealedCells.has(i), checked: checkedCells.has(cell.cell_id), 'bingo-sweep': sweepingCells.has(cell.cell_id) }"
                             :style="sweepingCells.has(cell.cell_id) ? { '--sweep-delay': sweepingCells.get(cell.cell_id)! * 90 + 'ms' } : {}"
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

        <Transition name="bingo-toast">
            <div v-if="bingoToast && bingoModalDismissed" class="bingo-toast">
                <span class="toast-cannon toast-cannon-left">
                    <span v-for="n in 12" :key="n" class="cannon-piece"
                          :style="{
                              '--angle': (180 + n * 15) + 'deg',
                              '--dist': (60 + (n % 4) * 30) + 'px',
                              '--delay': ((n % 5) * 0.06) + 's',
                              backgroundColor: confettiColors[n % confettiColors.length],
                          }"></span>
                </span>
                🎯 BINGO!
                <span class="toast-cannon toast-cannon-right">
                    <span v-for="n in 12" :key="n" class="cannon-piece"
                          :style="{
                              '--angle': (n * 15) + 'deg',
                              '--dist': (60 + (n % 4) * 30) + 'px',
                              '--delay': ((n % 5) * 0.06) + 's',
                              backgroundColor: confettiColors[(n + 3) % confettiColors.length],
                          }"></span>
                </span>
            </div>
        </Transition>

        <Teleport to="body">
            <Transition name="bingo-modal">
                <div v-if="showBingoModal" class="bingo-modal-overlay">
                    <div
                        v-for="p in bingoParticles" :key="p.id"
                        class="bingo-particle"
                        :style="p.style"
                    />
                    <div class="bingo-modal">
                        <p class="bingo-modal-title">BINGO!</p>
                        <p class="bingo-modal-sub">Du hast eine Reihe vervollständigt!</p>
                        <div class="bingo-modal-actions">
                            <button class="btn btn-secondary" @click="continuePlaying">Weiterspielen</button>
                            <button class="btn btn-primary" @click="finishGame">Fertig!</button>
                        </div>
                    </div>
                </div>
            </Transition>
        </Teleport>

        <Teleport to="body">
            <div v-if="gameState === 'completed'" class="party-overlay" @click="$router.push('/')">
                <div class="party-beat party-beat-1"></div>
                <div class="party-beat party-beat-2"></div>
                <div class="party-beat party-beat-3"></div>

                <div class="party-confetti">
                    <span v-for="n in 100" :key="'b' + n" class="bingo-ball"
                          :style="{
                              left: ((n * 1.05) % 100) + '%',
                              animationDelay: ((n % 17) * 0.15) + 's',
                              animationDuration: (1.8 + (n % 5) * 0.3) + 's',
                              backgroundColor: confettiColors[n % confettiColors.length],
                          }">{{ ((n * 7) % 75) + 1 }}</span>

                    <span v-for="n in 60" :key="'c' + n" class="party-card"
                          :style="{
                              left: ((n * 1.7) % 100) + '%',
                              animationDelay: ((n % 13) * 0.2) + 's',
                              animationDuration: (2.2 + (n % 4) * 0.4) + 's',
                              transform: `rotate(${(n * 53) % 360}deg)`,
                          }"></span>

                    <span v-for="n in 40" :key="'d' + n" class="party-falling-dice"
                          :style="{
                              left: ((n * 2.6) % 100) + '%',
                              animationDelay: ((n % 11) * 0.25) + 's',
                              animationDuration: (2 + (n % 3) * 0.5) + 's',
                          }">
                        <span class="dot" v-for="i in 9" :key="i"
                              :style="{ visibility: [0,2,4,6,8].includes(i - 1) ? 'visible' : 'hidden' }"></span>
                    </span>

                    <span v-for="n in 50" :key="'s' + n" class="party-star"
                          :style="{
                              left: ((n * 2.05) % 100) + '%',
                              animationDelay: ((n % 15) * 0.18) + 's',
                              animationDuration: (1.6 + (n % 4) * 0.35) + 's',
                          }">
                        <Star :size="24" fill="#F79F1F"/>
                    </span>

                    <span v-for="(emoji, n) in partyEmojis" :key="'e' + n" class="party-emoji"
                          :style="{
                              left: ((n * 4.7) % 100) + '%',
                              animationDelay: ((n % 9) * 0.3) + 's',
                              animationDuration: (2.5 + (n % 5) * 0.4) + 's',
                          }">{{ emoji }}</span>
                </div>

                <div class="party-banner">
                    <div class="banner-burst">
                        <span v-for="n in 60" :key="'burst' + n" class="burst-piece"
                              :style="{
                                  '--angle': ((n * 360) / 60) + 'deg',
                                  '--distance': (180 + (n % 5) * 60) + 'px',
                                  '--delay': ((n % 7) * 0.08) + 's',
                                  '--duration': (1 + (n % 4) * 0.3) + 's',
                                  backgroundColor: confettiColors[n % confettiColors.length],
                              }"></span>
                    </div>
                    <div class="party-dice-row">
                        <Dices :size="56" class="party-dice party-dice-1"/>
                        <Sparkles :size="40" class="party-sparkle"/>
                        <Dices :size="56" class="party-dice party-dice-2"/>
                    </div>
                    <h1 class="party-title">DOZINGO!</h1>
                    <p class="party-subtitle">{{ formattedTime }} · alle {{ selectedCells.length }} Felder geschafft</p>
                    <RouterLink to="/" class="btn btn-primary mt-3">Zur Startseite</RouterLink>
                </div>
            </div>
        </Teleport>
    </section>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, useTemplateRef, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Heart, Play, Timer, Sparkles, Dices, Star, ArrowLeft, Maximize2, Minimize2, Share2, Check } from 'lucide-vue-next'
import { usePageTitle } from '@/composables/usePageTitle'
import * as boardService from '@/services/board.service'
import * as gameService from '@/services/game.service'
import * as voteService from '@/services/vote.service'
import type { Board, Cell, GameCell } from '@/services/api.type'

useI18n()
const route = useRoute()
const router = useRouter()
const { pageTitle } = usePageTitle('Bingo Game')

const board = ref<Board | null>(null)
const selectedCells = ref<Cell[]>([])
const gameId = ref<string>('')
const userVote = ref<number | null>(null)
const error = ref('')

// 'stopped' | 'running' | 'completed'
const gameState = ref<'stopped' | 'running' | 'completed'>('stopped')
const revealedCells = ref<Set<number>>(new Set())
const checkedCells = ref<Set<string>>(new Set())
const isRevealing = ref(false)
const bingoToast = ref(false)
const showBingoModal = ref(false)
const bingoModalDismissed = ref(false)
const bingoParticles = ref<Array<{ id: number; style: Record<string, string> }>>([])

watch(showBingoModal, (val) => {
    if (!val) { bingoParticles.value = []; return }
    bingoParticles.value = Array.from({ length: 80 }, (_, i) => ({
        id: i,
        style: {
            left: `${(i / 80) * 100 + (Math.sin(i * 2.3) * 6)}%`,
            top: `-${10 + (i % 3) * 8}px`,
            background: confettiColors[i % confettiColors.length]!,
            width: `${6 + (i % 3) * 4}px`,
            height: `${6 + (i % 5) * 3}px`,
            'border-radius': i % 3 === 0 ? '50%' : '2px',
            'animation-delay': `${(i % 10) * 0.08}s`,
            'animation-duration': `${1.4 + (i % 7) * 0.2}s`,
        },
    }))
})
const completedLines = ref(new Set<string>())

// --- Fullscreen ---
const isFullscreen = ref(!!document.fullscreenElement)

function toggleFullscreen() {
    if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen()
    } else {
        document.exitFullscreen()
    }
}

function onFullscreenChange() {
    isFullscreen.value = !!document.fullscreenElement
}

const copyToast = ref(false)

function shareGame() {
    const url = window.location.href
    const title = board.value?.title ?? 'Dozingo'
    if (!navigator.share) {
        navigator.share({ title, url }).catch(() => {})
    } else {
        navigator.clipboard.writeText(url).then(() => {
            copyToast.value = true
            setTimeout(() => { copyToast.value = false }, 2000)
        }).catch(() => {})
    }
}


const sweepingCells = ref(new Map<string, number>())
let bingoToastTimeout: ReturnType<typeof setTimeout> | null = null
const confettiColors = ['var(--color-heading)', '#C0185A', '#2E7D32', '#F79F1F', 'var(--color-subheading)', 'var(--color-input-bg)', '#EA2027']
const partyEmojis = ['🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯', '🍾', '🎈', '💫', '🔥', '🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯']

// --- Techno-Beat (Web Audio, full-bar pre-scheduling) ---
let audioCtx: AudioContext | null = null
let beatScheduler: ReturnType<typeof setTimeout> | null = null

const BPM = 130
const STEP = 60 / BPM / 4          // 16tel in Sekunden ≈ 0.115s
const BAR  = STEP * 16              // eine Bar ≈ 1.846s

function startTechno() {
    if (audioCtx) return
    const Ctx = window.AudioContext ?? (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    audioCtx = new Ctx()
    scheduleBar(audioCtx.currentTime + 0.05)
}

function stopTechno() {
    if (beatScheduler) { clearTimeout(beatScheduler); beatScheduler = null }
    if (audioCtx) { audioCtx.close(); audioCtx = null }
}

function scheduleBar(barStart: number) {
    if (!audioCtx) return
    for (let step = 0; step < 16; step++) {
        playStep(step, barStart + step * STEP)
    }
    const msUntilNext = (barStart + BAR - audioCtx.currentTime - 0.1) * 1000
    beatScheduler = setTimeout(() => scheduleBar(barStart + BAR), Math.max(0, msUntilNext))
}

function playStep(beat: number, t: number) {
    if (!audioCtx) return
    if (beat % 4 === 0)              playKick(t)
    if (beat === 4 || beat === 12)   playClap(t)
    if (beat % 2 === 0)              playHat(t, beat % 4 === 2)
    const stabPattern: Record<number, number> = { 0: 55, 3: 62, 6: 69, 7: 55, 10: 62, 14: 73 }
    const stab = stabPattern[beat]
    if (stab !== undefined)          playStab(t, stab)
}

function playKick(t: number) {
    if (!audioCtx) return
    const osc = audioCtx.createOscillator()
    const gain = audioCtx.createGain()
    osc.frequency.setValueAtTime(150, t)
    osc.frequency.exponentialRampToValueAtTime(40, t + 0.15)
    gain.gain.setValueAtTime(1.2, t)
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.25)
    osc.connect(gain).connect(audioCtx.destination)
    osc.start(t); osc.stop(t + 0.3)
}

function playClap(t: number) {
    if (!audioCtx) return
    const buffer = audioCtx.createBuffer(1, audioCtx.sampleRate * 0.15, audioCtx.sampleRate)
    const data = buffer.getChannelData(0)
    for (let i = 0; i < data.length; i++) data[i] = (Math.random() * 2 - 1) * Math.exp(-i / (audioCtx.sampleRate * 0.04))
    const src = audioCtx.createBufferSource()
    src.buffer = buffer
    const filter = audioCtx.createBiquadFilter()
    filter.type = 'highpass'
    filter.frequency.value = 1500
    const gain = audioCtx.createGain()
    gain.gain.value = 0.5
    src.connect(filter).connect(gain).connect(audioCtx.destination)
    src.start(t)
}

function playHat(t: number, accent: boolean) {
    if (!audioCtx) return
    const buffer = audioCtx.createBuffer(1, audioCtx.sampleRate * 0.05, audioCtx.sampleRate)
    const data = buffer.getChannelData(0)
    for (let i = 0; i < data.length; i++) data[i] = (Math.random() * 2 - 1) * Math.exp(-i / (audioCtx.sampleRate * 0.01))
    const src = audioCtx.createBufferSource()
    src.buffer = buffer
    const filter = audioCtx.createBiquadFilter()
    filter.type = 'highpass'
    filter.frequency.value = 7000
    const gain = audioCtx.createGain()
    gain.gain.value = accent ? 0.25 : 0.12
    src.connect(filter).connect(gain).connect(audioCtx.destination)
    src.start(t)
}

function playStab(t: number, freq: number) {
    if (!audioCtx) return
    const osc = audioCtx.createOscillator()
    osc.type = 'sawtooth'
    osc.frequency.value = freq
    const filter = audioCtx.createBiquadFilter()
    filter.type = 'lowpass'
    filter.Q.value = 12
    filter.frequency.setValueAtTime(2000, t)
    filter.frequency.exponentialRampToValueAtTime(300, t + 0.18)
    const gain = audioCtx.createGain()
    gain.gain.setValueAtTime(0.35, t)
    gain.gain.exponentialRampToValueAtTime(0.001, t + 0.18)
    osc.connect(filter).connect(gain).connect(audioCtx.destination)
    osc.start(t); osc.stop(t + 0.2)
}

// --- Timer ---
const elapsedSeconds = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null
let heartbeatInterval: ReturnType<typeof setInterval> | null = null

const formattedTime = computed(() => {
    const m = Math.floor(elapsedSeconds.value / 60).toString().padStart(2, '0')
    const s = (elapsedSeconds.value % 60).toString().padStart(2, '0')
    return `${m}:${s}`
})

function startTimer() {
    if (timerInterval) return
    timerInterval = setInterval(() => { elapsedSeconds.value++ }, 1000)
    heartbeatInterval = setInterval(() => {
        fetch(`/api/games/${gameId.value}/heartbeat`, { method: 'POST', credentials: 'include' }).catch(() => {})
    }, 30000)
}

function stopTimer() {
    if (timerInterval) { clearInterval(timerInterval); timerInterval = null }
    if (heartbeatInterval) { clearInterval(heartbeatInterval); heartbeatInterval = null }
}

// --- Vote ---
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

// --- Data loading ---
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
        .map(gc => ({
            cell_id: gc.game_cell_id,
            content: gc.content,
            value: 0,
        }))

    checkedCells.value = new Set(
        gameCells!.filter(gc => gc.is_marked).map(gc => gc.game_cell_id)
    )

    // Bereits abgeschlossene Linien still eintragen (kein Toast/Modal)
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

async function completeGame() {
    if (gameState.value === 'completed') return
    gameState.value = 'completed'
    stopTimer()
    startTechno()
    try {
        await gameService.completeGame(gameId.value)
    } catch {
        // best-effort: UI bleibt locked auch bei Netzwerk-Fehler
    }
}

async function handleCellClick(cellId: string) {
    if (gameState.value !== 'running') return

    const wasChecked = checkedCells.value.has(cellId)
    const nextChecked = !wasChecked

    // optimistic update
    const next = new Set(checkedCells.value)
    nextChecked ? next.add(cellId) : next.delete(cellId)
    checkedCells.value = next

    try {
        await gameService.markGameCell(gameId.value, cellId, nextChecked)
    } catch {
        // rollback on failure
        const rollback = new Set(checkedCells.value)
        wasChecked ? rollback.add(cellId) : rollback.delete(cellId)
        checkedCells.value = rollback
        return
    }

    // alle Zellen abgehakt? → Spiel abschließen
    if (checkedCells.value.size === selectedCells.value.length && selectedCells.value.length > 0) {
        completeGame()
    } else {
        checkBingo()
    }
}

function continuePlaying() {
    showBingoModal.value = false
    bingoModalDismissed.value = true
}

function finishGame() {
    showBingoModal.value = false
    bingoModalDismissed.value = true
    stopTimer()
    gameService.completeGame(gameId.value).catch(() => {})
    router.push('/')
}

function seedCompletedLines(cells: Cell[], checked: Set<string>, size: number) {
    const isChecked = (r: number, c: number) => checked.has(cells[r * size + c]?.cell_id ?? '')
    for (let r = 0; r < size; r++)
        if (Array.from({ length: size }, (_, c) => isChecked(r, c)).every(Boolean))
            completedLines.value.add(`row${r}`)
    for (let c = 0; c < size; c++)
        if (Array.from({ length: size }, (_, r) => isChecked(r, c)).every(Boolean))
            completedLines.value.add(`col${c}`)
    if (Array.from({ length: size }, (_, i) => isChecked(i, i)).every(Boolean))
        completedLines.value.add('diag0')
    if (Array.from({ length: size }, (_, i) => isChecked(i, size - 1 - i)).every(Boolean))
        completedLines.value.add('diag1')
}

function checkBingo() {
    const size = board.value?.size ?? 4
    const cells = selectedCells.value
    const checked = checkedCells.value

    const isChecked = (row: number, col: number) =>
        checked.has(cells[row * size + col]?.cell_id ?? '')

    const lines: Array<{ key: string; indices: [number, number][] }> = []

    for (let r = 0; r < size; r++) {
        lines.push({ key: `row${r}`, indices: Array.from({ length: size }, (_, c) => [r, c]) })
    }
    for (let c = 0; c < size; c++) {
        lines.push({ key: `col${c}`, indices: Array.from({ length: size }, (_, r) => [r, c]) })
    }
    lines.push({ key: 'diag0', indices: Array.from({ length: size }, (_, i) => [i, i]) })
    lines.push({ key: 'diag1', indices: Array.from({ length: size }, (_, i) => [i, size - 1 - i]) })

    // Linien die jetzt nicht mehr komplett sind aus completedLines entfernen
    for (const line of lines) {
        if (completedLines.value.has(line.key) && !line.indices.every(([r, c]) => isChecked(r, c))) {
            completedLines.value.delete(line.key)
        }
    }

    let newBingo = false
    const newLineIndices: [number, number][][] = []
    for (const line of lines) {
        if (completedLines.value.has(line.key)) continue
        if (line.indices.every(([r, c]) => isChecked(r, c))) {
            completedLines.value.add(line.key)
            newBingo = true
            newLineIndices.push(line.indices)
        }
    }

    if (newBingo) {
        // Lauflicht: jede neue Linie sequenziell durchleuchten
        const next = new Map(sweepingCells.value)
        for (const indices of newLineIndices) {
            indices.forEach(([r, c], step) => {
                const cellId = cells[r * size + c]?.cell_id
                if (cellId) next.set(cellId, step)
            })
        }
        sweepingCells.value = next
        const duration = (size - 1) * 90 + 400 // letzter delay + animationsdauer
        setTimeout(() => {
            const cleaned = new Map(sweepingCells.value)
            for (const indices of newLineIndices) {
                indices.forEach(([r, c]) => {
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
    document.addEventListener('fullscreenchange', onFullscreenChange)
})

onUnmounted(() => {
    resizeObserver?.disconnect()
    stopTimer()
    stopTechno()
    if (bingoToastTimeout) clearTimeout(bingoToastTimeout)
    document.removeEventListener('fullscreenchange', onFullscreenChange)
})

</script>

<style scoped>

.game-load-error {
    color: var(--color-danger, #e53e3e);
    padding: 1rem 0;
}

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
    gap: 6px;
    font-size: 0.75rem;
    font-weight: 700;
    background-color: var(--color-input-bg);
    border-radius: var(--radius-sm);
    padding: 6px 12px;
    white-space: nowrap;
    text-decoration: none;
    color: var(--color-subheading);
}

.stat-item.back {
    color: var(--color-heading);
    border: none;
    cursor: pointer;
    transition: background-color 0.2s, color 0.2s;
}

.stat-item.back:hover {
    background-color: var(--color-interactive-track);
    color: var(--color-heading);
}

.stat-plays { color: var(--color-heading); }
.stat-likes {
    color: #C0185A;
    cursor: pointer;
    border: none;
    transition: background-color 0.2s;
}
.stat-likes:hover { background-color: var(--color-interactive-track); }
.stat-likes.liked { background-color: color-mix(in srgb, #C0185A 15%, var(--color-input-bg)); color: #C0185A; }

.top-stat-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    background-color: var(--color-input-bg);
    border-radius: var(--radius-sm);
    padding: 6px 12px;
    white-space: nowrap;
}

.top-action-btn {
    border: none;
    cursor: pointer;
    color: var(--color-subheading);
    padding: 6px 10px;
    transition: background-color 0.2s, color 0.2s;
}

.top-action-btn:hover {
    background-color: var(--color-interactive-track);
    color: var(--color-heading);
}

.top-action-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
}

.share-check {
    color: #2E7D32;
}

.top-fullscreen-btn {
    border: none;
    cursor: pointer;
    color: var(--color-subheading);
    padding: 6px 10px;
    transition: background-color 0.2s, color 0.2s;
}

.top-fullscreen-btn:hover {
    background-color: var(--color-interactive-track);
    color: var(--color-heading);
}

@media (max-width: 600px) {
    .top-fullscreen-btn {
        display: none;
    }
}

.top-stat-icon {
    color: var(--color-subheading);
    flex-shrink: 0;
}

.top-stat-value {
    font-size: 1rem;
    font-weight: 700;
    color: var(--color-heading);
    line-height: 1;
}

.top-stat-total {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-subheading);
}

.board{
    background-color: var(--color-input-bg);
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
    background-color: var(--game-cell-bg);
    border: solid 0.5rem var(--color-input-bg);
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
    transition: transform 0.6s ease, border-width 0.3s, padding 0.3s, border-color 0.3s, background-color 0.3s, color 0.3s;
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
    border-color: var(--color-primary-300);
}

.board-container button:focus-visible {
    outline: 3px solid var(--color-heading);
    outline-offset: -3px;
}

.board-shadow{
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

.board-shadow-right{
    box-shadow: inset -30px 0 30px var(--color-input-bg);
}

.board-shadow-left{
    box-shadow: inset 30px 0 30px var(--color-input-bg);
}

.board-container .checked{
    background-color: var(--game-cell-accent-bg);
    color: var(--game-cell-accent-text);
    transform: perspective(600px) rotateX(0deg) !important;
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
    color: var(--game-cell-accent-bg);
    background-color: var(--game-cell-accent-text);
    transform: translate(10%, -10%);
    box-shadow: 0 0 5px #000;
    opacity: 0;
    transition: 0.3s;
    overflow: hidden;
}

.board-container .checked::after{
    animation: check-marker 0.3s ease;
    top: 0;
    right: 0;
    height: 30px;
    width: 30px;
    opacity: 1;
}

@keyframes check-marker {
    75%{
        top: 0;
        right: 0;
        height: 40px;
        width: 40px;
        opacity: 1;
    }
}
/* === Party === */
.completed-actions {
    display: flex;
    justify-content: center;
}

.completed-actions .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
}

/* Bingo Toast */
/* === Bingo Modal === */
.bingo-modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(44, 42, 81, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 8000;
    backdrop-filter: blur(4px);
}

.bingo-modal {
    position: relative;
    background: linear-gradient(135deg, var(--game-cell-bg), var(--color-input-bg));
    border: 3px solid var(--color-subheading);
    border-radius: var(--radius-lg);
    padding: 32px 40px;
    text-align: center;
    box-shadow: 0 12px 48px rgba(44, 42, 81, 0.5);
    max-width: min(90vw, 380px);
}

.bingo-modal-cannon {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    pointer-events: none;
}

.bingo-modal-title {
    font-size: 3rem;
    font-weight: 900;
    letter-spacing: 0.06em;
    margin: 0 0 8px;
    background: linear-gradient(90deg, var(--color-heading), #C0185A, var(--color-subheading), #C0185A, var(--color-heading));
    background-size: 300% 100%;
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    animation: title-shine 4s linear infinite;
}

.bingo-modal-sub {
    color: var(--color-subheading);
    font-weight: 600;
    margin: 0 0 24px;
}

.bingo-modal-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
}

.bingo-particle {
    position: fixed;
    animation: confetti-drop linear forwards;
    pointer-events: none;
    z-index: 8001;
}

@keyframes confetti-drop {
    0%   { transform: translateY(0) rotate(0deg); opacity: 1; }
    85%  { opacity: 1; }
    100% { transform: translateY(105vh) rotate(680deg); opacity: 0; }
}

.bingo-modal-enter-active { animation: overlay-fade 0.25s ease-out; }
.bingo-modal-leave-active { animation: overlay-fade 0.2s ease-in reverse forwards; }
.bingo-modal-enter-active .bingo-modal { animation: banner-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1); }
.bingo-modal-leave-active .bingo-modal { animation: modal-out 0.2s ease-in forwards; }

@keyframes modal-out {
    from { opacity: 1; transform: scale(1); }
    to   { opacity: 0; transform: scale(0.9); }
}

.bingo-modal-enter-active .bingo-modal-cannon .cannon-piece,
.bingo-modal .cannon-piece {
    --duration: 0.9s;
}

.bingo-toast {
    position: fixed;
    bottom: 40px;
    left: 50%;
    translate: -50% 0;
    background: linear-gradient(135deg, var(--game-cell-accent-bg), var(--color-heading), #C0185A);
    background-size: 200% 100%;
    color: #fff;
    font-size: 2rem;
    font-weight: 900;
    letter-spacing: 0.1em;
    padding: 18px 48px;
    border-radius: var(--radius-lg);
    border: 2px solid rgba(255, 255, 255, 0.25);
    box-shadow:
        0 12px 48px rgba(64, 82, 182, 0.6),
        0 0 0 0 rgba(192, 24, 90, 0.4);
    z-index: 8888;
    pointer-events: none;
    white-space: nowrap;
    display: flex;
    align-items: center;
    gap: 12px;
    overflow: visible;
    animation: toast-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1), toast-pulse 0.6s ease-in-out 0.4s 3;
}

@keyframes toast-pulse {
    0%, 100% { box-shadow: 0 12px 48px rgba(64, 82, 182, 0.6), 0 0 0 0 rgba(192, 24, 90, 0.4); }
    50%       { box-shadow: 0 12px 48px rgba(64, 82, 182, 0.8), 0 0 0 16px rgba(192, 24, 90, 0); }
}

.toast-cannon {
    position: relative;
    display: inline-block;
    width: 0;
    height: 0;
}

.cannon-piece {
    position: absolute;
    bottom: 0;
    left: 0;
    width: 10px;
    height: 14px;
    border-radius: 2px;
    animation: cannon-shoot var(--duration, 0.9s) cubic-bezier(0.2, 0.8, 0.4, 1) var(--delay) both;
}

.bingo-toast-enter-active .cannon-piece {
    --duration: 0.9s;
}

@keyframes cannon-shoot {
    0% {
        transform: rotate(var(--angle)) translateX(0) scale(0.3);
        opacity: 1;
    }
    60% {
        opacity: 1;
    }
    100% {
        transform: rotate(var(--angle)) translateX(var(--dist)) scale(1);
        opacity: 0;
    }
}

.bingo-toast-enter-active { animation: toast-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1), toast-pulse 0.6s ease-in-out 0.4s 3; }
.bingo-toast-leave-active { animation: toast-out 0.25s ease-in forwards; }

@keyframes toast-in {
    from { opacity: 0; translate: -50% 60px; scale: 0.7; }
    to   { opacity: 1; translate: -50% 0;    scale: 1; }
}

@keyframes toast-out {
    from { opacity: 1; translate: -50% 0;     scale: 1; }
    to   { opacity: 0; translate: -50% -20px; scale: 0.9; }
}

.party-overlay {
    position: fixed;
    inset: 0;
    background: radial-gradient(circle at center, rgba(64, 82, 182, 0.92), rgba(44, 42, 81, 0.98));
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    overflow: hidden;
    cursor: pointer;
    animation: overlay-flash 0.6s ease-out, overlay-strobe 1.5s linear infinite 0.6s;
}

@keyframes overlay-flash {
    0%   { opacity: 0; }
    50%  { opacity: 1; background: radial-gradient(circle at center, rgba(255,255,255,0.95), rgba(255,255,255,0.7)); }
    100% { opacity: 1; background: radial-gradient(circle at center, rgba(64, 82, 182, 0.92), rgba(44, 42, 81, 0.98)); }
}

@keyframes overlay-strobe {
    0%, 100% { background: radial-gradient(circle at center, rgba(64, 82, 182, 0.92), rgba(44, 42, 81, 0.98)); }
    25%      { background: radial-gradient(circle at center, rgba(192, 24, 90, 0.92), rgba(44, 42, 81, 0.98)); }
    50%      { background: radial-gradient(circle at center, rgba(46, 125, 50, 0.92), rgba(44, 42, 81, 0.98)); }
    75%      { background: radial-gradient(circle at center, rgba(247, 159, 31, 0.92), rgba(44, 42, 81, 0.98)); }
}

.party-confetti {
    position: absolute;
    inset: 0;
    pointer-events: none;
}

/* Bingo-Bälle mit Nummer */
.bingo-ball {
    position: absolute;
    top: -60px;
    width: 42px;
    height: 42px;
    border-radius: 50%;
    color: var(--game-cell-accent-bg);
    font-weight: 800;
    font-size: 0.95rem;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow:
        inset -4px -6px 8px rgba(0, 0, 0, 0.25),
        inset 4px 4px 6px rgba(255, 255, 255, 0.5),
        0 4px 14px rgba(0, 0, 0, 0.4);
    animation-name: ball-fall;
    animation-timing-function: cubic-bezier(0.45, 0.05, 0.55, 0.95);
    animation-iteration-count: infinite;
}

.bingo-ball::before {
    content: '';
    position: absolute;
    inset: 5px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.9);
    z-index: -1;
}

.bingo-ball::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: inherit;
    z-index: -2;
}

@keyframes ball-fall {
    0%   { transform: translateY(-60px) rotate(0deg) scale(1);    opacity: 1; }
    50%  { transform: translateY(50vh) rotate(360deg) scale(1.1); }
    85%  { opacity: 1; }
    100% { transform: translateY(110vh) rotate(720deg) scale(0.8); opacity: 0; }
}

/* Spielkarten-Konfetti */
.party-card {
    position: absolute;
    top: -40px;
    width: 26px;
    height: 38px;
    background: var(--game-cell-bg);
    border: 2px solid var(--color-subheading);
    border-radius: 3px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
    animation-name: card-fall;
    animation-timing-function: linear;
    animation-iteration-count: infinite;
}

.party-card::before {
    content: '';
    position: absolute;
    inset: 3px;
    background: repeating-linear-gradient(
        45deg,
        var(--color-heading),
        var(--color-heading) 3px,
        var(--color-subheading) 3px,
        var(--color-subheading) 6px
    );
    border-radius: 1px;
}

@keyframes card-fall {
    0%   { transform: translateY(-40px) rotate(0deg);     opacity: 1; }
    85%  { opacity: 1; }
    100% { transform: translateY(110vh) rotate(1080deg);  opacity: 0; }
}

/* Würfel-Regen */
.party-falling-dice {
    position: absolute;
    top: -50px;
    width: 32px;
    height: 32px;
    background: var(--game-cell-bg);
    border: 3px solid var(--color-heading);
    border-radius: 6px;
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.4);
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    grid-template-rows: 1fr 1fr 1fr;
    padding: 4px;
    gap: 2px;
    animation-name: dice-fall;
    animation-timing-function: cubic-bezier(0.5, 0, 0.5, 1);
    animation-iteration-count: infinite;
}

.party-falling-dice .dot {
    background: var(--color-heading);
    border-radius: 50%;
}

@keyframes dice-fall {
    0%   { transform: translateY(-50px) rotate(0deg);   opacity: 1; }
    85%  { opacity: 1; }
    100% { transform: translateY(110vh) rotate(1440deg); opacity: 0; }
}

/* Sterne */
.party-star {
    position: absolute;
    top: -30px;
    color: #F79F1F;
    filter: drop-shadow(0 0 8px rgba(247, 159, 31, 0.8));
    animation-name: star-fall;
    animation-timing-function: linear;
    animation-iteration-count: infinite;
}

@keyframes star-fall {
    0%   { transform: translateY(-30px) rotate(0deg) scale(0.5);   opacity: 0; }
    10%  { opacity: 1; transform: translateY(10vh) rotate(180deg) scale(1.2); }
    85%  { opacity: 1; }
    100% { transform: translateY(110vh) rotate(720deg) scale(0.6); opacity: 0; }
}

/* Emoji */
.party-emoji {
    position: absolute;
    top: -40px;
    font-size: 2.5rem;
    animation-name: emoji-fall;
    animation-timing-function: ease-in;
    animation-iteration-count: infinite;
    filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
}

@keyframes emoji-fall {
    0%   { transform: translateY(-40px) rotate(-30deg) scale(1);  opacity: 1; }
    50%  { transform: translateY(50vh) rotate(180deg) scale(1.4); }
    100% { transform: translateY(110vh) rotate(360deg) scale(1);  opacity: 0; }
}

/* Beats - pulsierende Kreise vom Banner aus */
.party-beat {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 100px;
    height: 100px;
    border-radius: 50%;
    border: 4px solid;
    transform: translate(-50%, -50%);
    pointer-events: none;
    animation: beat-expand 1.2s ease-out infinite;
}

.party-beat-1 { border-color: #C0185A; animation-delay: 0s; }
.party-beat-2 { border-color: var(--color-heading); animation-delay: 0.4s; }
.party-beat-3 { border-color: #F79F1F; animation-delay: 0.8s; }

@keyframes beat-expand {
    0%   { width: 100px; height: 100px; opacity: 1;   border-width: 6px; }
    100% { width: 200vmax; height: 200vmax; opacity: 0; border-width: 1px; }
}

.party-banner {
    position: relative;
    background: linear-gradient(135deg, var(--game-cell-bg), var(--color-input-bg));
    border: 4px solid var(--color-subheading);
    border-radius: var(--radius-lg);
    padding: 24px clamp(20px, 5vw, 56px);
    text-align: center;
    box-shadow: 0 20px 60px rgba(44, 42, 81, 0.6);
    animation: banner-pop 0.6s cubic-bezier(0.34, 1.56, 0.64, 1), banner-shake 0.4s ease-in-out 0.6s infinite;
    z-index: 2;
    max-width: min(90vw, 500px);
    box-sizing: border-box;
}

/* Konfetti-Burst seitlich aus dem Banner */
.banner-burst {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 0;
    height: 0;
    pointer-events: none;
    z-index: 3;
}

.burst-piece {
    position: absolute;
    top: 0;
    left: 0;
    width: 10px;
    height: 16px;
    border-radius: 2px;
    transform-origin: center;
    animation: burst-fly var(--duration) cubic-bezier(0.2, 0.7, 0.4, 1) var(--delay) infinite;
    box-shadow: 0 0 6px rgba(0, 0, 0, 0.3);
}

@keyframes burst-fly {
    0% {
        transform: rotate(var(--angle)) translateX(0) rotate(calc(-1 * var(--angle))) scale(0.5);
        opacity: 1;
    }
    60% {
        opacity: 1;
    }
    100% {
        transform: rotate(var(--angle)) translateX(var(--distance)) rotate(calc(-1 * var(--angle) + 720deg)) scale(1);
        opacity: 0;
    }
}

@keyframes banner-pop {
    0%   { transform: scale(0) rotate(-15deg); opacity: 0; }
    60%  { transform: scale(1.15) rotate(3deg); opacity: 1; }
    100% { transform: scale(1) rotate(0deg);   opacity: 1; }
}

@keyframes banner-shake {
    0%, 100% { transform: translate(0, 0) rotate(0deg); }
    25%      { transform: translate(-3px, 2px) rotate(-1deg); }
    50%      { transform: translate(3px, -2px) rotate(1deg); }
    75%      { transform: translate(-2px, 3px) rotate(-0.5deg); }
}

.party-dice-row {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 12px;
}

.party-dice {
    color: var(--color-heading);
    filter: drop-shadow(0 4px 8px rgba(64, 82, 182, 0.4));
}

.party-dice-1 { animation: dice-roll-l 0.7s ease-in-out infinite; }
.party-dice-2 { animation: dice-roll-r 0.7s ease-in-out infinite; }

@keyframes dice-roll-l {
    0%, 100% { transform: rotate(-20deg) translateY(0) scale(1); }
    50%      { transform: rotate(20deg) translateY(-14px) scale(1.15); }
}

@keyframes dice-roll-r {
    0%, 100% { transform: rotate(20deg) translateY(0) scale(1); }
    50%      { transform: rotate(-20deg) translateY(-14px) scale(1.15); }
}

.party-sparkle {
    color: #C0185A;
    animation: sparkle-pulse 0.8s ease-in-out infinite;
    filter: drop-shadow(0 0 12px rgba(192, 24, 90, 0.8));
}

@keyframes sparkle-pulse {
    0%, 100% { transform: scale(1) rotate(0deg);     opacity: 1; }
    50%      { transform: scale(1.6) rotate(180deg); opacity: 0.7; }
}

.party-title {
    margin: 16px 0 4px;
    font-size: clamp(2.5rem, 12vw, 5rem);
    font-weight: 900;
    letter-spacing: 0.05em;
    line-height: 1;
    word-break: break-word;
    background: linear-gradient(90deg, var(--color-heading), #C0185A, #F79F1F, #2E7D32, var(--color-heading));
    background-size: 300% 100%;
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    color: transparent;
    animation: title-shine 1.2s linear infinite, title-bounce 0.5s ease-in-out infinite;
}

@keyframes title-shine {
    from { background-position: 0% 0; }
    to   { background-position: 300% 0; }
}

@keyframes title-bounce {
    0%, 100% { transform: scale(1) skewX(0deg); }
    50%      { transform: scale(1.08) skewX(-3deg); }
}

.party-subtitle {
    margin: 0;
    color: var(--color-subheading);
    font-weight: 600;
    animation: subtitle-pulse 0.8s ease-in-out infinite;
}

@keyframes subtitle-pulse {
    0%, 100% { opacity: 1; }
    50%      { opacity: 0.6; }
}

/* === Bingo sweep === */
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
