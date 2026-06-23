<template>
    <Teleport to="body">
        <div v-if="show" class="party-overlay" @click="router.push('/')">
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
                <h1 class="party-title">DOZINGO</h1>
                <p class="party-subtitle">{{ formattedTime }} · alle {{ cellCount }} Felder geschafft</p>
                <RouterLink to="/" class="btn btn-primary mt-3">Zur Startseite</RouterLink>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Sparkles, Dices, Star } from 'lucide-vue-next'

defineProps<{
    show: boolean
    formattedTime: string
    cellCount: number
}>()

const router = useRouter()

const confettiColors = ['var(--color-heading)', '#C0185A', '#2E7D32', '#F79F1F', 'var(--color-subheading)', 'var(--color-input-bg)', '#EA2027']
const partyEmojis = ['🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯', '🍾', '🎈', '💫', '🔥', '🎉', '🎊', '🥳', '🎲', '🏆', '⭐', '✨', '🎯']
</script>

<style scoped>
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
</style>
