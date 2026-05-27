<template>
    <section>
        <div class="container">
            <div class="list-header mb-4">
                <h2 class="mb-0">Explore all Cards</h2>
                <div class="header-actions">
                    <input class="btn btn-secondary" type="search" placeholder="Search.." v-model="search">
                    <select class="btn btn-secondary" v-model="appliedFiler">
                        <option value="">No Filter</option>
                        <option value="newest">Newest</option>
                        <option value="most-liked">Most liked</option>
                        <option value="most-played">Most played</option>
                        <option value="oldest">Oldest</option>
                        <option value="least-liked">Least liked</option>
                        <option value="least-played">Least played</option>
                    </select>
                </div>
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="grid">
                <button v-for="board in boards" :key="board.board_id" class="card card-border-blue col-4 md-6 sm-12">
                    <div class="card-body">
                        <h3>{{ board.title }}</h3>
                        <small>{{ board.description }}</small>
                    </div>
                    <hr class="mb-2">
                    <div class="card-footer">
                        <span class="card-meta-text">Played {{ board.vote_count}} times</span>
                        <div class="like-group">
                            <Heart :size="20" />
                            <span class="card-meta-text">{{ board.vote_count}}</span>
                        </div>
                    </div>
                </button>
            </div>
        </div>
    </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Heart, Variable } from 'lucide-vue-next'

interface Board {
    board_id: string
    title: string
    description: string
    play_count: number
    vote_count: number
}

useI18n()
const route = useRoute()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])

async function fetchAllCells() {
    const params = new URLSearchParams()
    if (appliedFiler.value) params.set('sort', appliedFiler.value)
    if (search.value) params.set('search', search.value)

    const query = params.toString() ? '?' + params.toString() : ''
    const boardsRes = await fetch('/api/boards' + query, { credentials: 'include' })
    if (!boardsRes.ok) {
        error.value = 'Failed to load boards'
        return
    }

    boards.value = await boardsRes.json()
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllCells()
}, { immediate: true })

</script>

<style scoped>
/* Header */
.list-header {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.header-actions {
    display: flex;
    flex-direction: row;
    gap: 8px;
}

/* Card body */
.card-body {
    text-align: left;
}

/* Card footer */
.card-footer {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.like-group {
    display: flex;
    flex-direction: row;
    gap: 4px;
}

.like-group svg {
    color: #5A5781;
}

.card-meta-text {
    color: #5A5781;
    font-weight: 600;
    font-size: 13px;
}
</style>
