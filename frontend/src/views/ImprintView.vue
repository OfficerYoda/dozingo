<template>
  <section>
    <div class="container">

      <!-- BINGO Overlay -->
      <Teleport to="body">
        <div v-if="isBingo" class="bingo-overlay" @click="resetBingo">
          <div
            v-for="p in particles"
            :key="p.id"
            class="confetti-particle"
            :style="p.style"
          />
          <div class="bingo-burst">
            <div class="bingo-text">BINGO!</div>
            <div class="bingo-sub">Du hast das Impressum gelesen 🎉</div>
            <div class="bingo-hint">Klick zum Schließen</div>
          </div>
        </div>
      </Teleport>

      <div class="imprint-hero">
        <div class="hero-badge">§ DDG</div>
        <h1>{{ t('imprint.title') }}</h1>
        <p class="hero-sub">{{ t('imprint.heroSub') }}<br>
          <span class="hero-sub-muted">{{ t('imprint.heroSubMuted') }}</span>
        </p>
      </div>

      <div class="imprint-board">
        <div
          class="board-cell cell-wide"
          :class="{ 'cell-checked': isChecked('details') }"
          role="button"
          tabindex="0"
          @click="toggleCell('details')"
          @keydown.enter.space.prevent="toggleCell('details')"
        >
          <span class="cell-label">{{ t('imprint.cells.details') }}</span>
          <div class="cell-content">
            <strong>Dozingo</strong><br>
            Erzbergerstraße 121<br>
            76133 Karlsruhe<br>
            Deutschland
          </div>
          <span class="cell-check" :class="{ 'cell-check-active': isChecked('details') }">✓</span>
        </div>

        <div
          class="board-cell"
          :class="{ 'cell-checked': isChecked('contact') }"
          role="button"
          tabindex="0"
          @click="toggleCell('contact')"
          @keydown.enter.space.prevent="toggleCell('contact')"
        >
          <span class="cell-label">{{ t('imprint.cells.contact') }}</span>
          <div class="cell-content">
            <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a>
          </div>
          <span class="cell-check" :class="{ 'cell-check-active': isChecked('contact') }">✓</span>
        </div>

        <div class="board-cell board-cell-free">
          <span class="cell-label">FREE</span>
          <div class="cell-content cell-free-text">{{ t('imprint.cells.free') }}</div>
        </div>

        <div
          class="board-cell"
          :class="{ 'cell-checked': isChecked('responsible') }"
          role="button"
          tabindex="0"
          @click="toggleCell('responsible')"
          @keydown.enter.space.prevent="toggleCell('responsible')"
        >
          <span class="cell-label">{{ t('imprint.cells.responsible') }}</span>
          <div class="cell-content">
            Dozingo-Team<br>
            DHBW Karlsruhe
          </div>
          <span class="cell-check" :class="{ 'cell-check-active': isChecked('responsible') }">✓</span>
        </div>

        <div
          class="board-cell cell-wide"
          :class="{ 'cell-checked': isChecked('liability') }"
          role="button"
          tabindex="0"
          @click="toggleCell('liability')"
          @keydown.enter.space.prevent="toggleCell('liability')"
        >
          <span class="cell-label">{{ t('imprint.cells.liability') }}</span>
          <div class="cell-content">{{ t('imprint.cells.liabilityContent') }}</div>
          <span class="cell-check" :class="{ 'cell-check-active': isChecked('liability') }">✓</span>
        </div>

        <div
          class="board-cell"
          :class="{ 'cell-checked': isChecked('copyright') }"
          role="button"
          tabindex="0"
          @click="toggleCell('copyright')"
          @keydown.enter.space.prevent="toggleCell('copyright')"
        >
          <span class="cell-label">{{ t('imprint.cells.copyright') }}</span>
          <div class="cell-content">{{ t('imprint.cells.copyrightContent') }}</div>
          <span class="cell-check" :class="{ 'cell-check-active': isChecked('copyright') }">✓</span>
        </div>
      </div>

      <p class="imprint-outro">
        {{ t('imprint.outro') }} <strong>BINGO!</strong> 🎯
        <span v-if="checkedCount > 0" class="imprint-progress">({{ checkedCount }}/5)</span>
      </p>

    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()
usePageTitle(t('imprint.title'))

const CELL_IDS = ['details', 'contact', 'responsible', 'liability', 'copyright'] as const
type CellId = typeof CELL_IDS[number]

const clickedCells = ref(new Set<CellId>())

