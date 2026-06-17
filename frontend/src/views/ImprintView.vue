<template>
  <section>
    <div class="container">

      <div class="imprint-hero">
        <div class="hero-badge">§ TMG</div>
        <h1>Impressum</h1>
        <p class="hero-sub">Das langweiligste Board, das wir je gebaut haben.<br>
          <span class="hero-sub-muted">Aber hey — auch Pflichtfelder verdienen Liebe.</span>
        </p>
      </div>

      <div class="imprint-board">
        <div class="board-cell cell-wide">
          <span class="cell-label">ANGABEN</span>
          <div class="cell-content">
            <strong>Dozingo</strong><br>
            Erzbergerstraße 121<br>
            76133 Karlsruhe<br>
            Deutschland
          </div>
          <span class="cell-check">✓</span>
        </div>

        <div class="board-cell">
          <span class="cell-label">KONTAKT</span>
          <div class="cell-content">
            <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a>
          </div>
          <span class="cell-check">✓</span>
        </div>

        <div class="board-cell board-cell-free">
          <span class="cell-label">FREE</span>
          <div class="cell-content cell-free-text">Keine Beschwerden bisher 🎉</div>
        </div>

        <div class="board-cell">
          <span class="cell-label">VERANTWORTLICH</span>
          <div class="cell-content">
            Dozingo-Team<br>
            DHBW Karlsruhe
          </div>
          <span class="cell-check">✓</span>
        </div>

        <div class="board-cell cell-wide">
          <span class="cell-label">HAFTUNG</span>
          <div class="cell-content">
            Inhalte wurden sorgfältig erstellt. Für Richtigkeit, Vollständigkeit und
            Aktualität übernehmen wir keine Gewähr. (Klassiker.)
          </div>
          <span class="cell-check">✓</span>
        </div>

        <div class="board-cell">
          <span class="cell-label">URHEBERRECHT</span>
          <div class="cell-content">
            Alle Inhalte © Dozingo-Team.<br>
            Bitte nicht klauen. 🙏
          </div>
          <span class="cell-check">✓</span>
        </div>
      </div>

      <p class="imprint-outro">
        Impressum vollständig ausgefüllt. <strong>BINGO!</strong> 🎯
      </p>

    </div>
  </section>
</template>

<script setup lang="ts">
import { usePageTitle } from '@/composables/usePageTitle'
usePageTitle('Impressum')
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
  background-color: #fff;
  border: 3px solid #E3DFFF;
  border-radius: var(--radius-lg);
  padding: 16px;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: border-color 0.2s;
}

.board-cell:hover {
  border-color: #5A5781;
}

.cell-wide {
  grid-column: span 2;
}

.board-cell-free {
  background-color: #5A5781;
  border-color: #5A5781;
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
  color: #4052B6;
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
  background-color: #E3DFFF;
  color: #5A5781;
  border-radius: 50%;
  font-size: 0.7rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.imprint-outro {
  color: var(--color-text-muted);
  font-size: 0.9rem;
  padding-bottom: 40px;
}

.imprint-outro strong {
  color: #C0185A;
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
