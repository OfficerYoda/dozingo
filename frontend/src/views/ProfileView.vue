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
      <h2 class="mb-0">Continue your unfinished Boards:</h2>

      <p v-if="error" class="error-text">{{ error }}</p>

      <SliderSection :items="activeGames" :per-page="3" :per-page-md="2" :per-page-sm="1">
        <template #slide="{ item: game }">
          <button class="card card-border-blue profile-slider-card" @click="router.push('/game/' + game.game_id)">
            <div class="card-body">
              <h3>{{ game.board_title }}</h3>
              <small>{{ game.marked_count }} / {{ game.total_count }} cells marked</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">Continue playing</span>
            </div>
          </button>
        </template>
      </SliderSection>

      <p v-if="error" class="error-text">{{ error }}</p>

      <h2 class="mb-0">Explore your {{ boards.length }} boards</h2>

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
    
        <h2 class="mb-0">Explore your {{ likedBoards.length }} liked boards</h2>

    <SliderSection :items="likedBoards" :per-page="3" :per-page-md="2" :per-page-sm="1">
      <template #slide="{ item: vote }">
        <button class="card card-border-blue profile-slider-card" @click="clickBoard(vote.board_id)">
          <div class="card-body">
            <h3>{{ vote.title }}</h3>
            <small>{{ vote.description }}</small>
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
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router'
import { Heart } from 'lucide-vue-next'
import SliderSection from '@/components/SliderSection.vue'
import ModalStartGame from '@/components/ModalStartGame.vue'
import { usePageTitle } from '@/composables/usePageTitle'
import * as boardService from '@/services/board.service'
import * as userService from '@/services/user.service'
import * as gameService from '@/services/game.service'
import * as voteService from '@/services/vote.service'
import type { Board, Cell, Vote, Game } from '@/services/api.type'

const auth = useAuth()
if (!auth.state.ready) {
  auth.fetchUser()
}

interface ActiveGame {
    game_id: string
    board_id: string
    board_title: string
    marked_count: number
    total_count: number
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const likedBoards = ref<Vote[]>([])
const activeGames = ref<ActiveGame[]>([])
const cells = ref<Cell[]>([])
const selecetedBoard = ref<Board>()
const selectedCells = ref<Cell[]>()
const authorName = ref<string | null>(null)

const { pageTitle } = usePageTitle(t('header.profile'))

async function fetchAllUserBoards() {
    try {
        boards.value = await boardService.getBoards({
            author_id: auth.state.user?.user_id,
            sort: appliedFiler.value || undefined,
            search: search.value || undefined,
        })
    } catch {
        error.value = 'Failed to load boards'
    }
}

async function fetchActiveGames() {
    if (!auth.state.user) return
    try {
        const games: Game[] = await gameService.getGamesByUser(auth.state.user.user_id)
        const active = games.filter(g => g.status === 'active')

        activeGames.value = await Promise.all(active.map(async (g) => {
            let board_title = g.board_id ?? 'Unknown board'
            try {
                const board = await boardService.getBoardById(g.board_id)
                board_title = board.title
            } catch { /* board title stays as fallback */ }

            let marked_count = 0
            let total_count = 0
            try {
                const gameCells = await gameService.getGameCells(g.game_id)
                total_count = gameCells.length
                marked_count = gameCells.filter(c => c.is_marked).length
            } catch { /* counts stay 0 */ }

            return { game_id: g.game_id, board_id: g.board_id, board_title, marked_count, total_count }
        }))
    } catch { /* ignore */ }
}

async function fetchLikedBoards() {
    try {
        const votes = await voteService.getMyVotes()
        likedBoards.value = votes.filter(v => v.vote_value === 1)
    } catch { /* ignore */ }
}

async function fetchAllCellsForBoard(boardID: string) {
    try {
        cells.value = await boardService.getCellsForBoard(boardID)
    } catch {
        error.value = 'Failed to load cells for board ' + boardID
        return
    }

    selecetedBoard.value = boards.value.find(b => b.board_id === boardID)
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)

    authorName.value = null
    if (selecetedBoard.value?.author_id) {
        try {
            const user = await userService.getUserById(selecetedBoard.value.author_id)
            authorName.value = user.username
        } catch { /* author is optional */ }
    }

    showModal.value = true
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllUserBoards()
}, { immediate: true })

fetchLikedBoards()
fetchActiveGames()

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
  min-height: 8rem;
  text-align: left;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
</style>