function toggleCell(id: CellId) {
  const next = new Set(clickedCells.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  clickedCells.value = next
}

function isChecked(id: CellId) {
  return clickedCells.value.has(id)
}

const checkedCount = computed(() => clickedCells.value.size)
const isBingo = computed(() => CELL_IDS.every(id => clickedCells.value.has(id)))

const COLORS = ['#4052B6', '#C0185A', '#5A5781', '#FFD700', '#E3DFFF', '#FF6B9D', '#7B6FFF']

interface Particle {
  id: number
  style: Record<string, string>
}

const particles = ref<Particle[]>([])

watch(isBingo, (val) => {
  if (val) {
    particles.value = Array.from({ length: 90 }, (_, i) => ({
      id: i,
      style: {
        left: `${(i / 90) * 100 + (Math.sin(i * 2.3) * 8)}%`,
        top: `-${10 + (i % 3) * 8}px`,
        background: COLORS[i % COLORS.length] ?? '#4052B6',
        width: `${6 + (i % 3) * 4}px`,
        height: `${6 + (i % 5) * 3}px`,
        'border-radius': i % 3 === 0 ? '50%' : '2px',
        'animation-delay': `${(i % 10) * 0.08}s`,
        'animation-duration': `${1.4 + (i % 7) * 0.2}s`,
      },
    }))
  }
})

function resetBingo() {
  clickedCells.value = new Set()
  particles.value = []
}
</script>

<style scoped>
.imprint-hero {
  padding: 40px 0 32px;
  max-width: 600px;
}

.hero-badge {
  display: inline-block;
  background-color: #4052B6;
  color: #fff;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  padding: 3px 10px;
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
}

.imprint-hero h1 {
  font-size: 2.4rem;
  font-weight: 900;
  color: var(--color-heading);
  margin: 0 0 10px;
}

.hero-sub {
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.hero-sub-muted {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

.imprint-board {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  max-width: 760px;
  margin-bottom: 28px;
}

.board-cell {
  background-color: var(--color-bg-surface);
  border: 3px solid var(--color-input-bg);
  border-radius: var(--radius-lg);
  padding: 16px;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: border-color 0.2s, background-color 0.2s, transform 0.1s;
  cursor: pointer;
  user-select: none;
}

.board-cell:hover {
  border-color: var(--color-subheading);
}

.board-cell:active {
  transform: scale(0.97);
}

.board-cell-free {
  cursor: default;
}

.cell-checked {
  border-color: var(--color-subheading) !important;
  background-color: var(--color-bg-card-tinted);
}

.cell-wide {
  grid-column: span 2;
}

.board-cell-free {
  background-color: var(--color-subheading);
  border-color: var(--color-subheading);
}

.cell-label {
  font-size: 0.6rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  color: var(--color-text-muted);
  text-transform: uppercase;
}

.board-cell-free .cell-label {
  color: #E3DFFF;
}

.cell-content {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.cell-content a {
  color: var(--card-blue);
}

.cell-free-text {
  color: #fff !important;
  font-weight: 600;
  font-size: 0.9rem;
}

.cell-check {
  position: absolute;
  top: 10px;
  right: 12px;
  width: 22px;
  height: 22px;
  background-color: var(--color-input-bg);
  color: var(--color-subheading);
  border-radius: 50%;
  font-size: 0.7rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s, color 0.2s, transform 0.15s;
}

.cell-check-active {
  background-color: var(--color-subheading);
  color: #fff;
  transform: scale(1.15);
}

.imprint-outro {
  color: var(--color-text-muted);
  font-size: 0.9rem;
  padding-bottom: 40px;
}

.imprint-outro strong {
  color: var(--card-red);
}

.imprint-progress {
  font-size: 0.8rem;
  color: var(--color-subheading);
  font-weight: 600;
  margin-left: 4px;
}

/* BINGO overlay */
.bingo-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(10, 8, 30, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
}

.bingo-burst {
  text-align: center;
  animation: bingo-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
  z-index: 1;
}

.bingo-text {
  font-size: clamp(4rem, 15vw, 9rem);
  font-weight: 900;
  color: #fff;
  letter-spacing: -0.02em;
  line-height: 1;
  background: linear-gradient(135deg, #FFD700 0%, #FF6B9D 50%, #7B6FFF 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 0 40px rgba(255, 215, 0, 0.4));
}

.bingo-sub {
  font-size: 1.3rem;
  color: #E3DFFF;
  margin-top: 12px;
  font-weight: 600;
}

.bingo-hint {
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.4);
  margin-top: 24px;
  letter-spacing: 0.05em;
}

.confetti-particle {
  position: fixed;
  animation: confetti-drop linear forwards;
  pointer-events: none;
}

@keyframes bingo-pop {
  0% { transform: scale(0.3) rotate(-8deg); opacity: 0; }
  60% { transform: scale(1.08) rotate(3deg); opacity: 1; }
  80% { transform: scale(0.96) rotate(-1deg); }
  100% { transform: scale(1) rotate(0deg); opacity: 1; }
}

@keyframes confetti-drop {
  0% {
    transform: translateY(0) rotate(0deg);
    opacity: 1;
  }
  85% { opacity: 1; }
  100% {
    transform: translateY(105vh) rotate(680deg);
    opacity: 0;
  }
}

@media (max-width: 600px) {
  .imprint-board {
    grid-template-columns: 1fr;
  }
  .cell-wide {
    grid-column: span 1;
  }
}
</style>
