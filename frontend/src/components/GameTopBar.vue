<template>
    <div class="top-item-bar">
        <div class="top-board-info">
            <div class="stats">
                <a href="/" class="stat-item back"><ArrowLeft :size="13"/> Back</a>
                <span class="stat-item stat-plays">
                    <Play :size="13"/> {{ board?.play_count ?? '—' }}
                </span>
                <button class="stat-item stat-likes" :class="{ liked: userVote === 1 }" @click="$emit('likeClick')" type="button" aria-label="Board liken">
                    <Heart :size="13" :fill="userVote === 1 ? 'currentColor' : 'none'"/> {{ board?.score ?? '—' }}
                </button>
            </div>
        </div>
        <div class="top-stat-pill">
            <Timer :size="15" class="top-stat-icon"/>
            <span class="top-stat-value">{{ formattedTime }}</span>
        </div>
        <button class="top-stat-pill top-action-btn" :class="{ 'share-copied': copyToast }" type="button" @click="shareGame" :aria-label="copyToast ? 'Link kopiert!' : 'Teilen'">
            <Check v-if="copyToast" :size="15" class="share-check"/>
            <Share2 v-else :size="15"/>
        </button>
        <button class="top-stat-pill top-action-btn top-fullscreen-btn" type="button" @click="toggleFullscreen" :aria-label="isFullscreen ? 'Fullscreen beenden' : 'Fullscreen'">
            <Minimize2 v-if="isFullscreen" :size="15"/>
            <Maximize2 v-else :size="15"/>
        </button>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Heart, Play, Timer, ArrowLeft, Maximize2, Minimize2, Share2, Check } from 'lucide-vue-next'
import type { Board } from '@/services/api.type'

const props = defineProps<{
    board: Board | null
    userVote: number | null
    formattedTime: string
}>()

defineEmits<{ likeClick: [] }>()

const isFullscreen = ref(!!document.fullscreenElement)
const copyToast = ref(false)

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

function shareGame() {
    const url = window.location.href
    const title = props.board?.title ?? 'Dozingo'
    if (navigator.share) {
        navigator.share({ title, url }).catch(() => {})
    } else {
        navigator.clipboard.writeText(url).then(() => {
            copyToast.value = true
            setTimeout(() => { copyToast.value = false }, 2000)
        }).catch(() => {})
    }
}

onMounted(() => {
    document.addEventListener('fullscreenchange', onFullscreenChange)
})

onUnmounted(() => {
    document.removeEventListener('fullscreenchange', onFullscreenChange)
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

.share-copied {
    background-color: color-mix(in srgb, #2E7D32 15%, var(--color-input-bg));
    color: #2E7D32;
}

.share-copied:hover {
    background-color: color-mix(in srgb, #2E7D32 25%, var(--color-input-bg));
}

.share-check {
    color: #2E7D32;
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
</style>
