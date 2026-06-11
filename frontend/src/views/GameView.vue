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
            
            <div :class="['board', 'mt-3', { stopped: gameState === 'stopped', completed: gameState === 'completed' }]">
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

            <div v-if="gameState === 'completed'" class="completed-actions mt-3">
                <RouterLink to="/" class="btn btn-primary">
                    <Home :size="18"/>
                    <span>Zur Startseite</span>
                </RouterLink>
            </div>
        </div>

        <Teleport to="body">
            <div v-if="showParty" class="party-overlay" @click="dismissParty">
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
                    <button class="btn btn-primary mt-3" @click.stop="dismissParty">Weiter</button>
                </div>
            </div>
        </Teleport>
    </section>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, useTemplateRef, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Heart, LayoutGrid, Play, Timer, CheckSquare, Sparkles, Dices, Star, Home } from 'lucide-vue-next'
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
const gameId = ref<string>('')

// 'stopped' | 'running' | 'completed'
const gameState = ref<'stopped' | 'running' | 'completed'>('stopped')
const revealedCells = ref<Set<number>>(new Set())
const checkedCells = ref<Set<string>>(new Set())
const isRevealing = ref(false)
const showParty = ref(false)
const confettiColors = ['#4052B6', '#C0185A', '#2E7D32', '#F79F1F', '#5A5781', '#E3DFFF', '#EA2027']
const partyEmojis = ['🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯', '🍾', '🎈', '💫', '🔥', '🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯']

function dismissParty() {
    showParty.value = false
    stopTechno()
}

// --- Techno-Beat (Web Audio) ---
let audioCtx: AudioContext | null = null
let beatScheduler: ReturnType<typeof setInterval> | null = null
let beatStep = 0

function startTechno() {
    if (audioCtx) return
    const Ctx = (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)
    audioCtx = new Ctx()

    // 130 BPM → 16th note = 60/130/4 ≈ 0.115s
    const stepMs = (60 / 130) * 1000 / 4
    beatStep = 0
    beatScheduler = setInterval(() => playStep(beatStep++), stepMs)
}

function stopTechno() {
    if (beatScheduler) { clearInterval(beatScheduler); beatScheduler = null }
    if (audioCtx) { audioCtx.close(); audioCtx = null }
}

function playStep(step: number) {
    if (!audioCtx) return
    const t = audioCtx.currentTime
    const beat = step % 16

    // Kick auf jedem Viertel (0, 4, 8, 12)
    if (beat % 4 === 0) playKick(t)
    // Snare/Clap auf 4 und 12
    if (beat === 4 || beat === 12) playClap(t)
    // Hi-Hat auf jedem 8tel (0, 2, 4, ...)
    if (beat % 2 === 0) playHat(t, beat % 4 === 2)
    // Synth-Stab Riff (Acid)
    const stabPattern: Record<number, number> = { 0: 55, 3: 62, 6: 69, 7: 55, 10: 62, 14: 73 }
    const stab = stabPattern[beat]
    if (stab !== undefined) {
        playStab(t, stab)
    }
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
    gameId.value = route.params.game_id as string
    error.value = null

    const gameRes = await fetch(`/api/games/${gameId.value}`, { credentials: 'include' })
    if (!gameRes.ok) { error.value = 'Spiel nicht gefunden'; return }
    const game = await gameRes.json()

    const [boardRes, cellsRes] = await Promise.all([
        fetch(`/api/boards/${game.board_id}`, { credentials: 'include' }),
        fetch(`/api/games/${gameId.value}/cells`, { credentials: 'include' }),
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

    // Falls Spiel server-seitig bereits abgeschlossen → direkt locken
    if (game.status === 'completed') {
        await startGame()
        gameState.value = 'completed'
        stopTimer()
        return
    }

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

async function completeGame() {
    if (gameState.value === 'completed') return
    gameState.value = 'completed'
    stopTimer()
    showParty.value = true
    startTechno()
    try {
        await fetch(`/api/games/${gameId.value}/status`, {
            method: 'PUT',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: 'completed' }),
        })
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
        const res = await fetch(`/api/games/${gameId.value}/cells/${cellId}`, {
            method: 'PUT',
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ is_marked: nextChecked }),
        })
        if (!res.ok) throw new Error('mark failed')
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
})

onUnmounted(() => {
    resizeObserver?.disconnect()
    stopTimer()
    stopTechno()
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
    transition: transform 0.6s ease, border-width 0.3s, padding 0.3s, border-color 0.3s, background-color 0.3s, color 0.3s;
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
    color: #5A5781;
    background-color: #E3DFFF;
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
    color: #2C2A51;
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
    background: #fff;
    border: 2px solid #5A5781;
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
        #4052B6,
        #4052B6 3px,
        #5A5781 3px,
        #5A5781 6px
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
    background: #fff;
    border: 3px solid #2C2A51;
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
    background: #2C2A51;
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
.party-beat-2 { border-color: #4052B6; animation-delay: 0.4s; }
.party-beat-3 { border-color: #F79F1F; animation-delay: 0.8s; }

@keyframes beat-expand {
    0%   { width: 100px; height: 100px; opacity: 1;   border-width: 6px; }
    100% { width: 200vmax; height: 200vmax; opacity: 0; border-width: 1px; }
}

.party-banner {
    position: relative;
    background: linear-gradient(135deg, #fff, #E3DFFF);
    border: 4px solid #5A5781;
    border-radius: var(--radius-lg);
    padding: 32px 56px;
    text-align: center;
    box-shadow: 0 20px 60px rgba(44, 42, 81, 0.6);
    animation: banner-pop 0.6s cubic-bezier(0.34, 1.56, 0.64, 1), banner-shake 0.4s ease-in-out 0.6s infinite;
    z-index: 2;
    max-width: 90vw;
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
    color: #4052B6;
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
    font-size: 5rem;
    font-weight: 900;
    letter-spacing: 0.08em;
    background: linear-gradient(90deg, #4052B6, #C0185A, #F79F1F, #2E7D32, #4052B6);
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
    color: #5A5781;
    font-weight: 600;
    animation: subtitle-pulse 0.8s ease-in-out infinite;
}

@keyframes subtitle-pulse {
    0%, 100% { opacity: 1; }
    50%      { opacity: 0.6; }
}

</style>
