<template>
  <section>
    <div class="container">
      <div class="hero-section grid">
        <div class="hero-banner mb-3 col-8 md-12 sm-12">
          <div class="hero-content">
            <h2 class="hero-title">Ready for your next lecture win?</h2>
            <p class="hero-subtitle">Quick-start a bingo grid for today's lecture and join other students currently playing.</p>
            <RouterLink to="/cards" class="hero-button">Browse Cards</RouterLink>
          </div>
        </div>

        <div class="card col-4 md-12 sm-12">
          <h2>Content coming soon</h2>
        </div>
      </div>

      <div class="mb-3">
        <div class="section-header mb-3">
          <div>
            <h2 class="mb-0 heading">Most Liked Cards</h2>
            <small class="subheading">The community's current favorites</small>
          </div>
          <RouterLink to="/cards?sort=most-liked">See all &rarr;</RouterLink>
        </div>

        <div class="grid">
          <button v-for="board in mostLikedBoards" :key="board.board_id" class="card card-border-blue col-4 md-6 sm-12">
            <div class="card-body">
              <h3>{{ board.title }}</h3>
              <small>{{ board.description ?? '—' }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">{{ formatCount(board.play_count) }} times</span>
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ formatCount(board.score) }}</span>
              </div>
            </div>
          </button>
        </div>
      </div>

      <div>
        <div class="section-header mb-3">
          <div>
            <h2 class="mb-0 heading">Recently Added Cards</h2>
            <small class="subheading">Fresh bingo cards from the community</small>
          </div>
          <RouterLink to="/cards?sort=newest">See all &rarr;</RouterLink>
        </div>

        <div class="grid">
          <button v-for="board in newestBoards" :key="board.board_id" class="card card-border-blue col-4 md-6 sm-12">
            <div class="card-body">
              <h3>{{ board.title }}</h3>
              <small>{{ board.description ?? '—' }}</small>
            </div>
            <hr class="mb-2">
            <div class="card-footer">
              <span class="card-meta-text">{{ formatCount(board.play_count) }} times</span>
              <div class="like-group">
                <Heart :size="20" />
                <span class="card-meta-text">{{ formatCount(board.score) }}</span>
              </div>
            </div>
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Heart } from 'lucide-vue-next'

interface Board {
  board_id: string
  title: string
  description: string | null
  size: number
  author_id: string
  score: number
  vote_count: number
  play_count: number
}

const mostLikedBoards = ref<Board[]>([])
const newestBoards = ref<Board[]>([])

function formatCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

async function fetchBoards() {
  const [likedRes, newestRes] = await Promise.all([
    fetch('/api/boards?sort=most-liked&limit=3', { credentials: 'include' }),
    fetch('/api/boards?sort=newest&limit=3', { credentials: 'include' }),
  ])

  if (likedRes.ok) mostLikedBoards.value = await likedRes.json()
  if (newestRes.ok) newestBoards.value = await newestRes.json()
}

onMounted(fetchBoards)
</script>

<style>
/* Hero */
.hero-section {
  display: flex;
  gap: 8px;
}

.hero-banner {
  background-color: #4B4AC8;
  border-radius: 16px;
  padding: 40px 36px;
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
  color: #2C2A51;
}

.subheading {
  color: #75729E;
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
  color: #5A5781;
  font-weight: 600;
  font-size: 13px;
}

/* Like group */
.like-group {
  display: flex;
  flex-direction: row;
  gap: 4px;
}

.like-group svg {
  color: #5A5781;
}
</style>
