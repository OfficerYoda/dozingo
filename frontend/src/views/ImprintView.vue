<template>
  <section>
    <div class="container">
      <div class="legal-wrapper">
        <div class="legal-section">
          <h2>Angaben gemäß § 5 TMG</h2>
          <p>
            Dozingo<br>
            Erzbergerstraße 121<br>
            76133 Karlsruhe<br>
            Deutschland
          </p>
        </div>

        <div class="legal-section">
          <h2>Kontakt</h2>
          <p>
            E-Mail: <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a>
          </p>
        </div>

        <div class="legal-section">
          <h2>Verantwortlich für den Inhalt</h2>
          <p>
            Dozingo-Team<br>
            Duale Hochschule Baden-Württemberg<br>
            Karlsruhe
          </p>
        </div>

        <div class="legal-section">
          <h2>Haftungsausschluss</h2>
          <p>
            Die Inhalte dieser Seite wurden mit größter Sorgfalt erstellt.
            Für die Richtigkeit, Vollständigkeit und Aktualität der Inhalte
            können wir jedoch keine Gewähr übernehmen.
          </p>
        </div>

        <div class="legal-section">
          <h2>Urheberrecht</h2>
          <p>
            Die durch die Seitenbetreiber erstellten Inhalte und Werke auf
            diesen Seiten unterliegen dem deutschen Urheberrecht. Die
            Vervielfältigung, Bearbeitung, Verbreitung und jede Art der
            Verwertung außerhalb der Grenzen des Urheberrechtes bedürfen der
            schriftlichen Zustimmung des jeweiligen Autors bzw. Erstellers.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { usePageTitle } from '@/composables/usePageTitle'
usePageTitle('Impressum')
</script>

<style scoped>
.legal-wrapper {
  max-width: 720px;
  padding: 20px 0 40px;
}

.legal-section {
  margin-bottom: 28px;
}

.legal-section h2 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-subheading);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
}

.legal-section p {
  color: var(--color-text-secondary);
  line-height: 1.7;
  margin: 0;
}

.legal-section a {
  color: var(--color-primary-600);
}
</style>
