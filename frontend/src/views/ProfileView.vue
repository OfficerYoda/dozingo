<template>
  <div className="container">
    <div v-if="!auth.state.ready">{{ t('profile.loading') }}</div>
    <div v-else-if="!auth.state.user">{{ t('profile.notLoggedIn') }}</div>
    <div v-else class="profile-header">
      <img v-if="auth.state.user.avatar_url" :src="auth.state.user.avatar_url" alt="Profile picture" class="profile-avatar" />
      <img v-else src="/user.png" alt="Profile picture" class="profile-avatar" />
      <h1>{{ t('profile.welcome', { username: auth.state.user.username }) }}</h1>
    </div>

    <div class="container" style="padding-right: 0%; padding-left: 0%;">
      <h2 class="mb-0">{{ t('profile.continueBoards') }}</h2>

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
              <span class="card-meta-text">{{ t('profile.continuePlaying') }}</span>
            </div>
          </button>
        </template>
      </SliderSection>

      <p v-if="error" class="error-text">{{ error }}</p>

      <h2 class="mb-0">{{ t('profile.exploreBoards', { count: boards.length }) }}</h2>

      <SliderSection :items="boards" :per-page="3" :per-page-md="2" :per-page-sm="1">
        <template #slide="{ item: board }">
          <div class="board-card-wrapper">
            <BoardCard :key="`${board.board_id}-${boardsVersion}`" :board="board" :played-label="`Played ${board.play_count} times`" @click="clickBoard(board.board_id)" @vote-changed="(v: number | null) => onBoardVoteChange(v, board)" />
            <div class="delete-overlay" @click.stop="boardToDelete = board; showDeleteModal = true">
              <img src="/trash.png" alt="Delete board" class="delete-icon" />
            </div>
          </div>
        </template>
      </SliderSection>
    </div>

          <h2 class="mb-0">{{ t('profile.likedBoards', { count: likedBoards.length }) }}</h2>

    <SliderSection :items="likedBoards" :per-page="3" :per-page-md="2" :per-page-sm="1">
      <template #slide="{ item: vote }">
        <BoardCard
          :key="vote.board_id"
          :board="{ board_id: vote.board_id, title: vote.title, description: vote.description, score: vote.vote_score, vote_count: vote.vote_count, size: 0, author_id: '', play_count: 0 }"
          @click="clickBoard(vote.board_id)"
          @vote-changed="(v: number | null) => onLikedBoardVoteChanged(v, vote.board_id)"
        />
      </template>
    </SliderSection>

  <ModalStartGame
    v-if="selecetedBoard"
    v-model="showModal"
    :board="selecetedBoard"
    :cells="selectedCells ?? []"
    :author-name="authorName"
  />

  <ModalDeleteBoard
    v-if="boardToDelete"
    v-model="showDeleteModal"
    :board="boardToDelete"
    @deleted="fetchAllUserBoards"
  />

  </div>
</template>

<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router'
import SliderSection from '@/components/SliderSection.vue'
import BoardCard from '@/components/BoardCard.vue'
import ModalStartGame from '@/components/ModalStartGame.vue'
import ModalDeleteBoard from '@/components/ModalDeleteBoard.vue'
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
const boardsVersion = ref(0)
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
const showDeleteModal = ref(false)
const boardToDelete = ref<Board | null>(null)

function clickBoard(boardID: string) {
    fetchAllCellsForBoard(boardID)
}

function onLikedBoardVoteChanged(vote: number | null, boardId: string) {
    if (!vote) {
        likedBoards.value = likedBoards.value.filter(b => b.board_id !== boardId)
        fetchAllUserBoards().then(() => boardsVersion.value++)
    }
}

function onBoardVoteChange(vote: number | null, board: Board) {
    if (vote === 1) {
        if (!likedBoards.value.find(b => b.board_id === board.board_id)) {
            likedBoards.value = [...likedBoards.value, {
                vote_id: '',
                board_id: board.board_id,
                vote_value: 1,
                title: board.title,
                description: board.description,
                vote_score: board.score,
                vote_count: board.vote_count,
            }]
        }
    } else {
        likedBoards.value = likedBoards.value.filter(b => b.board_id !== board.board_id)
    }
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

.board-card-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
}

.delete-overlay {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  opacity: 0;
  transition: opacity 0.15s;
  cursor: pointer;
}

.board-card-wrapper:hover .delete-overlay {
  opacity: 1;
}

.delete-icon {
  width: 1.2rem;
  height: 1.2rem;
}

[data-theme="dark"] .delete-icon {
  filter: brightness(0) invert(1);
}
</style>
