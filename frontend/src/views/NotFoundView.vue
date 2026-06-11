<template>
  <section>
    <div class="container">
      <div class="not-found-wrapper">
        <div class="bingo-grid" aria-hidden="true">
          <div
            v-for="(cell, i) in bingoGrid"
            :key="i"
            class="bingo-cell"
            :class="{ checked: cell.checked }"
          >
            {{ cell.text }}
          </div>
        </div>

        <div class="not-found-code">404</div>
        <h1 class="not-found-title">{{ t('notFound.title') }}</h1>
        <p class="not-found-subtitle">{{ t('notFound.subtitle') }}</p>
        <small class="not-found-hint">{{ t('notFound.hint') }}</small>
        <RouterLink to="/" class="hero-button">{{ t('notFound.cta') }}</RouterLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()
usePageTitle(t('notFound.title'))

const bingoGrid = [
  { text: '404', checked: true },
  { text: t('notFound.grid.wrongTurn'), checked: false },
  { text: t('notFound.grid.lost'), checked: true },
  { text: t('notFound.grid.blame'), checked: false },
  { text: 'BINGO?', checked: false },
  { text: t('notFound.grid.gone'), checked: true },
  { text: t('notFound.grid.oops'), checked: false },
  { text: t('notFound.grid.notHere'), checked: true },
  { text: t('notFound.grid.classic'), checked: false },
]
</script>

<style scoped>
.not-found-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 60px 24px;
  gap: 20px;
}

.bingo-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
  width: 260px;
  animation: tilt 3s ease-in-out infinite;
}

@keyframes tilt {
  0%, 100% { transform: rotate(-2deg); }
  50%       { transform: rotate(2deg); }
}

.bingo-cell {
  background: var(--color-bg-card-tinted);
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 10px 4px;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--color-subheading);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: background var(--transition-normal), color var(--transition-normal);
}

.bingo-cell.checked {
  background: var(--color-hero-bg);
  color: #fff;
  border-color: var(--color-hero-bg);
  position: relative;
}

.bingo-cell.checked::after {
  content: '✕';
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.4rem;
  font-weight: 900;
  color: rgba(255,255,255,0.35);
}

.not-found-code {
  font-size: 6rem;
  font-weight: 900;
  line-height: 1;
  color: var(--color-hero-bg);
  letter-spacing: -4px;
}

.not-found-title {
  font-size: 1.6rem;
  font-weight: 800;
  color: var(--color-heading);
  margin: 0;
}

.not-found-subtitle {
  color: var(--color-text-subtle);
  font-size: 0.95rem;
  max-width: 420px;
  margin: 0;
}

.not-found-hint {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  font-style: italic;
}

.hero-button {
  background-color: var(--color-hero-bg);
  color: #ffffff;
  border-color: var(--color-hero-bg);
}

.hero-button:hover {
  background-color: var(--color-primary-600);
  border-color: var(--color-primary-600);
}
</style>
