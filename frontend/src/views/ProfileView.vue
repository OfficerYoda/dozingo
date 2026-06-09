<template>
  <div className="container">
    <div v-if="!auth.state.ready">Loading...</div>
    <div v-else-if="!auth.state.user">Not logged in.</div>
    <div v-else class="profile-header">
      <img v-if="auth.state.user.avatar_url" :src="auth.state.user.avatar_url" alt="Profile picture" class="profile-avatar" />
      <img v-else src="/user.png" alt="Profile picture" class="profile-avatar" />
      <h1>Welcome, {{ auth.state.user.username }}!</h1>
    </div>
    <div class="container" style="padding-right: 0%; padding-left: 0%;">
      <h2 class="mb-0">Explore your {{boards.length}} boards</h2>

      <p v-if="error" class="error-text">{{ error }}</p>

      <SliderSection :items="boards" :per-page="3" :per-page-md="2" :per-page-sm="1">
        <template #slide="{ item: board }">
          <button class="card card-border-blue profile-slider-card" @click="clickBoard(board.board_id)">
            <div class="card-body">
              <h3>{{ board.title }}</h3>
              <small>{{ board.description }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">Played {{ board.play_count }} times</span>
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ board.score }}</span>
              </div>
            </div>
          </button>
        </template>
      </SliderSection>
    </div>
    <div class="container" style="padding-right: 0%; padding-left: 0%; padding-top: 2%;">
      <div class="list-header mb-4">
          <h2 class="mb-0">Explore your {{ likedBoards.length }} liked boards</h2>
      </div>

      <SliderSection :items="likedBoards" :per-page="4" :per-page-md="2" :per-page-sm="1">
        <template #slide="{ item: vote }">
          <button class="card card-border-blue profile-slider-card" @click="clickBoard(vote.board_id)">
            <div class="card-body">
              <h3>{{ vote.board_title }}</h3>
              <small>{{ vote.board_description }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ vote.vote_score }}</span>
              </div>
            </div>
          </button>
        </template>
      </SliderSection>
    </div>

  <ModalStartGame
    v-if="selecetedBoard"
    v-model="showModal"
    :board="selecetedBoard"
    :cells="selectedCells ?? []"
    :author-name="authorName"
  />

  </div>
</template>

<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Heart } from 'lucide-vue-next'
import SliderSection from '@/components/SliderSection.vue'
import ModalStartGame from '@/components/ModalStartGame.vue'

const auth = useAuth()
if (!auth.state.ready) {
  auth.fetchUser()
}

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
    cell_id: string,
    content: string,
    value: number,
}

interface Vote {
    vote_id: string
    board_id: string
    vote_value: number
    board_title: string
    board_description: string
    vote_score: number
    vote_count: number
}

const route = useRoute()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const likedBoards = ref<Vote[]>([])
const cells = ref<Cell[]>([])
const selecetedBoard = ref<Board>()
const selectedCells = ref<Cell[]>()
const authorName = ref<string | null>(null)

async function fetchAllUserBoards() {
    const params = new URLSearchParams()
    if(auth.state.user) params.set('author_id', auth.state.user.user_id)
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

async function fetchLikedBoards() {
    const res = await fetch('/api/users/me/votes', { credentials: 'include' })
    if (!res.ok) return
    const votes: Vote[] = await res.json()
    likedBoards.value = votes.filter(v => v.vote_value === 1)
}

async function fetchAllCellsForBoard(boardID: string) {
    const cellsRes = await fetch('/api/boards/' + boardID + '/cells')
    if (!cellsRes.ok) {
        error.value = 'Failed to load cells for board ' + boardID
        return
    }

    cells.value = await cellsRes.json()
    selecetedBoard.value = boards.value.find(b => b.board_id === boardID)
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)

    authorName.value = null
    if (selecetedBoard.value?.author_id) {
        const userRes = await fetch('/api/users/' + selecetedBoard.value.author_id)
        if (userRes.ok) {
            const user = await userRes.json()
            authorName.value = user.username
        }
    }

    showModal.value = true
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllUserBoards()
}, { immediate: true })

fetchLikedBoards()

const showModal = ref(false)


function clickBoard(boardID: string) {
    fetchAllCellsForBoard(boardID)
}
</script>

<style scoped>
.profile-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  margin-top: 1rem;
}

.profile-header h1 {
  margin: 0;
}

.profile-avatar {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.profile-slider-card {
  width: 100%;
  height: 100%;
  text-align: left;
}
</style>
