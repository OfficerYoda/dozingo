<template>
  <section>
    <div class="container">
      <div class="hero-section grid mb-3">
        <div class="hero-banner mb-3 col-8 md-12 sm-12">
          <div class="hero-content">
            <h2 class="hero-title">{{ t('home.hero.title') }}</h2>
            <p class="hero-subtitle">{{ t('home.hero.subtitle') }}</p>
            <RouterLink to="/boards" class="hero-button">{{ t('home.hero.cta') }}</RouterLink>
          </div>
        </div>

        <div class="card activity-card col-4 md-12 sm-12">
          <h3 class="mb-0">{{ t('home.activity.title') }}</h3>
          <div class="raster-container">
            <div class="icon-div box">
              <div class="icon-circle">
                <Medal :size="20" />
              </div>
              <div class="icon-text">
                <small class="category">{{ t('home.activity.bingos') }}</small>
                <span class="category-value">{{ stats.bingos }}</span>
              </div>
            </div>

            <div class="icon-div box">
              <div class="icon-circle">
                <GamepadDirectional :size="20" />
              </div>
              <div class="icon-text">
                <small class="category">{{ t('home.activity.games') }}</small>
                <span class="category-value">{{ stats.games }}</span>
              </div>
            </div>

            <div class="icon-div box">
              <div class="icon-circle">
                <LayoutGrid :size="20" />
              </div>
              <div class="icon-text">
                <small class="category">{{ t('home.activity.boards') }}</small>
                <span class="category-value">{{ stats.boards }}</span>
              </div>
            </div>

            <div class="icon-div box">
              <div class="icon-circle">
                <SquarePlus :size="18" />
              </div>
              <div class="icon-text">
                <small class="category">{{ t('home.activity.cells') }}</small>
                <span class="category-value">{{ stats.cells }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="section-header mb-2">
        <div>
          <h2 class="mb-0">{{ t('home.mostLiked.title') }}</h2>
          <small class="subheading">{{ t('home.mostLiked.subtitle') }}</small>
        </div>
        <RouterLink to="/boards?sort=most-liked">{{ t('home.mostLiked.seeAll') }} &rarr;</RouterLink>
      </div>

      <SliderSection
        :items="mostLikedBoards"
        :per-page="3"
        :per-page-md="2"
        :per-page-sm="1"
        :type="'slide'"
        class="mb-3"
      >
        <template #slide="{ item: board }">
          <button class="card card-border-blue slider-board-card" @click="clickBoard(board)">
            <div class="card-body">
              <h3>{{ board.title }}</h3>
              <small>{{ board.description ?? '—' }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">{{ t('home.card.times', { count: formatCount(board.play_count) }) }}</span>
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ formatCount(board.score) }}</span>
              </div>
            </div>
          </button>
        </template>
      </SliderSection>

      <div class="section-header mb-2">
        <div>
          <h2 class="mb-0">{{ t('home.newest.title') }}</h2>
          <small class="subheading">{{ t('home.newest.subtitle') }}</small>
        </div>
        <RouterLink to="/boards?sort=newest">{{ t('home.newest.seeAll') }} &rarr;</RouterLink>
      </div>

      <SliderSection
        :items="newestBoards"
        :per-page="3"
        :per-page-md="2"
        :per-page-sm="1"
      >
        <template #slide="{ item: board }">
          <button class="card card-border-blue slider-board-card" @click="clickBoard(board)">
            <div class="card-body">
              <h3>{{ board.title }}</h3>
              <small>{{ board.description ?? '—' }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">{{ t('home.card.times', { count: formatCount(board.play_count) }) }}</span>
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ formatCount(board.score) }}</span>
              </div>
            </div>
          </button>
        </template>
      </SliderSection>
    </div>
  </section>

  <ModalStartGame
    v-if="selectedBoard"
    v-model="showModal"
    :board="selectedBoard"
    :cells="selectedCells"
    :author-name="authorName"
  />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Heart, GamepadDirectional, SquarePlus, LayoutGrid, Medal } from 'lucide-vue-next'
import SliderSection from '@/components/SliderSection.vue'
import ModalStartGame from '@/components/ModalStartGame.vue'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()

interface Board {
  board_id: string
  title: string
  description: string
  size: number
  author_id: string
  score: number
  vote_count: number
  play_count: number
}

