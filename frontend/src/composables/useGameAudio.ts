let audioCtx: AudioContext | null = null
let beatScheduler: ReturnType<typeof setTimeout> | null = null

const BPM = 130
const STEP = 60 / BPM / 4  // 16th note ≈ 0.115s
const BAR  = STEP * 16      // one bar  ≈ 1.846s

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

export function useGameAudio() {
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

    return { startTechno, stopTechno }
}
