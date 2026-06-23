<template>
    <Teleport to="body">
        <Transition name="bingo-modal">
            <div v-if="modelValue" class="bingo-modal-overlay">
                <div
                    v-for="p in bingoParticles" :key="p.id"
                    class="bingo-particle"
                    :style="p.style"
                />
                <div class="bingo-modal">
                    <p class="bingo-modal-title">BINGO!</p>
                    <p class="bingo-modal-sub">Du hast eine Reihe vervollständigt!</p>
                    <div class="bingo-modal-actions">
                        <button class="btn btn-secondary" @click="$emit('continue')">Weiterspielen</button>
                        <button class="btn btn-primary" @click="$emit('finish')">Fertig!</button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ modelValue: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean]; continue: []; finish: [] }>()

const confettiColors = ['var(--color-heading)', '#C0185A', '#2E7D32', '#F79F1F', 'var(--color-subheading)', 'var(--color-input-bg)', '#EA2027']
const bingoParticles = ref<Array<{ id: number; style: Record<string, string> }>>([])

watch(() => props.modelValue, (val) => {
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
</script>

<style scoped>
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

@keyframes overlay-fade {
    from { opacity: 0; }
    to   { opacity: 1; }
}

@keyframes modal-out {
    from { opacity: 1; transform: scale(1); }
    to   { opacity: 0; transform: scale(0.9); }
}

@keyframes banner-pop {
    0%   { transform: scale(0) rotate(-15deg); opacity: 0; }
    60%  { transform: scale(1.15) rotate(3deg); opacity: 1; }
    100% { transform: scale(1) rotate(0deg);   opacity: 1; }
}

@keyframes title-shine {
    from { background-position: 0% 0; }
    to   { background-position: 300% 0; }
}
</style>