interface Cell {
  cell_id: string
  content: string
  value: number
}

interface Stats {
  bingos: number
  boards: number
  cells: number
  games: number
}

const { pageTitle } = usePageTitle(t('header.home'))

const mostLikedBoards = ref<Board[]>([])
const newestBoards = ref<Board[]>([])
const selectedBoard = ref<Board | null>(null)
const selectedCells = ref<Cell[]>([])
const authorName = ref<string | null>(null)
const showModal = ref(false)

const stats = ref<Stats>({
  bingos:0,
  boards: 0,
  cells: 0,
  games: 0,
})

async function loadStats(){
  const statsFetched = await fetch('api/stats/recent?duration=24h' , { credentials: 'include' })
  stats.value = await statsFetched.json()
}

function formatCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

async function fetchBoards() {
  const [likedRes, newestRes] = await Promise.all([
    fetch('/api/boards?sort=most-liked&limit=5', { credentials: 'include' }),
    fetch('/api/boards?sort=newest&limit=5', { credentials: 'include' }),
  ])

  if (likedRes.ok) mostLikedBoards.value = await likedRes.json()
  if (newestRes.ok) newestBoards.value = await newestRes.json()
}

async function clickBoard(board: Board) {
  selectedBoard.value = board
  authorName.value = null

  const cellsRes = await fetch('/api/boards/' + board.board_id + '/cells')
  if (cellsRes.ok) {
    const allCells: Cell[] = await cellsRes.json()
    const count = board.size ** 2
    selectedCells.value = [...allCells].sort(() => Math.random() - 0.5).slice(0, count)
  }

  if (board.author_id) {
    const userRes = await fetch('/api/users/' + board.author_id)
    if (userRes.ok) {
      const user = await userRes.json()
      authorName.value = user.username
    }
  }

  showModal.value = true
}

onMounted(fetchBoards)
onMounted(loadStats)
</script>

<style>
/* Hero */
.hero-section {
  display: flex;
  gap: 8px;
}

.hero-banner {
  background-color: var(--color-hero-bg);
  border-radius: 16px;
  padding: 40px 36px;
  height: 100%
}

.hero-content {
  max-width: 480px;
}

.hero-title {
  color: #ffffff;
  font-size: 2rem;
  font-weight: 800;
  line-height: 1.2;
  margin-bottom: 16px;
}

.hero-subtitle {
  color: rgba(255, 255, 255, 0.85);
  font-size: 0.95rem;
  line-height: 1.6;
  margin-bottom: 28px;
}

.hero-button {
  display: inline-block;
  background-color: rgba(255, 255, 255, 0.18);
  color: #ffffff;
  font-weight: 700;
  font-size: 0.95rem;
  padding: 14px 28px;
  border-radius: var(--radius-md);
  text-decoration: none;
  border: 1.5px solid rgba(255, 255, 255, 0.3);
  transition: background-color 0.2s;
}

.hero-button:hover {
  background-color: rgba(255, 255, 255, 0.28);
}

/* Section header */
.section-header {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
}

.heading {
  color: var(--color-heading);
}

.subheading {
  color: var(--color-text-subtle);
}

/* Activity card */
.activity-card {
  display: flex;
  flex-direction: column;
}

/* Card */
.card-body {
  text-align: left;
}

.card-footer {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
}

.card-meta-text {
  color: var(--color-subheading);
  font-weight: 600;
  font-size: 0.8125rem;
}

/* Like group */
.like-group {
  display: flex;
  flex-direction: row;
  gap: 4px;
}

.like-group svg {
  color: var(--color-subheading);
}

.icon-circle {
  width: 52px;
  height: 52px;
  min-width: 52px;
  border-radius: 50%;
  background-color: var(--color-icon-circle-bg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-div {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;
  flex-wrap: wrap;
  justify-content: center;
}

.icon-div svg {
  color: var(--color-icon-svg);
}

.icon-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.category {
  color: var(--color-text-subtle);
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.category-value {
  color: var(--color-heading);
  font-size: 1.5rem;
  font-weight: 800;
  line-height: 1;
}

.raster-container {
  width: 100%;
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-content: space-evenly;
  gap: 0;
}

.box {
  display: flex;
  justify-content: space-evenly;
  align-items: space-evenly;
}

.slider-board-card {
  width: 100%;
  height: 100%;
  text-align: left;
}
</style>
