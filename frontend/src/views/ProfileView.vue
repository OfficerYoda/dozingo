<template>
  <div className="container">
    <h1>Profile</h1>
    <div v-if="!auth.state.ready">Loading...</div>
    <div v-else-if="!auth.state.user">Not logged in.</div>
    <div v-else>
      <h2>Welcome, {{ auth.state.user.username }}!</h2>
      <p v-if="auth.state.user.email">{{ auth.state.user.email }}</p>
    </div>
    <div class="container" style="padding-right: 0%; padding-left: 0%;">
      <div class="list-header mb-4">
          <h2 class="mb-0">Explore your {{boards.length}} Cards</h2>
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

      <div class="carousel-wrapper">
        <button class="carousel-btn" @click="scrollCarousel(carouselRef, -1)">&#8592;</button>
        <div class="carousel" ref="carouselRef">
          <button v-for="board in boards" :key="board.board_id" @click="clickBoard(board.board_id)"
              class="card card-border-blue carousel-card">
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
        </div>
        <button class="carousel-btn" @click="scrollCarousel(carouselRef, 1)">&#8594;</button>
      </div>
    </div>
    <div class="container" style="padding-right: 0%; padding-left: 0%; padding-top: 2%;">
      <div class="list-header mb-4">
          <h2 class="mb-0">Explore your {{ likedBoards.length }} liked Cards</h2>
      </div>

      <div class="carousel-wrapper">
        <button class="carousel-btn" @click="scrollCarousel(likedCarouselRef, -1)">&#8592;</button>
        <div class="carousel" ref="likedCarouselRef">
          <button v-for="vote in likedBoards" :key="vote.board_id" @click="clickBoard(vote.board_id)"
              class="card card-border-blue carousel-card">
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
        </div>
        <button class="carousel-btn" @click="scrollCarousel(likedCarouselRef, 1)">&#8594;</button>
      </div>
    </div>
      <Teleport to="body">
        <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
            <div class="card">
                <div class="header-modal">
                    <div>
                        <h2 class="mb-0 header-modal-title">{{ selecetedBoard?.title }}</h2>
                        <small class="header-modal-subtitle">{{ selecetedBoard?.description }}</small>
                        <div class="modal-stats">
                            <span class="stat-item stat-plays">
                                <Play :size="15"/> {{ selecetedBoard?.play_count }}
                            </span>
                            <span class="stat-item stat-likes">
                                <Heart :size="15"/> {{ selecetedBoard?.score }}
                            </span>
                            <span class="stat-item stat-size">
                                <LayoutGrid :size="15"/> {{ selecetedBoard?.size }}x{{ selecetedBoard?.size }}
                            </span>
                        </div>
                    </div>
                    <X :size="20" @click="showModal = false" />
                </div>

                <hr class="mb-3">


                <ul class="background-seperate-cells">
                    <li v-for="cell in selectedCells" :key="cell.cell_id" class="card cell-btn">
                        <p>{{ cell.content }}</p>
                    </li>
                </ul>

                <hr>

                <div class="bottom-bar">
                    <div class="bottom-bar-text">
                        <small>Createt by</small>
                        <span>Hier Author eintragen</span>
                    </div>
                    <div class="right-buttons-bottom">
                        <button class="btn btn-secondary button-bottom-row" @click="shuffle">
                            <Dices :size="20" />
                            <p class="mb-0">Shuffle</p>
                        </button>
                        <button class="btn btn-primary button-bottom-row">
                            <Play :size="20" />
                            <p class="mb-0">Start the game</p>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { useAuth } from '@/composables/useAuth'
import { ref, watch, useTemplateRef } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Heart, X, Dices, LayoutGrid, Play } from 'lucide-vue-next'

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
}

interface Cell {
    cell_id: string,
    content: string,
    value: 0,
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

useI18n()
const route = useRoute()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const likedBoards = ref<Vote[]>([])
const cells = ref<Cell[]>([])
const selecetedBoard = ref<Board>()
const selectedCells = ref<Cell[]>()

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
    showModal.value = true
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : '')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllUserBoards()
}, { immediate: true })

fetchLikedBoards()

const showModal = ref(false)

const carouselRef = useTemplateRef<HTMLElement>('carouselRef')
const likedCarouselRef = useTemplateRef<HTMLElement>('likedCarouselRef')

function scrollCarousel(el: HTMLElement | null, direction: 1 | -1) {
    el?.scrollBy({ left: direction * (el.offsetWidth * 0.8), behavior: 'smooth' })
}

function shuffle() {
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)
}

function clickBoard(boardID: string) {
    console.log("Statet loading the cells for board with boardid " + boardID)
    fetchAllCellsForBoard(boardID)
}
</script>

<style scoped>
.carousel-wrapper {
  display: flex;
  align-items: center;
  position: relative;
}

.carousel {
  display: flex;
  gap: 1rem;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  flex: 1;
}

.carousel::-webkit-scrollbar {
  display: none;
}

.carousel-card {
  min-width: 25%;
  scroll-snap-align: start;
  flex-shrink: 0;
}

.carousel-btn {
  position: absolute;
  z-index: 1;
  background: none;
  border: none;
  border-radius: 50%;
  width: 2.25rem;
  height: 2.25rem;
  cursor: pointer;
  flex-shrink: 0;
  align-self: center;
}

.carousel-btn:hover {
  background: var(--color-background, white);
}

.carousel-btn:first-child {
  left: 0.25rem;
}

.carousel-btn:last-child {
  right: 0.25rem;
}
</style>
