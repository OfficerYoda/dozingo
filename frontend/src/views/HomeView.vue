<template>
  <section>
    <div class="container">
      <div class="grid">
        <div class="hero-banner mb-3 col-7 md-12 sm-12">
          <div class="hero-content">
            <h2 class="hero-title">{{ t('home.hero.title') }}</h2>
            <p class="hero-subtitle">{{ t('home.hero.subtitle') }}</p>
            <RouterLink to="/boards" class="hero-button">{{ t('home.hero.cta') }}</RouterLink>
          </div>
        </div>

        <div class="card activity-card col-5 md-12 sm-12">
          <h3 class="mb-0">{{ t('home.activity.title') }}</h3>
          <dl class="raster-container">
            <div class="stat-item">
              <div class="icon-circle">
                <Medal :size="20" />
              </div>
              <div class="icon-text">
                <dt class="category">{{ t('home.activity.bingos') }}</dt>
                <dd class="category-value">{{ stats.bingos }}</dd>
              </div>
            </div>

            <div class="stat-item">
              <div class="icon-circle">
                <GamepadDirectional :size="20" />
              </div>
              <div class="icon-text">
                <dt class="category">{{ t('home.activity.games') }}</dt>
                <dd class="category-value">{{ stats.games }}</dd>
              </div>
            </div>

            <div class="stat-item">
              <div class="icon-circle">
                <LayoutGrid :size="20" />
              </div>
              <div class="icon-text">
                <dt class="category">{{ t('home.activity.boards') }}</dt>
                <dd class="category-value">{{ stats.boards }}</dd>
              </div>
            </div>

            <div class="stat-item">
              <div class="icon-circle">
                <SquarePlus :size="18" />
              </div>
              <div class="icon-text">
                <dt class="category">{{ t('home.activity.cells') }}</dt>
                <dd class="category-value">{{ stats.cells }}</dd>
              </div>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </section>

  <section>
    <div class="container">
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
        class="mt-3"
      >
        <template #slide="{ item: board }">
          <BoardCard
            :board="board"
            class="slider-board-card"
            @click="clickBoard(board)"
          />
        </template>
      </SliderSection>
    </div>
  </section>

  <section>
    <div class="container">
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
          <BoardCard
            :board="board"
            class="slider-board-card"
            @click="clickBoard(board)"
          />
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
import { GamepadDirectional, SquarePlus, LayoutGrid, Medal } from 'lucide-vue-next'
import SliderSection from '@/components/SliderSection.vue'
import BoardCard from '@/components/BoardCard.vue'
import ModalStartGame from '@/components/ModalStartGame.vue'
import { usePageTitle } from '@/composables/usePageTitle'
import * as boardService from '@/services/board.service'
import * as userService from '@/services/user.service'
import * as statsService from '@/services/stats.service'
import type { Board, Cell, Stats } from '@/services/api.type'

const { t } = useI18n()

const { pageTitle } = usePageTitle(t('header.home'))

const mostLikedBoards = ref<Board[]>([])
const newestBoards = ref<Board[]>([])
const selectedBoard = ref<Board | null>(null)
const selectedCells = ref<Cell[]>([])
const authorName = ref<string | null>(null)
const showModal = ref(false)

const stats = ref<Stats>({
  bingos: 0,
  boards: 0,
  cells: 0,
  games: 0,
})

async function loadStats() {
  stats.value = await statsService.getRecentStats()
}


async function fetchBoards() {
  const [liked, newest] = await Promise.all([
    boardService.getBoards({ sort: 'most-liked', limit: 5 }),
    boardService.getBoards({ sort: 'newest', limit: 5 }),
  ])
  mostLikedBoards.value = liked
  newestBoards.value = newest
}

async function clickBoard(board: Board) {
  selectedBoard.value = board
  authorName.value = null

  const allCells = await boardService.getCellsForBoard(board.board_id)
  const count = board.size ** 2
  selectedCells.value = [...allCells].sort(() => Math.random() - 0.5).slice(0, count)

  if (board.author_id) {
    try {
      const user = await userService.getUserById(board.author_id)
      authorName.value = user.username
    } catch { /* author is optional */ }
  }

  showModal.value = true
}

onMounted(fetchBoards)
onMounted(loadStats)
</script>

<style scoped>
/* Hero */

.hero-banner {
  background-color: var(--color-hero-bg);
  border-radius: 16px;
  padding: 40px 36px;
  height: 100%
}

.hero-content {
  max-width: 480px;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-around;
  align-items: start;
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
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.subheading {
  color: var(--color-text-subtle);
}

/* Activity card */
.activity-card {
  display: flex;
  flex-direction: column;
}

.slider-board-card {
  width: 100%;
  height: 100%;
  text-align: left;
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

.stat-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 14px;
  padding-block: 12px;
  justify-content: center;
}

.stat-item svg {
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
  text-transform: uppercase;
}

.category-value {
  color: var(--color-heading);
  font-size: 1.5rem;
  font-weight: 800;
  line-height: 1;
}

.raster-container {
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-content: space-evenly;
  overflow-x: auto;
  gap: 8px;
}
</style>
