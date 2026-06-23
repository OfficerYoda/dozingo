<template>
    <Transition name="bingo-toast">
        <div v-if="show && dismissed" class="bingo-toast">
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
</template>

<script setup lang="ts">
defineProps<{
    show: boolean
    dismissed: boolean
}>()

const confettiColors = ['var(--color-heading)', '#C0185A', '#2E7D32', '#F79F1F', 'var(--color-subheading)', 'var(--color-input-bg)', '#EA2027']
</script>

<style scoped>
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
</style>
