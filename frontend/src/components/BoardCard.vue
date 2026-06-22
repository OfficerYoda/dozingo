<template>
    <button class="card card-border-blue board-card" @click="$emit('click')">
        <div class="card-body">
            <h3>{{ board.title }}</h3>
            <small>{{ board.description ?? '—' }}</small>
        </div>
        <hr class="mb-2">
        <div class="card-footer">
            <div class="play-group">
                <Play :size="20" />
                <span class="card-meta-text">{{ formattedPlayCount }}</span>
            </div>
            <button
                class="like-group"
                :class="{ liked: userVote === 1, interactive: !!user }"
                type="button"
                :aria-label="t('boards.likeBoard')"
                @click.stop="handleLike"
            >
                <Heart :size="20" :fill="userVote === 1 ? 'currentColor' : 'none'" />
                <span class="card-meta-text">{{ formattedScore }}</span>
            </button>
        </div>
    </button>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Heart, Play } from 'lucide-vue-next'
import { useAuth } from '@/composables/useAuth'
import * as voteService from '@/services/vote.service'
import type { Board } from '@/services/api.type'

const props = defineProps<{ board: Board }>()
const emit = defineEmits<{ click: []; 'vote-changed': [vote: number | null] }>()

const { t } = useI18n()
const { state } = useAuth()
const user = computed(() => state.user)

const score = ref(props.board.score)
const userVote = ref<number | null>(null)

const formattedScore = computed(() => {
    if (score.value >= 1000) return `${(score.value / 1000).toFixed(1)}k`
    return String(score.value)
})

const formattedPlayCount = computed(() => {
    const n = props.board.play_count
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
    return String(n)
})

// load initial vote state
voteService.getBoardVote(props.board.board_id)
    .then(data => {
        score.value = data.score
        userVote.value = data.user_vote
    })
    .catch(() => { /* not logged in */ })

async function handleLike() {
    if (!user.value) return
    const wasLiked = userVote.value === 1
    try {
        if (wasLiked) {
            await voteService.deleteVote(props.board.board_id)
            score.value--
            userVote.value = null
        } else {
            await voteService.voteBoard(props.board.board_id, 1)
            score.value += userVote.value === -1 ? 2 : 1
            userVote.value = 1
        }
        emit('vote-changed', userVote.value)
    } catch { /* state bleibt unverändert */ }
}
</script>

<style scoped>
.board-card {
    width: 100%;
    text-align: left;
}

.card-body {
    text-align: left;
}

.card-footer {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
}

.card-meta-text {
    color: var(--color-subheading);
    font-weight: 600;
    font-size: 0.8125rem;
}

.play-group {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 4px;
    padding: 2px 6px;
}

.like-group {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 4px;
    background: none;
    border: none;
    cursor: default;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    font: inherit;
    transition: background-color 0.15s;
}

.like-group.interactive {
    cursor: pointer;
}

.like-group.interactive:hover {
    background-color: var(--color-bg-hover);
}

.like-group svg {
    color: var(--color-subheading);
    transition: color 0.15s;
}

.like-group.liked svg {
    color: #C0185A;
}

.like-group.liked .card-meta-text {
    color: #C0185A;
}
</style>
