import { ref, computed } from 'vue'

const elapsedSeconds = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null
let heartbeatInterval: ReturnType<typeof setInterval> | null = null

const formattedTime = computed(() => {
    const m = Math.floor(elapsedSeconds.value / 60).toString().padStart(2, '0')
    const s = (elapsedSeconds.value % 60).toString().padStart(2, '0')
    return `${m}:${s}`
})

export function useGameTimer() {
    function startTimer(gameId: string) {
        if (timerInterval) return
        timerInterval = setInterval(() => { elapsedSeconds.value++ }, 1000)
        heartbeatInterval = setInterval(() => {
            fetch(`/api/games/${gameId}/heartbeat`, { method: 'POST', credentials: 'include' }).catch(() => {})
        }, 30000)
    }

    function stopTimer() {
        if (timerInterval) { clearInterval(timerInterval); timerInterval = null }
        if (heartbeatInterval) { clearInterval(heartbeatInterval); heartbeatInterval = null }
    }

    return { elapsedSeconds, formattedTime, startTimer, stopTimer }
}
